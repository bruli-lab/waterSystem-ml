//go:build infra

package postgres_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/bruli/watersystem-ml/internal/domain/ml"
	"github.com/bruli/watersystem-ml/internal/infra/postgres"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestModelTrainingStateRepository(t *testing.T) {
	sqldb, err := sql.Open("postgres", "postgres://userdb:passdb@localhost:5432/watersystem_ml?sslmode=disable")
	require.NoError(t, err)
	zone := "Bonsai big"
	db := bun.NewDB(sqldb, pgdialect.New())
	repo := postgres.NewModelTrainingStateRepository(db, noop.NewTracerProvider().Tracer("test"))
	t.Run(`Given a ModelTrainingStateRepository `, func(t *testing.T) {
		t.Run(`when GetModelTrainingState is called and does not exists values,
		then it returns a not found error`, func(t *testing.T) {
			got, err := repo.GetModelTrainingState(t.Context(), zone)
			require.Nil(t, got)
			require.ErrorIs(t, err, ml.ErrModelPredictionStateNotFound)
		})
		t.Run(`when Save is called,
		then it returns no error`, func(t *testing.T) {
			mps := ml.NewModelTrainingState(zone, time.Now())
			err := repo.Save(t.Context(), mps)
			require.NoError(t, err)
		})
		t.Run(`when GetModelTrainingState is called and exists values,
		then it returns values`, func(t *testing.T) {
			got, err := repo.GetModelTrainingState(t.Context(), zone)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	})
}
