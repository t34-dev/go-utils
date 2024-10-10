package auth

import (
	"context"
	"strings"

	"github.com/t34-dev/go-utils/pkg/sys"
	"github.com/t34-dev/go-utils/pkg/sys/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const userClaimsKey contextKey = "user_claims"

type UserClaims struct {
	UserID string
	Role   string
}

func JWTAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, sys.NewError("metadata is not provided", codes.Unauthenticated)
		}

		authHeader, ok := md["authorization"]
		if !ok || len(authHeader) == 0 {
			return nil, sys.NewError("authorization token is not provided", codes.Unauthenticated)
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")

		claims, err := extractClaimsFromToken(token)
		if err != nil {
			return nil, sys.NewError("invalid token", codes.Unauthenticated)
		}

		newCtx := context.WithValue(ctx, userClaimsKey, claims)

		return handler(newCtx, req)
	}
}

func extractClaimsFromToken(tokenString string) (*UserClaims, error) {
	// Here should be the logic for extracting data from the token
	// For example, we'll just create dummy data
	return &UserClaims{
		UserID: "123",
		Role:   "user",
	}, nil
}

func GetUserClaimsFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(userClaimsKey).(*UserClaims)
	return claims, ok
}
