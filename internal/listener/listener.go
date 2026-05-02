package listener

import (
	"context"

	"github.com/mata-elang-stable/sensor-snort-service/internal/logger"
	"github.com/mata-elang-stable/sensor-snort-service/internal/queue"
)

var log = logger.GetLogger()

type Listener interface {
	GetEventReadPerSecond() int64
	Start(ctx context.Context, q *queue.EventBatchQueue) error
	Stop() error
}
