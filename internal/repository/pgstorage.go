package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
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

	return &PGStorage{db: db}, nil
}

func (storage *PGStorage) Close() error {
	return storage.db.Close()
}
func (storage *PGStorage) Ping(ctx context.Context) error {

	return storage.db.PingContext(ctx)

}

func (storage *PGStorage) Insert(ctx context.Context, metric *metrics.Metric) error {

	result, err := storage.db.Exec(`INSERT INTO metrics(id, mtype, value, delta) VALUES ($1, $2, $3, $4)`, metric.ID, metric.MType, metric.Value, metric.Delta)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("metric not found")
	}
	return nil

}

func (storage *PGStorage) Update(ctx context.Context, metric *metrics.Metric) error {

	result, err := storage.db.Exec(`UPDATE metrics
	SET value=$1, delta=$2 WHERE id=$3 AND mtype=$4;`, metric.Value, metric.Delta, metric.ID, metric.MType)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("metric not found")
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

func (storage *PGStorage) GetByID(ctx context.Context, ID string) (bool, *metrics.Metric, error) {

	var (
		value sql.NullFloat64
		delta sql.NullInt64
	)

	row := storage.db.QueryRowContext(ctx, "SELECT id, mtype, value, delta FROM metrics WHERE id = $1;", ID)
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

func (storage *PGStorage) GetAll(ctx context.Context) ([]*metrics.Metric, error) {

	var (
		value sql.NullFloat64
		delta sql.NullInt64
	)

	result := []*metrics.Metric{}
	rows, err := storage.db.QueryContext(ctx, "SELECT id, mtype, value, delta FROM metrics;")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var qMetric metrics.Metric
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

func (storage *PGStorage) RestoreFromFile(ctx context.Context, fileStoragePath string) error {
	//Заглушка, не актуально для БД
	return nil
}

func (storage *PGStorage) SaveToFile(ctx context.Context, fileStoragePath string) error {
	//Заглушка, не актуально для БД
	return nil
}
