package output

import (
	"context"

	"github.com/mata-elang-stable/sensor-snort-service/internal/pb"
)

type MessengerInterface interface {
	StreamData(ctx context.Context, payload *pb.SensorEvent) error
	Disconnect()
}
