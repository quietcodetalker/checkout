package middleware

import (
	"context"
	"gitlab.ozon.dev/unknownspacewalker/homework3/internal/pkg/kafka/router"
	"log"
	"time"
)

// Logger write message handlers results to stdout.
func Logger(fn router.HandlerFunc) router.HandlerFunc {
	return func(ctx context.Context, topic string, msg []byte) error {
		start := time.Now()
		err := fn(ctx, topic, msg)

		log.Printf("%s %s %v %v", topic, msg, err, time.Now().Sub(start))

		return err
	}
}
