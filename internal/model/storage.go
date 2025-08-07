package models

type Storage interface {
	Update(ID string, MType string, Value any) error
}
