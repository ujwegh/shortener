package context

import (
	"context"
	"github.com/google/uuid"
)

type key string

const userUIDKey key = "userUID"

func WithUserUID(ctx context.Context, userUID *uuid.UUID) context.Context {
	return context.WithValue(ctx, userUIDKey, userUID)
}

func UserUID(ctx context.Context) *uuid.UUID {
	val := ctx.Value(userUIDKey)
	userUID, ok := val.(*uuid.UUID)
	if !ok {
		return nil
	}
	return userUID
}
