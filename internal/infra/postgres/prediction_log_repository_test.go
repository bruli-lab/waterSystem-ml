//go:build infra

package postgres_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/bruli-lab/watersystem-ml/internal/fixtures"
	"github.com/bruli-lab/watersystem-ml/internal/infra/postgres"
	"github.com/google/uuid"
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
		zone := uuid.NewString()
		createdAt := time.Now()
		repo := postgres.NewPredictionLogRepository(db, noop.NewTracerProvider().Tracer("test"))
		t.Run(`when Save method is called,
		then it insert data and returns nil error`, func(t *testing.T) {
			err = repo.Save(t.Context(), new(fixtures.PredictionLogBuilder{Zone: new(zone), CreatedAt: &createdAt}.Build(t)))
			require.NoError(t, err)
		})
		t.Run(`when GetPendingValidationZones method is called,
		then it should return true`, func(t *testing.T) {
			got, err := repo.GetPendingValidationZones(t.Context())
			require.NoError(t, err)
			require.True(t, got[zone])
		})
		t.Run(`when GetPendingByZone method is called `, func(t *testing.T) {
			t.Run(`and does not exists, 
			then it should return a prediction log not found error`, func(t *testing.T) {
				got, err := repo.GetPendingByZone(t.Context(), zone, createdAt.Add(time.Minute))
				require.Nil(t, got)
				require.ErrorIs(t, err, ml.ErrPredictionLogNotFound)
			})
			t.Run(`and exists, 
			then it should return a prediction log`, func(t *testing.T) {
				got, err := repo.GetPendingByZone(t.Context(), zone, createdAt.Add(16*time.Minute))
				require.NoError(t, err)
				require.NotNil(t, got)
			})
		})
	})
}

func TestPredictionLogRepository_ExpireStale(t *testing.T) {
	sqldb, err := sql.Open("postgres", "postgres://userdb:passdb@localhost:5432/watersystem_ml?sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, postgres.RunMigrations(sqldb))
	db := bun.NewDB(sqldb, pgdialect.New())
	repo := postgres.NewPredictionLogRepository(db, noop.NewTracerProvider().Tracer("test"))

	zone := uuid.NewString()
	createdAt := time.Now().Add(-3 * time.Hour).UTC().Truncate(time.Second)
	require.NoError(t, repo.Save(t.Context(), new(fixtures.PredictionLogBuilder{Zone: &zone, CreatedAt: &createdAt}.Build(t))))

	t.Run(`Given a pending prediction older than the limit,
	when GetPendingBefore is called,
	then it returns it`, func(t *testing.T) {
		stale, err := repo.GetPendingBefore(t.Context(), time.Now().Add(-ml.MaxPendingValidationAge))
		require.NoError(t, err)
		var found *ml.PredictionLog
		for _, pl := range stale {
			if pl.Zone() == zone {
				found = pl
			}
		}
		require.NotNil(t, found)
		require.Equal(t, createdAt, found.CreatedAt().UTC())

		found.Expire(time.Now())
		require.NoError(t, repo.Save(t.Context(), found))
	})

	t.Run(`Given an expired prediction,
	when GetPendingValidationZones is called,
	then the zone is no longer pending`, func(t *testing.T) {
		got, err := repo.GetPendingValidationZones(t.Context())
		require.NoError(t, err)
		require.False(t, got[zone])
	})
}
