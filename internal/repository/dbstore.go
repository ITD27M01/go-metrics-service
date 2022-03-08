package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
	var metricType string

	row := db.connection.QueryRowContext(ctx,
		"SELECT metric_type FROM metrics WHERE metric_id = $1", metricName)
	err := row.Scan(&metricType)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err := db.connection.ExecContext(ctx,
			"INSERT INTO metrics (metric_id, metric_type, metric_delta) VALUES ($1, $2, $3)",
			metricName, metrics.CounterMetricTypeName, metricData)

		return err
	case errors.Is(err, nil):
		if metricType != metrics.CounterMetricTypeName {
			return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, metricType)
		}

		_, err := db.connection.ExecContext(ctx,
			"UPDATE metrics set metric_delta = $1 WHERE metric_id = $2",
			metricData, metricName)

		return err
	default:
		return err
	}
}

func (db *DBStore) ResetCounterMetric(ctx context.Context, metricName string) error {
	var zero metrics.Counter
	var metricType string

	row := db.connection.QueryRowContext(ctx,
		"SELECT metric_type FROM metrics WHERE metric_id = $1", metricName)
	err := row.Scan(&metricType)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err := db.connection.ExecContext(ctx,
			"INSERT INTO metrics (metric_id, metric_type, metric_delta) VALUES ($1, $2, $3)",
			metricName, metrics.CounterMetricTypeName, zero)

		return err
	case errors.Is(err, nil):
		if metricType != metrics.CounterMetricTypeName {
			return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, metricType)
		}

		_, err := db.connection.ExecContext(ctx,
			"UPDATE metrics set metric_delta = $1 WHERE metric_id = $2",
			zero, metricName)

		return err
	default:
		return err
	}
}

func (db *DBStore) UpdateGaugeMetric(ctx context.Context, metricName string, metricData metrics.Gauge) error {
	var metricType string

	row := db.connection.QueryRowContext(ctx,
		"SELECT metric_type FROM metrics WHERE metric_id = $1",
		metricName)
	err := row.Scan(&metricType)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		_, err := db.connection.ExecContext(ctx,
			"INSERT INTO metrics (metric_id, metric_type, metric_value) VALUES ($1, $2, $3)",
			metricName, metrics.GaugeMetricTypeName, metricData)

		return err
	case errors.Is(err, nil):
		if metricType != metrics.GaugeMetricTypeName {
			return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, metricType)
		}

		_, err := db.connection.ExecContext(ctx,
			"UPDATE metrics set metric_value = $1 WHERE metric_id = $2",
			metricData, metricName)

		return err
	default:
		return err
	}
}

func (db *DBStore) GetMetric(ctx context.Context, metricName string) (*metrics.Metric, bool, error) {
	metric := metrics.Metric{}

	row := db.connection.QueryRowContext(ctx,
		"SELECT metric_id,metric_type,metric_delta,metric_value FROM metrics WHERE metric_id = $1",
		metricName)
	err := row.Scan(&metric.ID, &metric.MType, metric.Delta, metric.Value)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return nil, false, nil
	case errors.Is(err, nil):
		return &metric, true, nil
	default:
		log.Printf("Could't get mertic: %q", err)

		return nil, false, err
	}
}

func (db *DBStore) GetMetrics(ctx context.Context) (map[string]*metrics.Metric, error) {
	metricsCache := make(map[string]*metrics.Metric)

	rows, err := db.connection.QueryContext(ctx,
		"SELECT metric_id,metric_type,metric_delta,metric_value FROM $1", psqlMetricsTableName)

	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Printf("Couldn't close rows: %q", err)
		}
	}(rows)

	for rows.Next() {
		var metric metrics.Metric
		err = rows.Scan(&metric.ID, &metric.MType, metric.Delta, metric.Value)
		if err != nil {
			return nil, err
		}

		metricsCache[metric.ID] = &metric
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metricsCache, nil
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
