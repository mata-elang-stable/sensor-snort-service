package listener

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync/atomic"
	"time"

	"github.com/mata-elang-stable/sensor-snort-service/internal/logger"
	"github.com/mata-elang-stable/sensor-snort-service/internal/processor"

	"github.com/mata-elang-stable/sensor-snort-service/internal/config"
	"github.com/mata-elang-stable/sensor-snort-service/internal/queue"
	"github.com/mata-elang-stable/sensor-snort-service/internal/types"
	"github.com/mata-elang-stable/sensor-snort-service/internal/util"
	"github.com/nxadm/tail"
)

type FileListener struct {
	filename     string
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
		filename: filename,
		tail:     t,
	}, nil
}

func (f *FileListener) clearFileContent() {
	file, err := os.OpenFile(f.filename, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.WithFields(logger.Fields{
			"error":   err,
			"package": "file_listener",
		}).Warnln("failed to clear file content")
		return
	}
	defer file.Close()

	log.Infof("Cleared file content: %s", f.filename)
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
		f.clearFileContent()
		f.linesThisSec.Store(0)
		f.linesPerSec.Store(0)

		// Wait for all goroutines to finish
		log.WithField("package", "file_listener").Infoln("Shutting down ListenFile process.")
	}()

MainLoop:
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

				continue MainLoop
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
