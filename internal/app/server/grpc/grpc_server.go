package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/aseptimu/url-shortener/internal/app/config"
	"github.com/aseptimu/url-shortener/internal/app/handlers/grpc/proto"
	"github.com/aseptimu/url-shortener/internal/app/handlers/http/dbhandlers"
	"github.com/aseptimu/url-shortener/internal/app/handlers/http/shortenurlhandlers"
	"github.com/aseptimu/url-shortener/internal/app/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
)

type Server struct {
	proto.UnimplementedURLShortenerServer
	cfg    *config.ConfigType
	svc    service.URLShortener
	getSvc shortenurlhandlers.URLGetter
	delSvc service.URLDeleter
	ping   dbhandlers.Pinger
}

func NewServer(
	cfg *config.ConfigType,
	svc service.URLShortener,
	getSvc shortenurlhandlers.URLGetter,
	delSvc service.URLDeleter,
	ping dbhandlers.Pinger,
) *Server {
	return &Server{cfg: cfg, svc: svc, getSvc: getSvc, delSvc: delSvc, ping: ping}
}

func (s *Server) GetURL(ctx context.Context, req *proto.GetURLRequest) (*proto.GetURLResponse, error) {
	original, err := s.getSvc.GetOriginalURL(ctx, req.Url)
	switch {
	case errors.Is(err, service.ErrURLNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrURLDeleted):
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &proto.GetURLResponse{OriginalUrl: original}, nil
}

func (s *Server) GetUserURLs(ctx context.Context, _ *proto.GetUserURLsRequest) (*proto.GetUserURLsResponse, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no metadata in context")
	}
	users := md.Get("userID")
	if len(users) == 0 {
		return nil, status.Error(codes.Unauthenticated, "no userID provided")
	}

	userID := users[0]

	urls, err := s.getSvc.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("No urls for user: %s", userID))
	}

	var responseUrls []*proto.UserURL
	for _, url := range urls {
		responseUrls = append(responseUrls, &proto.UserURL{
			ShortUrl:    s.cfg.BaseAddress + "/" + url.ShortURL,
			OriginalUrl: url.OriginalURL,
		})
	}

	return &proto.GetUserURLsResponse{Urls: responseUrls}, nil
}

func (s *Server) GetStats(ctx context.Context, _ *proto.GetStatsRequest) (*proto.GetStatsResponse, error) {
	if s.cfg.TrustedSubnet == "" {
		return nil, status.Error(codes.FailedPrecondition, "No trusted subnet provided")
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("X-Real-IP"); len(vals) != 0 {
			clientIP := net.ParseIP(vals[0])
			_, trustedNet, err := net.ParseCIDR(s.cfg.TrustedSubnet)
			if err != nil {
				return nil, status.Error(codes.FailedPrecondition, "Invalid CIDR in config.TrustedSubnet")
			}

			if clientIP == nil || !trustedNet.Contains(clientIP) {
				return nil, status.Error(codes.FailedPrecondition, "IP not in trusted subnet")
			}
			stats, err := s.getSvc.GetStats(ctx)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			return &proto.GetStatsResponse{
				TotalUrls:  int32(stats.Urls),
				TotalUsers: int32(stats.Users),
			}, nil
		}
	}
	return nil, status.Error(codes.FailedPrecondition, "No metadata in context")
}

// URLCreator обрабатывает создание URL
// Возвращает новый короткий URL в виде text/plain.
// В случае конфликта возвращает 6 AlreadyExists с уже существующим ключом.
func (s *Server) URLCreator(ctx context.Context, req *proto.URLCreatorRequest) (*proto.URLCreatorResponse, error) {
	var userIDStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if val := md.Get("userID"); len(val) != 0 {
			userIDStr = val[0]
		} else {
			return nil, status.Error(codes.PermissionDenied, "Unauthorized")
		}

		shortURL, err := s.svc.ShortenURL(ctx, req.OriginalUrl, userIDStr)
		if err != nil && !errors.Is(err, service.ErrConflict) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if errors.Is(err, service.ErrConflict) {
			st := status.New(codes.AlreadyExists, "Conflict")
			stWithDetails, err2 := st.WithDetails(&proto.URLCreatorResponse{
				ShortenUrl: shortURL,
			})
			if err2 != nil {
				return nil, status.Error(codes.Internal, "cannot attach details")
			}
			return nil, stWithDetails.Err()
		} else {
			return &proto.URLCreatorResponse{
				ShortenUrl: shortURL,
			}, nil
		}
	}

	return nil, status.Error(codes.FailedPrecondition, "No metadata in context")
}
func (s *Server) Ping(ctx context.Context, _ *proto.Empty) (*proto.PingResponse, error) {
	if s.ping == nil {
		return nil, status.Error(codes.FailedPrecondition, "Database is not initialized ")
	}
	err := s.ping.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &proto.PingResponse{
		Status: "OK",
	}, nil
}

