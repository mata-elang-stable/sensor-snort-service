package queue

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mata-elang-stable/sensor-snort-service/internal/output/grpc"
	"golang.org/x/sync/semaphore"

	"github.com/mata-elang-stable/sensor-snort-service/internal/util"

	"github.com/mata-elang-stable/sensor-snort-service/internal/logger"
	"github.com/mata-elang-stable/sensor-snort-service/internal/pb"
)

var log = logger.GetLogger()

// SensorEventRecord represents a sensor event record.
type SensorEventRecord struct {
	Payload   *pb.SensorEvent
	CreatedAt atomic.Int64
	UpdatedAt atomic.Int64
	mu        sync.Mutex
}

// EventBatchQueue represents a queue for storing sensor event records.
type EventBatchQueue struct {
	delta                int
	queue                sync.Map
	latestEventPerSec    atomic.Int64
	EventThisSec         atomic.Int64
	latestBatchPerSec    atomic.Int64
	BatchThisSec         atomic.Int64
	TotalSentEvents      atomic.Int64
	TotalProcessedEvents atomic.Int64
}

// NewEventBatchQueue creates a new instance of EventBatchQueue.
func NewEventBatchQueue() *EventBatchQueue {
	return &EventBatchQueue{
		delta: 1,
	}
}

// AddRecordToQueue adds a sensor event record to the queue.
// If the record already exists, it will update the record with the new metric.
// The record is identified by the SHA256 hash of the metadata.
func (q *EventBatchQueue) AddRecordToQueue(pbRecord *pb.SensorEvent, metric *pb.Metric) {
	now := time.Now().Unix()
	newEventRecord := &SensorEventRecord{Payload: pbRecord}
	newEventRecord.CreatedAt.Store(now)

	selectedRecord, _ := q.queue.LoadOrStore(pbRecord.EventHashSha256, newEventRecord)
	record := selectedRecord.(*SensorEventRecord)
	record.mu.Lock()
	record.Payload.Metrics = append(record.Payload.Metrics, metric)
	record.Payload.EventMetricsCount = int64(len(record.Payload.Metrics))
	record.UpdatedAt.Store(now)
	record.mu.Unlock()

	q.EventThisSec.Add(1)
}

// StartWatcher starts a watcher to process the queue.
// The watcher will send the sensor events to the channel if the record is already older than the delta time.
func (q *EventBatchQueue) StartWatcher(ctx context.Context, handler *grpc.Messenger) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	sem := semaphore.NewWeighted(10)

	for {
		select {
		case <-ctx.Done():
			// Context is done, wait for all goroutines to finish and return
			log.WithField("package", "queue").Infof("Stopping watcher due to context cancellation")
			return nil
		case <-ticker.C:
			util.UpdateAndReset(&q.latestEventPerSec, &q.EventThisSec)
			util.UpdateAndReset(&q.latestBatchPerSec, &q.BatchThisSec)

			eventsBatch := q.processQueue()

			if len(eventsBatch) > 0 {
				go func() {
					// Acquire the semaphore
					if err := sem.Acquire(ctx, 1); err != nil {
						log.Errorf("Failed to acquire semaphore: %v", err)
						return
					}

					defer sem.Release(1)

					if err := handler.StreamData(ctx, eventsBatch); err != nil {
						log.WithField("package", "queue").Errorf("Failed to send data to gRPC server: %v", err)
					}
				}()
			}

		}
	}
}

func (q *EventBatchQueue) processQueue() []*pb.SensorEvent {
	now := time.Now().Unix()

	eventsBatch := make([]*pb.SensorEvent, 0)

	q.queue.Range(func(key, value any) bool {
		record := value.(*SensorEventRecord)
		updatedAt := record.UpdatedAt.Load()

		if now <= updatedAt+int64(q.delta) {
			return true
		}

		record.mu.Lock()
		payloadCopy := *record.Payload
		payloadCopy.Metrics = append([]*pb.Metric{}, record.Payload.Metrics...)
		eventMetricsCount := record.Payload.EventMetricsCount
		record.mu.Unlock()

		//ch <- &payloadCopy

		eventsBatch = append(eventsBatch, &payloadCopy)

		q.updateMetricsCounter(eventMetricsCount)
		q.queue.Delete(key)

		return true
	})

	return eventsBatch
}

func (q *EventBatchQueue) updateMetricsCounter(metricsCount int64) {
	q.BatchThisSec.Add(1)
	q.TotalSentEvents.Add(metricsCount)
	q.TotalProcessedEvents.Store(metricsCount)
}

// GetEventProcessedPerSecond retrieves the latest event per second.
func (q *EventBatchQueue) GetEventProcessedPerSecond() int64 {
	return q.latestEventPerSec.Load()
}

// GetEventBatchSentPerSecond retrieves the latest batch per second.
func (q *EventBatchQueue) GetEventBatchSentPerSecond() int64 {
	return q.latestBatchPerSec.Load()
}

// GetTotalSentEvents retrieves the total number of sent events.
func (q *EventBatchQueue) GetTotalSentEvents() int64 {
	return q.TotalSentEvents.Swap(0)
}

// GetTotalProcessedEvents retrieves the total number of processed events.
func (q *EventBatchQueue) GetTotalProcessedEvents() int64 {
	return q.TotalProcessedEvents.Swap(0)
}

// getSyncMapSize retrieves the size of a sync.Map.
func getSyncMapSize(m *sync.Map) int {
	size := 0
	m.Range(func(key, value interface{}) bool {
		size++
		return true
	})
	return size
}

// GetQueueSize retrieves the size of the queue.
func (q *EventBatchQueue) GetQueueSize() int {
	return getSyncMapSize(&q.queue)
}

// GetEventQueueSize retrieves the size of the event queue.
func (q *EventBatchQueue) GetEventQueueSize() int {
	size := 0
	q.queue.Range(func(key, value any) bool {
		record := value.(*SensorEventRecord)
		record.mu.Lock()
		eventMetricsCount := record.Payload.EventMetricsCount
		record.mu.Unlock()
		size += int(eventMetricsCount)
		return true
	})
	return size
}
