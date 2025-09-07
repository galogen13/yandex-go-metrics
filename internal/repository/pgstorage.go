package storage

import (
	"context"
	"database/sql"

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

func (storage *PGStorage) Update(ctx context.Context, metric *metrics.Metric) {

}

func (storage *PGStorage) Get(ctx context.Context, metric *metrics.Metric) (bool, *metrics.Metric) {
	return false, nil
}

func (storage *PGStorage) GetByID(ctx context.Context, ID string) (bool, *metrics.Metric) {
	return false, nil
}

func (storage *PGStorage) GetAll(ctx context.Context) []*metrics.Metric {
	return nil
}

func (storage *PGStorage) RestoreFromFile(ctx context.Context, fileStoragePath string) error {
	return nil
}

func (storage *PGStorage) SaveToFile(ctx context.Context, fileStoragePath string) error {
	return nil
}
