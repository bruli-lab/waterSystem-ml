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
	e := unreachableExecutor()

	_, err := e.GetStatus(t.Context())
	require.Error(t, err)

	err = e.Execute(t.Context(), watering.New("bonsai big", 20))
	require.Error(t, err)
}
