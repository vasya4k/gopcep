package grpcapi

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Auth struct {
	Tokens []string
}

func extractHeader(ctx context.Context, header string) (string, error) {
	metadata, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "no headers found in request")
	}
	authHeaders, ok := metadata[header]
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "no header: %s in request", header)
	}
	if len(authHeaders) != 1 {
		return "", status.Error(codes.Unauthenticated, "more than 1 header in request")
	}
	return authHeaders[0], nil
}

func deleteHeader(ctx context.Context, header string) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	mdCopy := md.Copy()
	mdCopy[header] = nil
	return metadata.NewIncomingContext(ctx, mdCopy)
}

func (a *Auth) TokenAuth(ctx context.Context) (context.Context, error) {
	auth, err := extractHeader(ctx, "authorization")
	if err != nil {
		return ctx, err
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ctx, status.Error(codes.Unauthenticated, `missing "Bearer " prefix in "Authorization" header`)
	}

	for _, token := range a.Tokens {
		if strings.TrimPrefix(auth, prefix) == token {
			// Delete token from headers
			return deleteHeader(ctx, "authorization"), nil

		}
	}

	return ctx, status.Error(codes.Unauthenticated, "invalid token")
}
