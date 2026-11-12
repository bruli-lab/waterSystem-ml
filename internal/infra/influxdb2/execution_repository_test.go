//go:build integration

package influxdb2_test

import (
	"testing"

	"github.com/bruli-lab/watersystem-ml/internal/infra/influxdb2"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestExecutionRepository_GetLastExecution(t *testing.T) {
	t.Run(`Given an ExecutionRepository with two zones,
		when GetLastExecution method is called, then it returns the last execution of each zone`, func(t *testing.T) {
		token := ""
		repo := influxdb2.NewExecutionRepository(
			"http://localhost:8086", token, "home", "bonsai-data",
			[]string{"Bonsai big", "Bonsai small"},
			noop.NewTracerProvider().Tracer("test"),
		)
		got, err := repo.GetLastExecution(t.Context())
		require.NoError(t, err)
		require.Len(t, got, 2)
	})
}
