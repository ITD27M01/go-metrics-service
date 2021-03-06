package repository

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver

	"github.com/itd27m01/go-metrics-service/internal/models/metrics"
	"github.com/itd27m01/go-metrics-service/pkg/logging/log"
)

const (
	psqlDriverName = "pgx"
)

var (
	_ Store = (*DBStore)(nil)
)

// DBStore implements Store interface to store metrics in database
type DBStore struct {
	connection *sql.DB
}

// NewDBStore creates db store
func NewDBStore(databaseDSN string) (*DBStore, error) {
	var db DBStore

	conn, err := sql.Open(psqlDriverName, databaseDSN)
	if err != nil {
		return nil, err
	}

	db = DBStore{
		connection: conn,
	}

	return &db, nil
}

// UpdateCounterMetric updates counter metric type
func (db *DBStore) UpdateCounterMetric(ctx context.Context, metricName string, metricData metrics.Counter) error {
	var counter metrics.Counter
	row := db.connection.QueryRowContext(ctx,
		"SELECT metric_delta FROM counter WHERE metric_id = $1", metricName)

	err := row.Scan(&counter)
	if !errors.Is(err, nil) && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	counter += metricData
	_, err = db.connection.ExecContext(ctx,
		"INSERT INTO counter (metric_id, metric_delta) VALUES ($1, $2) "+
			"ON CONFLICT (metric_id) DO UPDATE SET metric_delta = $2",
		metricName, counter)

	return err
}

// ResetCounterMetric resets counter to default zero value
func (db *DBStore) ResetCounterMetric(ctx context.Context, metricName string) error {
	var zero metrics.Counter
	_, err := db.connection.ExecContext(ctx,
		"INSERT INTO counter (metric_id, metric_delta) VALUES ($1, $2) "+
			"ON CONFLICT (metric_id) DO UPDATE SET metric_delta = $2",
		metricName, zero)

	return err
}

// UpdateGaugeMetric updates gauge type metric
func (db *DBStore) UpdateGaugeMetric(ctx context.Context, metricName string, metricData metrics.Gauge) error {
	_, err := db.connection.ExecContext(ctx,
		"INSERT INTO gauge (metric_id, metric_value) VALUES ($1, $2) "+
			"ON CONFLICT (metric_id) DO UPDATE SET metric_value = $2",
		metricName, metricData)

	return err
}

// GetMetric return metric by name
func (db *DBStore) GetMetric(ctx context.Context, metricName string, metricType string) (*metrics.Metric, error) {
	metric := metrics.Metric{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case metrics.MetricTypeCounter:
		var counter metrics.Counter
		row := db.connection.QueryRowContext(ctx,
			"SELECT metric_delta FROM counter WHERE metric_id = $1", metricName)

		err := row.Scan(&counter)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrMetricNotFound
		case !errors.Is(err, nil):
			return nil, err
		}
		metric.Delta = &counter
	case metrics.MetricTypeGauge:
		var gauge metrics.Gauge
		row := db.connection.QueryRowContext(ctx,
			"SELECT metric_value FROM gauge WHERE metric_id = $1", metricName)

		err := row.Scan(&gauge)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrMetricNotFound
		case !errors.Is(err, nil):
			return nil, err
		}
		metric.Value = &gauge
	default:
		return nil, ErrMetricNotFound
	}

	return &metric, nil
}

// UpdateMetrics update number of metrics
func (db *DBStore) UpdateMetrics(ctx context.Context, metricsBatch []*metrics.Metric) error {
	tx, err := db.connection.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	stmtInsertGauge, err := tx.Prepare("INSERT INTO gauge (metric_id, metric_value) VALUES ($1, $2) " +
		"ON CONFLICT (metric_id) DO UPDATE SET metric_value = $2")
	if err != nil {
		return err
	}
	defer func(stmtInsertGauge *sql.Stmt) {
		if err := stmtInsertGauge.Close(); err != nil {
			log.Error().Err(err).Msg("Failed to close insert statement")
		}
	}(stmtInsertGauge)

	stmtSelectCounter, err := tx.Prepare("SELECT metric_delta FROM counter WHERE metric_id = $1")
	if err != nil {
		return err
	}
	defer func(stmtSelectCounter *sql.Stmt) {
		if err := stmtSelectCounter.Close(); err != nil {
			log.Error().Err(err).Msgf("Failed to close insert statement")
		}
	}(stmtSelectCounter)

	stmtInsertCounter, err := tx.Prepare("INSERT INTO counter (metric_id, metric_delta) VALUES ($1, $2) " +
		"ON CONFLICT (metric_id) DO UPDATE SET metric_delta = $2")
	if err != nil {
		return err
	}
	defer func(stmtInsertCounter *sql.Stmt) {
		if err := stmtInsertCounter.Close(); err != nil {
			log.Error().Err(err).Msgf("Failed to close insert statement")
		}
	}(stmtInsertCounter)

	for _, metric := range metricsBatch {
		switch {
		case metric.MType == metrics.MetricTypeGauge:
			if _, err := stmtInsertGauge.Exec(metric.ID, *(metric.Value)); err != nil {
				if err := tx.Rollback(); err != nil {
					log.Error().Err(err).Msg("unable to rollback transaction")
				}

				return err
			}
		case metric.MType == metrics.MetricTypeCounter:
			var counter metrics.Counter
			query := stmtSelectCounter.QueryRow(metric.ID)

			err = query.Scan(&counter)
			if !errors.Is(err, nil) && !errors.Is(err, sql.ErrNoRows) {
				if err := tx.Rollback(); err != nil {
					log.Error().Err(err).Msgf("unable to rollback transaction")
				}

				return err
			}

			counter += *(metric.Delta)

			if _, err := stmtInsertCounter.Exec(metric.ID, counter); err != nil {
				if err := tx.Rollback(); err != nil {
					log.Error().Err(err).Msgf("unable to rollback transaction")
				}

				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// GetMetrics returns all of stored metrics
func (db *DBStore) GetMetrics(ctx context.Context) (map[string]*metrics.Metric, error) {
	metricsMap := make(map[string]*metrics.Metric)

	counters, err := db.connection.QueryContext(ctx,
		"SELECT metric_id,metric_delta FROM counter")

	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Couldn't close rows")
		}
	}(counters)

	for counters.Next() {
		var counter metrics.Counter
		metric := metrics.Metric{
			MType: metrics.MetricTypeCounter,
			Delta: &counter,
		}
		err = counters.Scan(&metric.ID, metric.Delta)
		if err != nil {
			return nil, err
		}

		metricsMap[metric.ID] = &metric
	}

	err = counters.Err()
	if err != nil {
		return nil, err
	}

	gauges, err := db.connection.QueryContext(ctx,
		"SELECT metric_id,metric_value FROM gauge")

	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msgf("Couldn't close rows")
		}
	}(gauges)

	for gauges.Next() {
		var gauge metrics.Gauge
		metric := metrics.Metric{
			MType: metrics.MetricTypeGauge,
			Value: &gauge,
		}

		err = gauges.Scan(&metric.ID, metric.Value)
		if err != nil {
			return nil, err
		}

		metricsMap[metric.ID] = &metric
	}

	err = gauges.Err()
	if err != nil {
		return nil, err
	}

	return metricsMap, nil
}

// Ping checks that underlying store is alive
func (db *DBStore) Ping(ctx context.Context) error {
	return db.connection.PingContext(ctx)
}

// Close closes database connection
func (db *DBStore) Close() error {
	log.Info().Msgf("Close database connection")

	return db.connection.Close()
}
