package models

type Storage interface {
	UpdateGauge(ID string, Value float64) error
	UpdateCounter(ID string, Delta int64) error
}
