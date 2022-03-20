package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/itd27m01/go-metrics-service/internal/pkg/logging/log"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
)

const (
	fileMode = 0640
)

type FileStore struct {
	file         *os.File
	syncChannel  chan struct{}
	metricsCache map[string]*metrics.Metric
	mu           sync.Mutex
}

func NewFileStore(filePath string, syncChannel chan struct{}) (*FileStore, error) {
	var fs FileStore

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, fileMode)
	if err != nil {
		return nil, err
	}

	metricsCache := make(map[string]*metrics.Metric)
	fs = FileStore{
		file:         file,
		syncChannel:  syncChannel,
		metricsCache: metricsCache,
	}

	return &fs, nil
}

func (fs *FileStore) UpdateCounterMetric(_ context.Context, metricName string, metricData metrics.Counter) error {
	fs.mu.Lock()
	defer fs.sync()
	defer fs.mu.Unlock()

	currentMetric, ok := fs.metricsCache[metricName]
	switch {
	case ok && currentMetric.Delta != nil:
		*(currentMetric.Delta) += metricData
	case ok && currentMetric.Delta == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		fs.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.MetricTypeCounter,
			Delta: &metricData,
		}
	}

	return nil
}

func (fs *FileStore) ResetCounterMetric(_ context.Context, metricName string) error {
	fs.mu.Lock()
	defer fs.sync()
	defer fs.mu.Unlock()

	var zero metrics.Counter
	currentMetric, ok := fs.metricsCache[metricName]
	switch {
	case ok && currentMetric.Delta != nil:
		*(currentMetric.Delta) = zero
	case ok && currentMetric.Delta == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		fs.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.MetricTypeCounter,
			Delta: &zero,
		}
	}

	return nil
}

func (fs *FileStore) UpdateGaugeMetric(_ context.Context, metricName string, metricData metrics.Gauge) error {
	fs.mu.Lock()
	defer fs.sync()
	defer fs.mu.Unlock()

	currentMetric, ok := fs.metricsCache[metricName]
	switch {
	case ok && currentMetric.Value != nil:
		*(currentMetric.Value) = metricData
	case ok && currentMetric.Value == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		fs.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.MetricTypeGauge,
			Value: &metricData,
		}
	}

	return nil
}

func (fs *FileStore) UpdateMetrics(_ context.Context, metricsBatch []*metrics.Metric) error {
	fs.mu.Lock()
	defer fs.sync()
	defer fs.mu.Unlock()

	for _, metric := range metricsBatch {
		currentMetric, ok := fs.metricsCache[metric.ID]
		switch {
		case ok && metric.MType == metrics.MetricTypeGauge && currentMetric.Value != nil:
			currentMetric.Value = metric.Value
		case ok && metric.MType == metrics.MetricTypeGauge && currentMetric.Value == nil:
			return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metric.ID, currentMetric.MType)
		case ok && metric.MType == metrics.MetricTypeCounter && currentMetric.Delta != nil:
			*(currentMetric.Delta) += *(metric.Delta)
		case ok && metric.MType == metrics.MetricTypeCounter && currentMetric.Delta == nil:
			return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metric.ID, currentMetric.MType)
		default:
			fs.metricsCache[metric.ID] = metric
		}
	}

	return nil
}

func (fs *FileStore) GetMetric(_ context.Context, metricName string, _ string) (*metrics.Metric, error) {
	metric, ok := fs.metricsCache[metricName]
	if !ok {
		return nil, ErrMetricNotFound
	}

	return metric, nil
}

func (fs *FileStore) GetMetrics(_ context.Context) (map[string]*metrics.Metric, error) {
	return fs.metricsCache, nil
}

func (fs *FileStore) sync() {
	fs.syncChannel <- struct{}{}
}

func (fs *FileStore) Ping(_ context.Context) error {
	_, err := fs.file.Stat()

	return err
}

func (fs *FileStore) Close() error {
	if err := fs.SaveMetrics(); err != nil {
		log.Error().Err(err).Msg("Something went wrong durin metrics preserve")
	}

	if err := fs.file.Sync(); err != nil {
		log.Error().Err(err).Msg("Failed to sync metrics")
	}

	return fs.file.Close()
}

func (fs *FileStore) LoadMetrics() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	jsonDecoder := json.NewDecoder(fs.file)

	log.Info().Msgf("Load metrics from %s", fs.file.Name())

	return jsonDecoder.Decode(&(fs.metricsCache))
}

func (fs *FileStore) SaveMetrics() (err error) {
	log.Info().Msgf("Dump metrics to %s", fs.file.Name())

	fs.mu.Lock()
	defer fs.mu.Unlock()

	const (
		offset     = 0
		whence     = 0
		truncateTo = 0
	)
	_, err = fs.file.Seek(offset, whence)
	if err != nil {
		return err
	}

	if err := fs.file.Truncate(truncateTo); err != nil {
		return err
	}

	encoder := json.NewEncoder(fs.file)

	return encoder.Encode(&fs.metricsCache)
}
