package middleware

import (
	"context"

	"github.com/grnsv/shortener/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCAuthenticateInterceptor returns a gRPC unary interceptor that authenticates users using JWT from metadata.
// If token is missing or invalid, generates a new userID and continues (like HTTP middleware).
func GRPCAuthenticateInterceptor(key string, logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var token string
		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get(cookieName)
			if len(values) > 0 {
				token = values[0]
			}
		}

		var userID string
		claims, err := parseClaims(token, key)
		if err != nil {
			logger.Debug(err)
			userID, err = generateUserID()
			if err != nil {
				logger.Error(err)
				return nil, status.Error(codes.Internal, "Failed to generate userID")
			}
			token, err = BuildJWTString(key, userID)
			if err != nil {
				logger.Error(err)
				return nil, status.Error(codes.Internal, "Failed to build token")
			}
		} else {
			userID = claims.Subject
			if userID == "" {
				return nil, status.Error(codes.Unauthenticated, "Empty userID")
			}
		}

		ctx = context.WithValue(ctx, UserIDContextKey, userID)
		md = metadata.Pairs(cookieName, token)
		ctx = metadata.NewOutgoingContext(ctx, md)
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		if err := grpc.SendHeader(ctx, md); err != nil {
			logger.Errorf("Failed to send headers: %v", err)
		}

		return resp, nil
	}
}
