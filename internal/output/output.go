package output

import (
	"context"
	"gitlab.com/mata-elang/v2/mes-snort/internal/pb"
)

type MessengerInterface interface {
	StreamData(ctx context.Context, payload *pb.SensorEvent) error
	Disconnect()
}
