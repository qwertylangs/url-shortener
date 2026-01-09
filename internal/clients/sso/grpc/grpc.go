package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpcLogger "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	ssov1 "github.com/qwertylangs/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api ssov1.AuthClient
}

func New(
	ctx context.Context,
	log *slog.Logger,
	addr string,
	timeout time.Duration,
	retriescount int,
) (*Client, error) {


	retryOpts := []grpcRetry.CallOption{
		grpcRetry.WithMax(uint(retriescount)),
		grpcRetry.WithCodes(codes.DeadlineExceeded, codes.Internal, codes.Unavailable),
		grpcRetry.WithPerRetryTimeout(timeout),
	}

	loggerOpts := []grpcLogger.Option{
		grpcLogger.WithLogOnEvents(grpcLogger.PayloadReceived, grpcLogger.PayloadSent),
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), 
		grpc.WithChainUnaryInterceptor(
			grpcLogger.UnaryClientInterceptor(InterceptorLogger(log), loggerOpts...),
			grpcRetry.UnaryClientInterceptor(retryOpts...),
		))
	
	if err != nil {
		return nil, fmt.Errorf("failed to create grpc client: %w", err)
	}

	return &Client{
		api: ssov1.NewAuthClient(conn),
	}, nil
}

func InterceptorLogger(l *slog.Logger) grpcLogger.Logger {
	return grpcLogger.LoggerFunc(func(ctx context.Context, level grpcLogger.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

func (c *Client) IsAdmin(ctx context.Context, userId int64) (bool, error) {
	const op = "clients.sso.grpc.IsAdmin"
	resp, err := c.api.IsAdmin(ctx, &ssov1.IsAdminRequest{
		UserId: userId,
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	return resp.IsAdmin, nil
}

func (c *Client) Login(ctx context.Context, email string, password string, appId int) (string, error) {
	const op = "clients.sso.grpc.Login"
	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{Email: email, Password: password, AppId: int32(appId)})
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return resp.Token, nil
}

func (c *Client) Register(ctx context.Context, email string, password string, appId int) (int64, error) {
	const op = "clients.sso.grpc.Register"
	resp, err := c.api.Register(ctx, &ssov1.RegisterRequest{Email: email, Password: password})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return resp.UserId, nil
}