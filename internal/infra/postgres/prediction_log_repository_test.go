//go:build infra

package postgres_test

import (
	"database/sql"
	"testing"

	"github.com/bruli/watersystem-ml/internal/fixtures"
	"github.com/bruli/watersystem-ml/internal/infra/postgres"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestPredictionLogRepository(t *testing.T) {
	t.Run(`Given a PredictionLogRepository`, func(t *testing.T) {
		sqldb, err := sql.Open("postgres", "postgres://userdb:passdb@localhost:5432/watersystem_ml?sslmode=disable")
		require.NoError(t, err)
		db := bun.NewDB(sqldb, pgdialect.New())
		repo := postgres.NewPredictionLogRepository(db, noop.NewTracerProvider().Tracer("test"))
		t.Run(`when Save method is called,
		then it insert data and returns nil error`, func(t *testing.T) {
			err = repo.Save(t.Context(), new(fixtures.PredictionLogBuilder{}.Build(t)))
			require.NoError(t, err)
		})
	})
}
