package pgstorage

import (
	"errors"
	"syscall"

	"github.com/galogen13/yandex-go-metrics/internal/classification"
	"github.com/galogen13/yandex-go-metrics/internal/retry"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) retry.ErrorClassification {
	if err == nil {
		return retry.NonRetriable
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return СlassifyPgError(pgErr)
	}

	var reqErr syscall.Errno
	if errors.As(err, &reqErr) {
		return classifySyscallError(reqErr)
	}

	return retry.NonRetriable

}

func СlassifyPgError(pgErr *pgconn.PgError) retry.ErrorClassification {

	switch pgErr.Code {
	// Класс 08 - Ошибки соединения
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return retry.Retriable

	// Класс 40 - Откат транзакции
	case pgerrcode.TransactionRollback, // 40000
		pgerrcode.SerializationFailure, // 40001
		pgerrcode.DeadlockDetected:     // 40P01
		return retry.Retriable

	// Класс 57 - Ошибка оператора
	case pgerrcode.CannotConnectNow: // 57P03
		return retry.Retriable
	}

	switch pgErr.Code {
	// Класс 22 - Ошибки данных
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return retry.NonRetriable

	// Класс 23 - Нарушение ограничений целостности
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return retry.NonRetriable

	// Класс 42 - Синтаксические ошибки
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return retry.NonRetriable
	}

	return retry.NonRetriable
}

func classifySyscallError(reqErr syscall.Errno) retry.ErrorClassification {

	if classification.IsRetriableSyscallError(reqErr) {
		return retry.Retriable
	}

	return retry.NonRetriable
}
