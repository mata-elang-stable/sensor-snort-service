package listener

import (
	"context"
	"encoding/json"
	"errors"
	"gitlab.com/mata-elang/v2/mes-snort/internal/logger"
	"gitlab.com/mata-elang/v2/mes-snort/internal/processor"
	"sync/atomic"
	"time"

	"github.com/nxadm/tail"
	"gitlab.com/mata-elang/v2/mes-snort/internal/config"
	"gitlab.com/mata-elang/v2/mes-snort/internal/queue"
	"gitlab.com/mata-elang/v2/mes-snort/internal/types"
	"gitlab.com/mata-elang/v2/mes-snort/internal/util"
)

type FileListener struct {
	tail         *tail.Tail
	linesPerSec  atomic.Int64
	linesThisSec atomic.Int64
}

func NewFileListener(filename string) (*FileListener, error) {
	t, err := tail.TailFile(filename, config.GetConfig().GetTailConfig())
	if err != nil {
		return nil, err
	}

	return &FileListener{
		tail: t,
	}, nil
}

func (f *FileListener) Start(ctx context.Context, q *queue.EventBatchQueue) error {
	if f.tail == nil {
		return errors.New("listener is not initialized properly")
	}

	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				util.UpdateAndReset(&f.linesPerSec, &f.linesThisSec)
			}
		}
	}()

	conf := config.GetConfig()

	defer func() {
		ticker.Stop()
		f.tail.Cleanup()
		f.linesThisSec.Store(0)
		f.linesPerSec.Store(0)

		// Wait for all goroutines to finish
		log.WithField("package", "file_listener").Infoln("Shutting down ListenFile process.")
	}()

	for line := range f.tail.Lines {
		select {
		case <-ctx.Done():
			log.WithField("package", "file_listener").Infoln("Context is done, stopping the listener.")
			return nil
		default:
			var payload types.SnortAlert
			if err := json.Unmarshal([]byte(line.Text), &payload); err != nil {
				log.WithFields(logger.Fields{
					"error":   err,
					"package": "file_listener",
				}).Debugln("failed to parse log line")
				return nil
			}

			payload.Metadata.SensorID = conf.ClientConfig.SensorID
			payload.Metadata.ReadAt = time.Now().UnixMicro()
			pbRecord, metric := processor.ConvertSnortAlertToSensorEvent(&payload)

			q.AddRecordToQueue(pbRecord, metric)
			f.linesThisSec.Add(1)
		}
	}

	return nil
}

// GetEventReadPerSecond returns the number of events read per second.
func (f *FileListener) GetEventReadPerSecond() int64 {
	return f.linesPerSec.Load()
}

// Stop stops the file listener.
func (f *FileListener) Stop() error {
	return f.tail.Stop()
}
