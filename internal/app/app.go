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

// New creates and initializes a new instance of App
func New(cfg *config.Config, l *slog.Logger) (*App, error) {
	a := &App{
		cfg: cfg,
		l:   l,
	}

	if err := a.initListener(); err != nil {
		return nil, err
	}

	a.initGRPCServer()

	svc := service.New(l)
	handler.NewHandler(l, a.grpcSrv, svc)

	return a, nil
}

// Start performs a start of all functional services
func (a *App) Start(errChan chan<- error) {
	a.l.Info("starting server", slog.String("addr", a.cfg.ServerAddress))
	if err := a.grpcSrv.Serve(*a.lis); err != nil {
		errChan <- err
	}
}

// Stop performs a graceful shutdown for all components
func (a *App) Stop() {
	a.l.Info("[!] Shutting down...")
	a.grpcSrv.Stop()
}

// initGRPCServer sets up a gRPC server with interceptor logger
func (a *App) initGRPCServer() {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
	}

	a.grpcSrv = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(InterceptorLogger(a.l), opts...),
		),
	)

	reflection.Register(a.grpcSrv)
}

// initListener sets up a tcp listener ready for gRPC
func (a *App) initListener() error {
	lis, err := net.Listen("tcp", a.cfg.ServerAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	a.lis = &lis
	return nil
}
