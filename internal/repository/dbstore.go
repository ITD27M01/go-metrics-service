package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	bindata "github.com/golang-migrate/migrate/v4/source/go_bindata"
	"github.com/itd27m01/go-metrics-service/internal/db/migrations"
	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver
)

const (
	psqlDriverName      = "pgx"
	psqlTimeout         = 1 * time.Second
	migrationSourceName = "go-bindata"
)

type DBStore struct {
	connection   *sql.DB
	context      context.Context
	syncChannel  chan struct{}
	metricsCache map[string]*metrics.Metric
	mu           sync.Mutex
}

func NewDBStore(ctx context.Context, databaseDSN string, syncChannel chan struct{}) (*DBStore, error) {
	var db DBStore

	conn, err := sql.Open(psqlDriverName, databaseDSN)
	if err != nil {
		return nil, err
	}

	metricsCache := make(map[string]*metrics.Metric)
	db = DBStore{
		connection:   conn,
		context:      ctx,
		syncChannel:  syncChannel,
		metricsCache: metricsCache,
	}

	if err := db.migrate(); err != nil {
		return nil, err
	}

	return &db, nil
}

func (db *DBStore) UpdateCounterMetric(metricName string, metricData metrics.Counter) error {
	db.mu.Lock()
	defer db.sync()
	defer db.mu.Unlock()

	currentMetric, ok := db.metricsCache[metricName]
	switch {
	case ok && currentMetric.Delta != nil:
		*(currentMetric.Delta) += metricData
	case ok && currentMetric.Delta == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		db.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.CounterMetricTypeName,
			Delta: &metricData,
		}
	}

	return nil
}

func (db *DBStore) ResetCounterMetric(metricName string) error {
	db.mu.Lock()
	defer db.sync()
	defer db.mu.Unlock()

	var zero metrics.Counter
	currentMetric, ok := db.metricsCache[metricName]
	switch {
	case ok && currentMetric.Delta != nil:
		*(currentMetric.Delta) = zero
	case ok && currentMetric.Delta == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		db.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.CounterMetricTypeName,
			Delta: &zero,
		}
	}

	return nil
}

func (db *DBStore) UpdateGaugeMetric(metricName string, metricData metrics.Gauge) error {
	db.mu.Lock()
	defer db.sync()
	defer db.mu.Unlock()

	currentMetric, ok := db.metricsCache[metricName]
	switch {
	case ok && currentMetric.Value != nil:
		*(currentMetric.Value) = metricData
	case ok && currentMetric.Value == nil:
		return fmt.Errorf("%w %s:%s", ErrMetricTypeMismatch, metricName, currentMetric.MType)
	default:
		db.metricsCache[metricName] = &metrics.Metric{
			ID:    metricName,
			MType: metrics.GaugeMetricTypeName,
			Value: &metricData,
		}
	}

	return nil
}

func (db *DBStore) GetMetric(metricName string) (*metrics.Metric, bool) {
	metric, ok := db.metricsCache[metricName]

	return metric, ok
}

func (db *DBStore) GetMetrics() map[string]*metrics.Metric {
	return db.metricsCache
}

func (db *DBStore) Ping() error {
	ctx, cancel := context.WithTimeout(db.context, psqlTimeout)
	defer cancel()

	return db.connection.PingContext(ctx)
}

func (db *DBStore) SaveMetrics() error { return nil }

func (db *DBStore) LoadMetrics() error { return nil }

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

	return migration.Up()
}

func (db *DBStore) sync() {
	db.syncChannel <- struct{}{}
}
