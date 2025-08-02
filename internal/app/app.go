package app

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/misshanya/pubbet/internal/config"
	"github.com/misshanya/pubbet/internal/service"
	handler "github.com/misshanya/pubbet/internal/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
)

type App struct {
	cfg     *config.Config
	l       *slog.Logger
	lis     *net.Listener
	grpcSrv *grpc.Server
}

// InterceptorLogger adapts slog logger to interceptor logger.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func New(cfg *config.Config, l *slog.Logger) (*App, error) {
	a := &App{
		cfg: cfg,
		l:   l,
	}

	lis, err := net.Listen("tcp", cfg.ServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}
	a.lis = &lis

	// Configure interceptor logger
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	// Create a gRPC server
	a.grpcSrv = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(InterceptorLogger(a.l), opts...),
		),
	)

	// Use gRPC reflection
	reflection.Register(a.grpcSrv)

	svc := service.New(l)

	handler.NewHandler(l, a.grpcSrv, svc)

	return a, nil
}

func (a *App) Start(errChan chan<- error) {
	a.l.Info("starting server", slog.String("addr", a.cfg.ServerAddress))
	if err := a.grpcSrv.Serve(*a.lis); err != nil {
		errChan <- err
	}
}

func (a *App) Stop() {
	a.l.Info("[!] Shutting down...")
	a.grpcSrv.Stop()
}
