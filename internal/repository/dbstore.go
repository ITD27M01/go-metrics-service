package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/itd27m01/go-metrics-service/db/migrations"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver
)

const (
	psqlDriverName       = "pgx"
	psqlMetricsTableName = "metrics"
	migrationSourceName  = "go-bindata"
)

type DBStore struct {
	connection *sql.DB
}

func NewDBStore(databaseDSN string) (*DBStore, error) {
	var db DBStore

	conn, err := sql.Open(psqlDriverName, databaseDSN)
	if err != nil {
		return nil, err
	}

	db = DBStore{
		connection: conn,
	}

	if err := db.migrate(); err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *DBStore) UpdateCounterMetric(ctx context.Context, metricName string, metricData metrics.Counter) error {
	_, err := db.connection.ExecContext(ctx,
		"INSERT INTO counter (metric_id, metric_delta) VALUES ($1, $2) "+
			"ON CONFLICT (metric_id) DO UPDATE SET metric_delta = $2",
		metricName, metricData)

	return err
}

func (db *DBStore) ResetCounterMetric(ctx context.Context, metricName string) error {
	var zero metrics.Counter
	_, err := db.connection.ExecContext(ctx,
		"INSERT INTO counter (metric_id, metric_delta) VALUES ($1, $2) "+
			"ON CONFLICT (metric_id) DO UPDATE SET metric_delta = $2",
		metricName, zero)

	return err
}

func (db *DBStore) UpdateGaugeMetric(ctx context.Context, metricName string, metricData metrics.Gauge) error {
	_, err := db.connection.ExecContext(ctx,
		"INSERT INTO gauge (metric_id, metric_value) VALUES ($1, $2) "+
			"ON CONFLICT (metric_id) DO UPDATE SET metric_value = $2",
		metricName, metricData)

	return err
}

func (db *DBStore) GetMetric(ctx context.Context, metricName string, metricType string) (*metrics.Metric, bool, error) {
	metric := metrics.Metric{
		ID:    metricName,
		MType: metricType,
	}

	switch metricType {
	case metrics.CounterMetricTypeName:
		var counter metrics.Counter
		row := db.connection.QueryRowContext(ctx,
			"SELECT metric_delta FROM counter WHERE metric_id = $1", metricName)

		err := row.Scan(&counter)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, false, nil
		case !errors.Is(err, nil):
			return nil, false, err
		}
		metric.Delta = &counter
	case metrics.GaugeMetricTypeName:
		var gauge metrics.Gauge
		row := db.connection.QueryRowContext(ctx,
			"SELECT metric_value FROM gauge WHERE metric_id = $1", metricName)

		err := row.Scan(&gauge)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, false, nil
		case !errors.Is(err, nil):
			return nil, false, err
		}
		metric.Value = &gauge
	default:
		return nil, false, nil
	}

	return &metric, true, nil
}

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
			log.Printf("Couldn't close rows: %q", err)
		}
	}(counters)

	for counters.Next() {
		var metric metrics.Metric
		err = counters.Scan(&metric.ID, metrics.CounterMetricTypeName, metric.Delta)
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
			log.Printf("Couldn't close rows: %q", err)
		}
	}(gauges)

	for gauges.Next() {
		var metric metrics.Metric
		err = gauges.Scan(&metric.ID, metrics.GaugeMetricTypeName, metric.Value)
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

func (db *DBStore) Ping(ctx context.Context) error {
	return db.connection.PingContext(ctx)
}

func (db *DBStore) Close() error {
	log.Println("Close database connection")

	return db.connection.Close()
}

func (db *DBStore) migrate() error {
	data := bindata.Resource(migrations.AssetNames(), migrations.Asset)

	sourceDriver, err := bindata.WithInstance(data)
	if err != nil {
		return err
	}

	dbDriver, err := postgres.WithInstance(db.connection, &postgres.Config{})
	if err != nil {
		return err
	}

	migration, err := migrate.NewWithInstance(migrationSourceName, sourceDriver, psqlDriverName, dbDriver)
	if err != nil {
		return err
	}

	if err := migration.Up(); !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}
