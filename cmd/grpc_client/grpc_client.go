package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	"time"

	"github.com/aseptimu/url-shortener/internal/app/config"
	pb "github.com/aseptimu/url-shortener/internal/app/handlers/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Загружаем конфиг (должен содержать ServerAddress, например "localhost:50051")
	appConf, err := config.NewConfig()
	if err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	dialCtx, dialCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dialCancel()

	conn, err := grpc.DialContext(
		dialCtx,
		appConf.GRPCServerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}
	fmt.Println("connection established")

	defer conn.Close()

	client := pb.NewURLShortenerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pingResp, err := client.Ping(ctx, &pb.Empty{})
	if err != nil {
		log.Printf("could not ping: %v", err)
	} else {
		log.Println("Ping status:", pingResp.GetStatus())
	}

	md := metadata.NewOutgoingContext(ctx, metadata.Pairs("userID", "12345"))

	resp, err := client.URLCreator(md, &pb.URLCreatorRequest{
		OriginalUrl: "http://example.com",
	})
	if err == nil {
		fmt.Println("New short URL:", resp.GetShortenUrl())
		return
	}

	st, ok := status.FromError(err)
	if !ok {
		log.Fatalf("не–gRPC ошибка: %v", err)
	}

	if st.Code() == codes.AlreadyExists {
		for _, d := range st.Details() {
			if info, ok := d.(*pb.URLCreatorResponse); ok {
				fmt.Println("Existing short URL:", info.GetShortenUrl())
				return
			}
		}
		for _, anyMsg := range st.Proto().GetDetails() {
			var info pb.URLCreatorResponse
			if err2 := anypb.UnmarshalTo(anyMsg, &info, proto.UnmarshalOptions{}); err2 == nil {
				fmt.Println("Existing short URL:", info.GetShortenUrl())
				return
			}
		}

		log.Fatalf("не смогли извлечь детали из ошибки: %v", st.Proto().GetDetails())
	}

	log.Fatalf("URLCreator упал с другой ошибкой: %v", st)

}
