package pgstorage

import (
	"errors"
	"syscall"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sys/windows"
)

type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota

	Retriable
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return СlassifyPgError(pgErr)
	}

	var reqErr syscall.Errno
	if errors.As(err, &reqErr) {
		return classifySyscallError(reqErr)
	}

	return NonRetriable

}

func СlassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {

	switch pgErr.Code {
	// Класс 08 - Ошибки соединения
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable

	// Класс 40 - Откат транзакции
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return Retriable

	// Класс 57 - Ошибка оператора
	case pgerrcode.CannotConnectNow: // 57P03
		return Retriable
	}

	switch pgErr.Code {
	// Класс 22 - Ошибки данных
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable

	// Класс 23 - Нарушение ограничений целостности
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable

	// Класс 42 - Синтаксические ошибки
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}

	return NonRetriable
}

func classifySyscallError(reqErr syscall.Errno) PGErrorClassification {

	switch reqErr {
	//unix
	case syscall.ECONNREFUSED,
		syscall.ECONNRESET,
		syscall.ETIMEDOUT,
		syscall.EAGAIN:
		return Retriable

	//windows
	case windows.WSAECONNREFUSED,
		windows.WSAECONNRESET,
		windows.WSAETIMEDOUT:
		return Retriable

	}

	return NonRetriable
}
