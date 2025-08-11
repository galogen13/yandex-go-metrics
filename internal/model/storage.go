package models

type Storage interface {
	Update(ID string, MType string, Value any) error
	Get(ID string, MType string) (Metrics, error)
	GetAll() []Metrics
}
