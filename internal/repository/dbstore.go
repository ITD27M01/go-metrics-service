package repository

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/itd27m01/go-metrics-service/internal/pkg/metrics"
	_ "github.com/jackc/pgx/v4/stdlib" // init postgresql driver
)

const (
	psqlDriverName = "pgx"
	psqlTimeout    = 1 * time.Second
)

type DBStore struct {
	connection   *sql.DB
	context      context.Context
	syncChannel  chan struct{}
	metricsCache map[string]*metrics.Metric
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

	return &db, nil
}

func (db *DBStore) UpdateCounterMetric(name string, value metrics.Counter) error { return nil }
func (db *DBStore) ResetCounterMetric(name string) error                         { return nil }
func (db *DBStore) UpdateGaugeMetric(name string, value metrics.Gauge) error     { return nil }

func (db *DBStore) GetMetric(name string) (*metrics.Metric, bool) { return &metrics.Metric{}, true }
func (db *DBStore) GetMetrics() map[string]*metrics.Metric        { return nil }

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
