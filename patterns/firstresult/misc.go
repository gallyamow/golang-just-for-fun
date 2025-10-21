package firstresult

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type ResolvedGetter func(ctx context.Context, address string, key string) string

var resolvedGetter ResolvedGetter = func(ctx context.Context, address string, key string) string {
	time.Sleep(time.Duration(rand.Intn(200)+10) * time.Millisecond)
	return fmt.Sprintf("val of %s from %s", key, address)
}

type Getter func(ctx context.Context, address string, key string) (string, error)

var RandGetter Getter = func(ctx context.Context, address string, key string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	if rand.Intn(10) == 0 {
		return "", errors.New("failed")
	}

	time.Sleep(time.Duration(rand.Intn(200)+10) * time.Millisecond)
	return fmt.Sprintf("val of %s from %s", key, address), nil
}

var FailedGetter Getter = func(ctx context.Context, address string, key string) (string, error) {
	return "", errors.New("failed")
}
