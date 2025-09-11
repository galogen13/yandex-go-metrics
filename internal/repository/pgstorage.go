package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/galogen13/yandex-go-metrics/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PGStorage struct {
	db *sql.DB
}

func NewPGStorage(ps string) (*PGStorage, error) {

	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}

	storage := PGStorage{db: db}

	if err := storage.Ping(context.Background()); err != nil {
		return nil, err
	}

	if err := storage.runMigrations(); err != nil {
		return nil, err
	}

	return &storage, nil
}

func (storage *PGStorage) Close() error {
	return storage.db.Close()
}

func (storage *PGStorage) Ping(ctx context.Context) error {
	return storage.db.PingContext(ctx)
}

func (storage *PGStorage) Insert(ctx context.Context, metricsInsert []*metrics.Metric) error {

	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO metrics(id, mtype, value, delta) VALUES ($1, $2, $3, $4)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, metric := range metricsInsert {
		_, err := stmt.ExecContext(ctx, metric.ID, metric.MType, metric.Value, metric.Delta)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}

func (storage *PGStorage) Update(ctx context.Context, metricsUpdate []*metrics.Metric) error {

	tx, err := storage.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`UPDATE metrics SET value=$1, delta=$2 WHERE id=$3 AND mtype=$4;`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, metric := range metricsUpdate {
		result, err := stmt.ExecContext(ctx, metric.Value, metric.Delta, metric.ID, metric.MType)
		if err != nil {
			return err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if affected == 0 {
			return metrics.ErrMetricNotFound
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil

}

func (storage *PGStorage) Get(ctx context.Context, metric *metrics.Metric) (bool, *metrics.Metric, error) {

	var (
		value sql.NullFloat64
		delta sql.NullInt64
	)

	row := storage.db.QueryRowContext(ctx, "SELECT id, mtype, value, delta FROM metrics WHERE id = $1 AND mtype = $2;", metric.ID, metric.MType)
	qMetric := metrics.Metric{}
	err := row.Scan(&qMetric.ID, &qMetric.MType, &value, &delta)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil, nil
		}
		return false, nil, err
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

	result := make(map[string]*metrics.Metric)
	rows, err := storage.db.QueryContext(ctx, "SELECT id, mtype, value, delta FROM metrics WHERE id = ANY($1);", ids)
	if err != nil {
		return nil, err
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
			return nil, err
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

	result := []*metrics.Metric{}
	rows, err := storage.db.QueryContext(ctx, "SELECT id, mtype, value, delta FROM metrics;")
	if err != nil {
		return nil, err
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
			return nil, err
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

	driver, err := postgres.WithInstance(storage.db, &postgres.Config{})
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
