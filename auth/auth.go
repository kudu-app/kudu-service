package auth

import (
	"golang.org/x/net/context"

	"github.com/knq/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// key is context key.
type key int

// UserIDKey is context key for user ID.
const UserIDKey key = 1

// UnaryInterceptor is grpc interceptor that responsible to handle authentication,
// by parsing the token field from metadata.
// func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
// 	var err error

// 	md, ok := metadata.FromContext(ctx)
// 	if !ok {
// 		return nil, grpc.Errorf(codes.DataLoss, "auth unary interceptor: failed to get metadata")
// 	}

// 	var userID string
// 	if token, ok := md["token"]; ok {
// 		userID, err = jwt.PeekPayloadField([]byte(token[0]), "uid")
// 		if err != nil {
// 			return nil, err
// 		}
// 		newCtx := context.WithValue(ctx, UserIDKey, userID)
// 		return handler(newCtx, req)
// 	}
// 	return nil, grpc.Errorf(codes.Unauthenticated, "authentication required")
// }

func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var err error

		md, ok := metadata.FromContext(ctx)
		if !ok {
			return nil, grpc.Errorf(codes.DataLoss, "auth unary interceptor: failed to get metadata")
		}

		var userID string
		if token, ok := md["token"]; ok {
			userID, err = jwt.PeekPayloadField([]byte(token[0]), "uid")
			if err != nil {
				return nil, err
			}
			newCtx := context.WithValue(ctx, UserIDKey, userID)
			return handler(newCtx, req)
		}
		return nil, grpc.Errorf(codes.Unauthenticated, "authentication required")
	}
}
