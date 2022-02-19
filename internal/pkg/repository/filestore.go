package repository

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

const (
	fileMode = 0640
)

type FileStore struct {
	encoder      *json.Encoder
	file         *os.File
	syncInterval time.Duration
	metricsCache map[string]*metrics.Metric
	mu           sync.Mutex
}

func NewFileStore(filePath string, syncInterval time.Duration) (*FileStore, error) {
	var fs FileStore

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, fileMode)
	if err != nil {
		return nil, err
	}

	metricsCache := make(map[string]*metrics.Metric)
	fs = FileStore{
		encoder:      json.NewEncoder(file),
		file:         file,
		syncInterval: syncInterval,
		metricsCache: metricsCache,
	}

	return &fs, nil
}

func (fs *FileStore) UpdateCounterMetric(metricName string, metricData metrics.Counter) {
	fs.mu.Lock()

	currentMetric, ok := fs.metricsCache[metricName]
	if ok {
		*(currentMetric.Delta) += metricData
	} else {
		fs.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.CounterMetricTypeName,
			Delta: &metricData,
		}
	}

	fs.mu.Unlock()
	fs.flush()
}

func (fs *FileStore) ResetCounterMetric(metricName string) {
	fs.mu.Lock()

	var zero metrics.Counter
	currentMetric, ok := fs.metricsCache[metricName]
	if ok && currentMetric.Delta != nil {
		*(currentMetric.Delta) = zero
	} else {
		fs.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.CounterMetricTypeName,
			Delta: &zero,
		}
	}

	fs.mu.Unlock()
	fs.flush()
}

func (fs *FileStore) UpdateGaugeMetric(metricName string, metricData metrics.Gauge) {
	fs.mu.Lock()

	currentMetric, ok := fs.metricsCache[metricName]
	if ok && currentMetric.Value != nil {
		*(currentMetric.Value) = metricData
	} else {
		fs.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.GaugeMetricTypeName,
			Value: &metricData,
		}
	}

	fs.mu.Unlock()
	fs.flush()
}

func (fs *FileStore) GetMetric(metricName string) (*metrics.Metric, bool) {
	metric, ok := fs.metricsCache[metricName]

	return metric, ok
}

func (fs *FileStore) GetMetrics() map[string]*metrics.Metric {
	return fs.metricsCache
}

func (fs *FileStore) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	log.Println("Closing filestore...")
	if err := fs.encoder.Encode(&fs.metricsCache); err != nil {
		log.Printf("Failed to save metrics: %q", err)
	}

	if err := fs.file.Sync(); err != nil {
		log.Printf("Failed to sync metrics: %q", err)
	}

	return fs.file.Close()
}

func (fs *FileStore) LoadMetrics() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	jsonDecoder := json.NewDecoder(fs.file)

	log.Printf("Load metrics from %s", fs.file.Name())
	err := jsonDecoder.Decode(&(fs.metricsCache))
	if err != nil && err.Error() == "EOF" {
		log.Printf("%s is empty", fs.file.Name())
	}

	return err
}

func (fs *FileStore) RunPreserver(ctx context.Context) {
	log.Printf("Preserve metrics in %s", fs.file.Name())

	if fs.syncInterval == 0 {
		return
	}

	pollTicker := time.NewTicker(fs.syncInterval)
	defer pollTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			fs.saveMetrics()

			return
		case <-pollTicker.C:
			fs.saveMetrics()
		}
	}
}

func (fs *FileStore) saveMetrics() {
	log.Printf("Dump metrics to %s", fs.file.Name())

	fs.mu.Lock()
	defer fs.mu.Unlock()

	if err := fs.encoder.Encode(&fs.metricsCache); err != nil {
		log.Printf("Failed to save metrics: %q", err)
	}
}

func (fs *FileStore) flush() {
	if fs.syncInterval == 0 {
		fs.saveMetrics()
	}
}
