package listener

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"os"
	"sync/atomic"
	"time"

	"github.com/mata-elang-stable/sensor-snort-service/internal/config"
	"github.com/mata-elang-stable/sensor-snort-service/internal/logger"
	"github.com/mata-elang-stable/sensor-snort-service/internal/processor"
	"github.com/mata-elang-stable/sensor-snort-service/internal/queue"
	"github.com/mata-elang-stable/sensor-snort-service/internal/types"
	"github.com/mata-elang-stable/sensor-snort-service/internal/util"
)

type UnixListener struct {
	listener     net.Listener
	conn         net.Conn
	linesPerSec  atomic.Int64
	linesThisSec atomic.Int64
	socketPath   string
}

func NewUnixListener(socketPath string) (*UnixListener, error) {
	return &UnixListener{
		socketPath: socketPath,
	}, nil
}

func (u *UnixListener) Start(ctx context.Context, q *queue.EventBatchQueue) error {
	// Remove existing socket file if it exists
	if err := os.RemoveAll(u.socketPath); err != nil {
		log.WithFields(logger.Fields{
			"error":   err,
			"package": "unix_listener",
		}).Errorln("failed to remove existing socket file")
		return err
	}

	// Create the listener (server mode)
	listener, err := net.Listen("unix", u.socketPath)
	if err != nil {
		log.WithFields(logger.Fields{
			"error":   err,
			"package": "unix_listener",
		}).Errorln("failed to create unix socket listener")
		return err
	}
	u.listener = listener

	log.WithFields(logger.Fields{
		"package": "unix_listener",
		"socket":  u.socketPath,
	}).Infoln("Unix socket listener created, waiting for Snort to connect...")

	// Set socket permissions so Snort can write to it
	if err := os.Chmod(u.socketPath, 0666); err != nil {
		log.WithFields(logger.Fields{
			"error":   err,
			"package": "unix_listener",
		}).Warnln("failed to set socket permissions")
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				util.UpdateAndReset(&u.linesPerSec, &u.linesThisSec)
			}
		}
	}()

	conf := config.GetConfig()

	defer func() {
		if u.conn != nil {
			u.conn.Close()
		}
		if u.listener != nil {
			u.listener.Close()
		}
		// Clean up socket file
		os.RemoveAll(u.socketPath)
		u.linesThisSec.Store(0)
		u.linesPerSec.Store(0)

		log.WithField("package", "unix_listener").Infoln("Shutting down UnixListener process.")
	}()

	// Accept connection from Snort
	// In a typical scenario where Snort connects and stays connected, we might want a loop here if we expect reconnections.
	// However, following the reference implementation which handles a single connection session (or maybe the loop is implicit in how it runs? No, the reference accepts once).
	// Snort usually keeps the connection open. If it drops, the listener might need to re-accept.
	// For now, I'll stick to the reference implementation from `ravenxcope-sensor-suricata`.
	conn, err := u.listener.Accept()
	if err != nil {
		log.WithFields(logger.Fields{
			"error":   err,
			"package": "unix_listener",
		}).Errorln("failed to accept connection")
		return err
	}
	u.conn = conn

	log.WithFields(logger.Fields{
		"package": "unix_listener",
	}).Infoln("Snort connected to unix socket")

	scanner := bufio.NewScanner(u.conn)

	// Increase buffer size for large JSON payloads
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			log.WithField("package", "unix_listener").Infoln("Context is done, stopping the listener.")
			return nil
		default:
			line := scanner.Text()

			log.WithFields(logger.Fields{
				"package": "unix_listener",
			}).Debugln("read log line")

			var payload types.SnortAlert
			if err := json.Unmarshal([]byte(line), &payload); err != nil {
				log.WithFields(logger.Fields{
					"error":   err,
					"package": "unix_listener",
				}).Debugln("failed to parse log line")
				continue
			}

			payload.Metadata.SensorID = conf.ClientConfig.SensorID
			payload.Metadata.ReadAt = time.Now().UnixMicro()
			pbRecord, metric := processor.ConvertSnortAlertToSensorEvent(&payload)

			if pbRecord == nil || metric == nil {
				log.WithFields(logger.Fields{
					"package": "unix_listener",
					"reason":  "nil return from ConvertSnortAlertToSensorEvent",
				}).Debugln("skipping event")
				continue
			}

			q.AddRecordToQueue(pbRecord, metric)
			u.linesThisSec.Add(1)
		}
	}

	if err := scanner.Err(); err != nil {
		log.WithFields(logger.Fields{
			"error":   err,
			"package": "unix_listener",
		}).Errorln("scanner error")
		return err
	}

	return nil
}

// GetEventReadPerSecond returns the number of events read per second.
func (u *UnixListener) GetEventReadPerSecond() int64 {
	return u.linesPerSec.Load()
}

// Stop stops the unix listener.
func (u *UnixListener) Stop() error {
	if u.conn != nil {
		u.conn.Close()
	}
	if u.listener != nil {
		return u.listener.Close()
	}
	return nil
}