func (s *Server) URLCreatorJSON(ctx context.Context, req *proto.URLCreatorJSONRequest) (*proto.URLCreatorJSONResponse, error) {
	var userIDStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if val := md.Get("userID"); len(val) != 0 {
			userIDStr = val[0]
		} else {
			return nil, status.Error(codes.PermissionDenied, "Unauthorized")
		}

		shortURL, err := s.svc.ShortenURL(ctx, req.JsonOriginalUrl, userIDStr)
		if err != nil && !errors.Is(err, service.ErrConflict) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if errors.Is(err, service.ErrConflict) {
			st := status.New(codes.AlreadyExists, "Conflict")
			stWithDetails, err2 := st.WithDetails(&proto.URLCreatorResponse{
				ShortenUrl: shortURL,
			})
			if err2 != nil {
				return nil, status.Error(codes.Internal, "cannot attach details")
			}
			return nil, stWithDetails.Err()
		} else {
			return &proto.URLCreatorJSONResponse{
				JsonResult: shortURL,
			}, nil
		}
	}

	return nil, status.Error(codes.FailedPrecondition, "No metadata in context")
}
func (s *Server) URLCreatorBatch(ctx context.Context, req *proto.URLCreatorBatchRequest) (*proto.URLCreatorBatchResponse, error) {
	var userIDStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if val := md.Get("userID"); len(val) != 0 {
			userIDStr = val[0]
		} else {
			return nil, status.Error(codes.PermissionDenied, "Unauthorized")
		}
		origs := make([]string, len(req.Requests))
		for i, item := range req.Requests {
			origs[i] = item.OriginalUrl
		}
		inputURLs := make([]string, len(req.Requests))
		for i, request := range req.Requests {
			inputURLs[i] = request.OriginalUrl
		}

		shortenedURLs, err := s.svc.ShortenURLs(ctx, inputURLs, userIDStr)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		shortenedURLsMap := make(map[string]string, len(shortenedURLs))
		for short, orig := range shortenedURLs {
			shortenedURLsMap[orig] = short
		}

		responseURLs := make([]*proto.URLResponse, len(req.Requests))
		for i, request := range req.Requests {
			shortURL, ok := shortenedURLsMap[request.OriginalUrl]
			if !ok {
				return nil, status.Error(codes.Internal, "Mismatch in shortened URLs")
			}
			responseURLs[i] = &proto.URLResponse{
				CorrelationId: request.CorrelationId,
				ShortUrl:      s.cfg.BaseAddress + "/" + shortURL,
			}
		}
		return &proto.URLCreatorBatchResponse{
			Responses: responseURLs,
		}, nil
	}

	return nil, status.Error(codes.FailedPrecondition, "No metadata in context")
}
func (s *Server) DeleteUserURLs(ctx context.Context, req *proto.DeleteUserURLsRequest) (*proto.DeleteUserURLsResponse, error) {
	var userIDStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if val := md.Get("userID"); len(val) != 0 {
			userIDStr = val[0]
		} else {
			return nil, status.Error(codes.Unauthenticated, "Unauthorized")
		}
		task := shortenurlhandlers.DeleteTask{
			URLs:   req.Urls,
			UserID: userIDStr,
		}

		shortenurlhandlers.DeleteTaskCh <- task
	}

	return nil, status.Error(codes.FailedPrecondition, "No metadata in context")

}
