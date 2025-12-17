package postgres

import (
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func toPostgresUUID(id uuid.UUID) pgtype.UUID {
	if id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: [16]byte(id), Valid: true}
}

func toPostgresText(str string) pgtype.Text {
	str = strings.TrimSpace(str)
	if str == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: str, Valid: true}
}
