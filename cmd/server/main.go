package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	shortenerv1 "github.com/voidroute/protos/gen/shortener/v1"
	"github.com/voidroute/protos/gen/shortener/v1/v1connect"
	"github.com/voidroute/shortener/internal/cache"
	"github.com/voidroute/shortener/internal/config"
	"github.com/voidroute/shortener/internal/geo"
	"github.com/voidroute/shortener/internal/interceptor"
	"github.com/voidroute/shortener/internal/producer"
	"github.com/voidroute/shortener/internal/repository"
	"github.com/voidroute/shortener/internal/rpc"
	"github.com/voidroute/shortener/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Init()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	repo, err := initRepository(cfg.Database)
	if err != nil {
		slog.Error("Failed to initialize repository", "error", err)
		os.Exit(1)
	}

	var appCache cache.Cache

	if cfg.Cache.UseRedis {
		redisCache, err := cache.NewRedisCache(cfg.Cache.RedisAddr, cfg.Cache.TTL)
		if err != nil {
			slog.Error("Failed to connect to redis", "error", err)
			os.Exit(1)
		}

		slog.Info("Using Redis cache")
		appCache = redisCache
	} else {
		slog.Info("Using in-memory cache")
		appCache = cache.NewInMemoryCache(cfg.Cache.TTL, cfg.Cache.CleanupInterval)
	}

	geoIP, err := geo.NewGeoIP(cfg.App.GeoIPPath)
	if err != nil {
		slog.Error("Failed to load GeoIP database", "error", err)
		os.Exit(1)
	}

	svc := service.NewLinkService(repo, appCache)
	kafkaProducer := producer.NewKafkaProducer(cfg.Kafka.Addr, cfg.Kafka.Topic, geoIP)

	interceptors := connect.WithInterceptors(
		interceptor.NewClientIPInterceptor(),
		validate.NewInterceptor(),
	)

	mux := http.NewServeMux()
	server := rpc.NewShortenerServer(svc, kafkaProducer)
	path, handler := v1connect.NewShortenerServiceHandler(server, interceptors)
	mux.Handle(path, handler)

	mux.HandleFunc("GET /{code}", func(w http.ResponseWriter, r *http.Request) {
		code := r.PathValue("code")
		resp, err := server.GetLink(r.Context(), &shortenerv1.GetLinkRequest{Code: code})

		if err != nil {
			var connectErr *connect.Error
			if errors.As(err, &connectErr) {
				switch connectErr.Code() {
				case connect.CodeNotFound:
					http.Error(w, "Link not found", http.StatusNotFound)
				case connect.CodeFailedPrecondition:
					http.Error(w, "Link expired", http.StatusGone)
				default:
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, resp.Url, http.StatusFound)
	})

	p := new(http.Protocols)
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)

	hServer := &http.Server{
		Addr:      cfg.Server.Addr,
		Handler:   mux,
		Protocols: p,
	}

	errChan := make(chan error, 1)

	go func() {
		slog.Info("Starting server", "address", hServer.Addr)
		if err = hServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		slog.Error("Server failed", "error", err)
	case <-quit:
		slog.Info("Shutting down server...")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err = hServer.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	if err = appCache.Close(); err != nil {
		slog.Error("Failed to close cache", "error", err)
	} else {
		slog.Info("Cache closed")
	}

	if err = repo.Close(); err != nil {
		slog.Error("Failed to close repository", "error", err)
	} else {
		slog.Info("Database connection closed")
	}

	if err = kafkaProducer.Close(); err != nil {
		slog.Error("Failed to close kafka producer", "error", err)
	} else {
		slog.Info("Kafka producer closed")
	}

	if err = geoIP.Close(); err != nil {
		slog.Error("Failed to close geoip", "error", err)
	} else {
		slog.Info("GeoIP closed")
	}

	slog.Info("Server stopped")
}

func initRepository(cfg config.DatabaseConfig) (repository.LinkRepository, error) {
	if cfg.UsePostgres {
		repo, err := repository.NewPostgresRepository(cfg)

		if err != nil {
			return nil, err
		}

		slog.Info("Using PostgreSQL repository. Applying migrations...")

		err = repository.Migrate(repo, "migrations")
		if err != nil {
			return nil, fmt.Errorf("failed to apply migrations: %w", err)
		}

		return repo, nil
	}

	slog.Info("Using in-memory repository")
	return repository.NewInMemoryRepository(), nil
}
