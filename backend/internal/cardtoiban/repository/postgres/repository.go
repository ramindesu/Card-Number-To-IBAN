package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"

	cardrepository "github.com/ramindesu/Card-Number-To-IBAN/backend/internal/cardtoiban/repository"
)

// CardToIBANDBTX is implemented by pgxpool.Pool and pgx.Tx.
type CardToIBANDBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

// CardToIBANRepository stores Card-to-IBAN audit records in PostgreSQL.
type CardToIBANRepository struct {
	database CardToIBANDBTX
}

// NewCardToIBANRepository creates a PostgreSQL-backed repository.
func NewCardToIBANRepository(database CardToIBANDBTX) *CardToIBANRepository {
	return &CardToIBANRepository{database: database}
}

var _ cardrepository.CardToIBANRepository = (*CardToIBANRepository)(nil)
