package watersystem

import (
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/bruli-lab/watersystem-ml/internal/domain/watering"
)

func unreachableExecutor() *Executor {
	return &Executor{
		cl:     &http.Client{Timeout: time.Second},
		tracer: noop.NewTracerProvider().Tracer("test"),
		host:   "http://127.0.0.1",
		port:   "1",
		zones:  map[string]string{"bonsai big": "1"},
		logger: slog.New(slog.DiscardHandler),
	}
}

func TestExecutor_UnreachableServerReturnsError(t *testing.T) {
	t.Parallel()

	t.Run(`Given an Executor with an unreachable server,
		when GetStatus method is called, then it returns an error`, func(t *testing.T) {
		t.Parallel()
		_, err := unreachableExecutor().GetStatus(t.Context())
		require.Error(t, err)
	})

	t.Run(`Given an Executor with an unreachable server,
		when Execute method is called, then it returns an error`, func(t *testing.T) {
		t.Parallel()
		err := unreachableExecutor().Execute(t.Context(), watering.New("bonsai big", 20))
		require.Error(t, err)
	})
}
