//go:build infra

package postgres_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli-lab/watersystem-ml/internal/fixtures"
	"github.com/bruli-lab/watersystem-ml/internal/infra/postgres"
)

func TestModelHealthRepository(t *testing.T) {
	sqldb, err := sql.Open("postgres", "postgres://userdb:passdb@localhost:5432/watersystem_ml?sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, postgres.RunMigrations(sqldb))
	db := bun.NewDB(sqldb, pgdialect.New())
	tracer := noop.NewTracerProvider().Tracer("test")
	predictionRepo := postgres.NewPredictionLogRepository(db, tracer)
	repo := postgres.NewModelHealthRepository(db, tracer)

	zone := uuid.NewString()
	lastTrainingAt := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	saveValidated := func(createdAt time.Time, reachedTarget bool) {
		moistureAfter := 20.0
		if reachedTarget {
			moistureAfter = 5.0
		}
		pl := fixtures.PredictionLogBuilder{Zone: &zone, WateringExecuted: true, CreatedAt: &createdAt}.Build(t)
		pl.AddValidation(new(createdAt.Add(15*time.Minute)), &moistureAfter)
		require.NoError(t, predictionRepo.Save(t.Context(), &pl))
	}
	saveValidated(lastTrainingAt.Add(-2*time.Hour), false)
	saveValidated(lastTrainingAt.Add(-3*time.Hour), false)
	saveValidated(lastTrainingAt.Add(10*time.Minute), true)

	t.Run(`Given a ModelHealthRepository,
	when GetModelHealth is called with the last training date,
	then it only counts predictions created after it`, func(t *testing.T) {
		got, err := repo.GetModelHealth(t.Context(), zone, lastTrainingAt)
		require.NoError(t, err)
		require.Equal(t, ml.NewModelHealth(zone, 1, 0), got)
	})

	t.Run(`Given a ModelHealthRepository,
	when GetModelHealth is called without a training date,
	then it counts every prediction`, func(t *testing.T) {
		got, err := repo.GetModelHealth(t.Context(), zone, time.Time{})
		require.NoError(t, err)
		require.Equal(t, ml.NewModelHealth(zone, 1, 2), got)
	})
}
