package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/galogen13/yandex-go-metrics/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	maxAttepmts = 3
	firstDelay  = 1
)

type PGStorage struct {
	pool *pgxpool.Pool
}

func NewPGStorage(ctx context.Context, ps string) (*PGStorage, error) {

	config, err := pgxpool.ParseConfig(ps)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	config.MaxConns = 5
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := PGStorage{pool: pool}

	if err := storage.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &storage, nil
}

func (storage *PGStorage) Close() error {
	storage.pool.Close()
	return nil
}

func (storage *PGStorage) Ping(ctx context.Context) error {
	return storage.pool.Ping(ctx)
}

func (storage *PGStorage) Insert(ctx context.Context, metricsInsert []*metrics.Metric) error {
	err := storage.insertNoRetry(ctx, metricsInsert)
	if err == nil {
		return nil
	}

	classifier := NewPostgresErrorClassifier()
	for attempt := 0; attempt < maxAttepmts; attempt++ {
		classification := classifier.Classify(err)
		switch classification {
		case Retriable:
			delay := firstDelay + attempt*2
			logger.Log.Info("retryable error, Insert delayed",
				zap.Int("delay", delay),
				zap.Error(err),
			)
			time.Sleep(time.Duration(delay) * time.Second)
			err = storage.insertNoRetry(ctx, metricsInsert)
		case NonRetriable:
			return fmt.Errorf("non retriable error Insert: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("Insert aborted after %d attempts: err: %w", maxAttepmts, err)
	}

	return nil
}

func (storage *PGStorage) insertNoRetry(ctx context.Context, metricsInsert []*metrics.Metric) error {

	tx, err := storage.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction insert: %w", err)
	}

	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	for _, metric := range metricsInsert {
		batch.Queue(
			`INSERT INTO metrics(id, mtype, value, delta) VALUES ($1, $2, $3, $4)`,
			metric.ID, metric.MType, metric.Value, metric.Delta,
		)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for range metricsInsert {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch item: %w", err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil

}

func (storage *PGStorage) Update(ctx context.Context, metricsUpdate []*metrics.Metric) error {
	err := storage.updateNoRetry(ctx, metricsUpdate)
	if err == nil {
		return nil
	}

	classifier := NewPostgresErrorClassifier()
	for attempt := 0; attempt < maxAttepmts; attempt++ {
		classification := classifier.Classify(err)
		switch classification {
		case Retriable:
			delay := firstDelay + attempt*2
			logger.Log.Info("retryable error, Update delayed",
				zap.Int("delay", delay),
				zap.Error(err),
			)
			time.Sleep(time.Duration(delay) * time.Second)
			err = storage.updateNoRetry(ctx, metricsUpdate)
		case NonRetriable:
			return fmt.Errorf("non retriable error Update: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("Update aborted after %d attempts: err: %w", maxAttepmts, err)
	}

	return nil
}

func (storage *PGStorage) updateNoRetry(ctx context.Context, metricsUpdate []*metrics.Metric) error {

	tx, err := storage.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction insert: %w", err)
	}

	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}

	for _, metric := range metricsUpdate {
		batch.Queue(
			`UPDATE metrics SET value=$1, delta=$2 WHERE id=$3 AND mtype=$4;`,
			metric.Value, metric.Delta, metric.ID, metric.MType,
		)
	}

	results := tx.SendBatch(ctx, batch)
	defer results.Close()

	for range metricsUpdate {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to execute batch item: %w", err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("failed to close batch results: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil

}

func (storage *PGStorage) Get(ctx context.Context, metric *metrics.Metric) (bool, *metrics.Metric, error) {

	ok, qMetric, err := storage.getNoRetry(ctx, metric)
	if err == nil {
		return ok, qMetric, nil
	}

	classifier := NewPostgresErrorClassifier()
	for attempt := 0; attempt < maxAttepmts; attempt++ {
		classification := classifier.Classify(err)
		switch classification {
		case Retriable:
			delay := firstDelay + attempt*2
			logger.Log.Info("retryable error, Get delayed",
				zap.Int("delay", delay),
				zap.Error(err),
			)
			time.Sleep(time.Duration(delay) * time.Second)
			ok, qMetric, err = storage.getNoRetry(ctx, metric)
		case NonRetriable:
			return false, nil, fmt.Errorf("non retriable error Get: %w", err)
		}
	}

	if err != nil {
		return false, nil, fmt.Errorf("Get aborted after %d attempts: err: %w", maxAttepmts, err)
	}

	return ok, qMetric, nil
}

func (storage *PGStorage) getNoRetry(ctx context.Context, metric *metrics.Metric) (bool, *metrics.Metric, error) {

	var (
		value sql.NullFloat64
		delta sql.NullInt64
	)

	row := storage.pool.QueryRow(ctx, "SELECT id, mtype, value, delta FROM metrics WHERE id = $1 AND mtype = $2;", metric.ID, metric.MType)
	qMetric := metrics.Metric{}
	err := row.Scan(&qMetric.ID, &qMetric.MType, &value, &delta)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil
		}
		return false, nil, fmt.Errorf("failed to scan query result Get: %w", err)
	}

	if value.Valid {
		qMetric.Value = &value.Float64
	}

	if delta.Valid {
		qMetric.Delta = &delta.Int64
	}

	return true, &qMetric, nil
}

func (storage *PGStorage) GetByIDs(ctx context.Context, ids []string) (map[string]*metrics.Metric, error) {
	result, err := storage.getByIDsNoRetry(ctx, ids)
	if err == nil {
		return result, nil
	}

	classifier := NewPostgresErrorClassifier()
	for attempt := 0; attempt < maxAttepmts; attempt++ {
		classification := classifier.Classify(err)
		switch classification {
		case Retriable:
			delay := firstDelay + attempt*2
			logger.Log.Info("retryable error, GetByIDs delayed",
				zap.Int("delay", delay),
				zap.Error(err),
			)
			time.Sleep(time.Duration(delay) * time.Second)
			result, err = storage.getByIDsNoRetry(ctx, ids)
		case NonRetriable:
			return nil, fmt.Errorf("non retriable error GetByIDs: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("GetByIDs aborted after %d attempts: err: %w", maxAttepmts, err)
	}

	return result, nil
}

func (storage *PGStorage) getByIDsNoRetry(ctx context.Context, ids []string) (map[string]*metrics.Metric, error) {

	rows, err := storage.pool.Query(ctx, "SELECT id, mtype, value, delta FROM metrics WHERE id = ANY($1);", ids)
	if err != nil {
		return nil, fmt.Errorf("failed to do query GetByIDs: %w", err)
	}

	defer rows.Close()

	result := make(map[string]*metrics.Metric)

	for rows.Next() {
		var (
			value   sql.NullFloat64
			delta   sql.NullInt64
			qMetric metrics.Metric
		)

		err = rows.Scan(&qMetric.ID, &qMetric.MType, &value, &delta)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query result GetByIDs: %w", err)
		}

		if value.Valid {
			qMetric.Value = &value.Float64
		}

		if delta.Valid {
			qMetric.Delta = &delta.Int64
		}

		result[qMetric.ID] = &qMetric
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (storage *PGStorage) GetAll(ctx context.Context) ([]*metrics.Metric, error) {
	result, err := storage.getAllNoRetry(ctx)
	if err == nil {
		return result, nil
	}

	classifier := NewPostgresErrorClassifier()
	for attempt := 0; attempt < maxAttepmts; attempt++ {
		classification := classifier.Classify(err)
		switch classification {
		case Retriable:
			delay := firstDelay + attempt*2
			logger.Log.Info("retryable error, GetAll delayed",
				zap.Int("delay", delay),
				zap.Error(err),
			)
			time.Sleep(time.Duration(delay) * time.Second)
			result, err = storage.getAllNoRetry(ctx)
		case NonRetriable:
			return nil, fmt.Errorf("non retriable error GetAll: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("GetAll aborted after %d attempts: err: %w", maxAttepmts, err)
	}

	return result, nil
}

func (storage *PGStorage) getAllNoRetry(ctx context.Context) ([]*metrics.Metric, error) {

	result := []*metrics.Metric{}
	rows, err := storage.pool.Query(ctx, "SELECT id, mtype, value, delta FROM metrics;")
	if err != nil {
		return nil, fmt.Errorf("failed to do query GetAll: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			value   sql.NullFloat64
			delta   sql.NullInt64
			qMetric metrics.Metric
		)

		err = rows.Scan(&qMetric.ID, &qMetric.MType, &value, &delta)
		if err != nil {
			return nil, fmt.Errorf("failed to scan query result GetAll: %w", err)
		}

		if value.Valid {
			qMetric.Value = &value.Float64
		}

		if delta.Valid {
			qMetric.Delta = &delta.Int64
		}

		result = append(result, &qMetric)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (storage *PGStorage) runMigrations() error {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	sqlDB := stdlib.OpenDBFromPool(storage.pool)
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if errors.Is(err, migrate.ErrNoChange) {
		logger.Log.Info("migrations are already installed")
	} else {
		logger.Log.Info("migrations installed succesfully")
	}

	return nil
}
