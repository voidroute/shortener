package rpc

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"connectrpc.com/connect"
	shortenerv1 "github.com/voidroute/protos/gen/shortener/v1"
	"github.com/voidroute/protos/gen/shortener/v1/v1connect"
	"github.com/voidroute/shortener/internal/domain"
	"github.com/voidroute/shortener/internal/interceptor"
	"github.com/voidroute/shortener/internal/producer"
	"github.com/voidroute/shortener/internal/repository"
	"github.com/voidroute/shortener/internal/service"
)

type ShortenerServer struct {
	v1connect.UnimplementedShortenerServiceHandler
	svc      service.LinkServiceInterface
	producer *producer.KafkaProducer
}

func NewShortenerServer(svc service.LinkServiceInterface, producer *producer.KafkaProducer) *ShortenerServer {
	return &ShortenerServer{svc: svc, producer: producer}
}

func (s *ShortenerServer) CreateLink(ctx context.Context, req *shortenerv1.CreateLinkRequest) (*shortenerv1.CreateLinkResponse, error) {
	var expiresAt *time.Time
	if req.ExpiresAt != nil {
		t := req.ExpiresAt.AsTime()
		expiresAt = &t
	}

	link, err := s.svc.Create(ctx, req.Url, req.Alias, expiresAt)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidLink):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, repository.ErrCodeExists):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	return &shortenerv1.CreateLinkResponse{ShortUrl: link.Code}, nil
}

func (s *ShortenerServer) GetLink(ctx context.Context, req *shortenerv1.GetLinkRequest) (*shortenerv1.GetLinkResponse, error) {
	link, err := s.svc.Get(ctx, req.Code)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrLinkExpired):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, repository.ErrLinkNotFound):
			return nil, connect.NewError(connect.CodeNotFound, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	go func() {
		ip := ctx.Value(interceptor.ClientIPKey).(string)
		if err = s.producer.SendClickEvent(context.Background(), link, ip); err != nil {
			slog.Warn("Failed to send click event", "error", err)
		}
	}()

	return &shortenerv1.GetLinkResponse{Url: link.URL}, nil
}
