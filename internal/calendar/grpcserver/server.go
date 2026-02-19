package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mrvin/calendar/internal/calendar/auth"
	"github.com/mrvin/calendar/internal/logger"
	"github.com/mrvin/calendar/internal/storage"
	"github.com/mrvin/calendar/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type Conf struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type Server struct {
	serv    *grpc.Server
	conn    net.Listener
	addr    string
	storage storage.Storage
	auth    *auth.Auth
}

func New(ctx context.Context, conf *Conf, storage storage.Storage, auth *auth.Auth) (*Server, error) {
	var server Server

	server.storage = storage
	server.auth = auth

	var err error
	lc := net.ListenConfig{} //nolint:exhaustruct
	server.addr = net.JoinHostPort(conf.Host, conf.Port)
	server.conn, err = lc.Listen(ctx, "tcp", server.addr)
	if err != nil {
		return nil, fmt.Errorf("establish tcp connection: %w", err)
	}

	server.serv = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggerInterceptor,
			server.authInterceptor,
		),
	)
	api.RegisterCalendarServiceServer(server.serv, &server)

	reflection.Register(server.serv)

	return &server, nil
}

func (s *Server) Run(ctx context.Context) {
	ctx, cancel := signal.NotifyContext(
		ctx,
		os.Interrupt,    // SIGINT, (Control-C)
		syscall.SIGTERM, // systemd
		syscall.SIGQUIT,
	)

	go func() {
		defer cancel()
		if err := s.serv.Serve(s.conn); err != nil {
			slog.Error("Failed to start gRPC server: " + err.Error())
			return
		}
	}()
	slog.Info("Start gRPC server: " + s.addr)

	<-ctx.Done()

	s.serv.GracefulStop()
	s.conn.Close()

	slog.Info("Stop gRPC server")
}

func loggerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	var addr string
	p, ok := peer.FromContext(ctx)
	if !ok {
		slog.Warn("Cant get perr")
	} else {
		addr = p.Addr.String()
	}
	slog.Info("Request gRPC",
		slog.String("addr", addr),
		slog.String("Method", info.FullMethod),
	)
	// Last but super important, execute the handler so that the actually gRPC request is also performed
	return handler(ctx, req)
}

func (s *Server) authInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (any, error) {
	publicMethods := map[string]bool{
		"/calendar.CalendarService/Register": true,
		"/calendar.CalendarService/Login":    true,
	}
	if publicMethods[info.FullMethod] {
		return handler(ctx, req)
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "metadata is empty")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "metadata does not contain an authorization token")
	}

	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader[0], bearerPrefix) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid authorization header format, expected 'Bearer <token>'")
	}
	tokenString := strings.TrimPrefix(authHeader[0], bearerPrefix)

	claims, err := s.auth.ParseToken(tokenString)
	if err != nil {
		//TODO: обробатывать внутринию ошибку
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	username := claims["username"]

	ctx = logger.WithUsername(ctx, username.(string)) //nolint:forcetypeassert

	return handler(ctx, req)
}
