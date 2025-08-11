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
	original, err := s.getSvc.GetOriginalURL(ctx, req.GetUrl())
	switch {
	case errors.Is(err, service.ErrURLNotFound):
		return nil, status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrURLDeleted):
		return nil, status.Error(codes.NotFound, err.Error())
	}

	resp := &proto.GetURLResponse{}
	resp.SetOriginalUrl(original)

	return resp, nil
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
		userURL := &proto.UserURL{}
		userURL.SetShortUrl(s.cfg.BaseAddress + "/" + url.ShortURL)
		userURL.SetOriginalUrl(url.OriginalURL)
		responseUrls = append(responseUrls, userURL)
	}

	response := &proto.GetUserURLsResponse{}
	response.SetUrls(responseUrls)
	return response, nil
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
			resp := &proto.GetStatsResponse{}
			resp.SetTotalUsers(int32(stats.Users))
			resp.SetTotalUrls(int32(stats.Urls))
			return resp, nil
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

		shortURL, err := s.svc.ShortenURL(ctx, req.GetOriginalUrl(), userIDStr)
		if err != nil && !errors.Is(err, service.ErrConflict) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if errors.Is(err, service.ErrConflict) {
			st := status.New(codes.AlreadyExists, "Conflict")
			resp := &proto.URLCreatorResponse{}
			resp.SetShortenUrl(shortURL)
			stWithDetails, err2 := st.WithDetails(resp)
			if err2 != nil {
				return nil, status.Error(codes.Internal, "cannot attach details")
			}
			return nil, stWithDetails.Err()
		} else {
			resp := &proto.URLCreatorResponse{}
			resp.SetShortenUrl(shortURL)
			return resp, nil
		}
	}

	return nil, status.Error(codes.FailedPrecondition, "No metadata in context")
}
func (s *Server) Ping(ctx context.Context, _ *proto.PingRequest) (*proto.PingResponse, error) {
	if s.ping == nil {
		return nil, status.Error(codes.FailedPrecondition, "Database is not initialized ")
	}
	err := s.ping.Ping(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &proto.PingResponse{}
	resp.SetStatus("OK")

	return resp, nil
}

func (s *Server) URLCreatorJSON(ctx context.Context, req *proto.URLCreatorJSONRequest) (*proto.URLCreatorJSONResponse, error) {
	var userIDStr string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if val := md.Get("userID"); len(val) != 0 {
			userIDStr = val[0]
		} else {
			return nil, status.Error(codes.PermissionDenied, "Unauthorized")
		}

		shortURL, err := s.svc.ShortenURL(ctx, req.GetJsonOriginalUrl(), userIDStr)
		if err != nil && !errors.Is(err, service.ErrConflict) {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if errors.Is(err, service.ErrConflict) {
			st := status.New(codes.AlreadyExists, "Conflict")
			resp := &proto.URLCreatorJSONResponse{}
			resp.SetJsonResult(shortURL)
			stWithDetails, err2 := st.WithDetails(resp)
			if err2 != nil {
				return nil, status.Error(codes.Internal, "cannot attach details")
			}
			return nil, stWithDetails.Err()
		} else {
			resp := &proto.URLCreatorJSONResponse{}
			resp.SetJsonResult(shortURL)
			return resp, nil
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
		origs := make([]string, len(req.GetRequests()))
		for i, item := range req.GetRequests() {
			origs[i] = item.GetOriginalUrl()
		}
		inputURLs := make([]string, len(req.GetRequests()))
		for i, request := range req.GetRequests() {
			inputURLs[i] = request.GetOriginalUrl()
		}

		shortenedURLs, err := s.svc.ShortenURLs(ctx, inputURLs, userIDStr)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		shortenedURLsMap := make(map[string]string, len(shortenedURLs))
		for short, orig := range shortenedURLs {
			shortenedURLsMap[orig] = short
		}

		responseURLs := make([]*proto.URLResponse, len(req.GetRequests()))
		for i, request := range req.GetRequests() {
			shortURL, ok := shortenedURLsMap[request.GetOriginalUrl()]
			if !ok {
				return nil, status.Error(codes.Internal, "Mismatch in shortened URLs")
			}
			res := &proto.URLResponse{}
			res.SetCorrelationId(request.GetCorrelationId())
			res.SetShortUrl(s.cfg.BaseAddress + "/" + shortURL)
			responseURLs[i] = res
		}
		res := &proto.URLCreatorBatchResponse{}
		res.SetResponses(responseURLs)
		return res, nil
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
			URLs:   req.GetUrls(),
			UserID: userIDStr,
		}

		shortenurlhandlers.DeleteTaskCh <- task
	}

	return nil, status.Error(codes.FailedPrecondition, "No metadata in context")

}
