package postgres

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"strings"
	"time"
)

func toNonEmptyTrimmedPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func postgresTextToStrPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	v := t.String
	return &v
}

func strPrtToAny(s *string) any {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return v
}

func toPostgresUUIDPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil || *id == uuid.Nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toPostgresTextPtr(str *string) pgtype.Text {
	if str == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *str, Valid: true}
}

func toPostgresFloat8Ptr(v *float64) pgtype.Float8 {
	if v == nil {
		return pgtype.Float8{Valid: false}
	}
	return pgtype.Float8{Float64: *v, Valid: true}
}

func toPostgresDatePtr(v *time.Time) pgtype.Date {
	if v == nil {
		return pgtype.Date{Valid: false}
	}
	d := time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, time.UTC)
	return pgtype.Date{Time: d, Valid: true}
}
