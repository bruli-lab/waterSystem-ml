package ml_test

import (
	"testing"

	"github.com/bruli-lab/watersystem-ml/internal/domain/ml"
	"github.com/stretchr/testify/require"
)

func TestModelHealth_Check(t *testing.T) {
	type fields struct {
		zone                  string
		successfulPredictions int
		failedPredictions     int
	}
	tests := []struct {
		name          string
		fields        fields
		expectedEvent bool
	}{
		{
			name: "and failure rate is greater than max failure rate, then it returns event",
			fields: fields{
				zone:                  bonsaiSmallZone,
				successfulPredictions: 10,
				failedPredictions:     10,
			},
			expectedEvent: true,
		},
		{
			name: "and failure rate is lower than max failure rate, then it returns no event",
			fields: fields{
				zone:                  bonsaiSmallZone,
				successfulPredictions: 19,
				failedPredictions:     1,
			},
			expectedEvent: false,
		},
		{
			name: "with not complete number of predictions, then it returns no event",
			fields: fields{
				zone:                  bonsaiSmallZone,
				successfulPredictions: 1,
				failedPredictions:     1,
			},
			expectedEvent: false,
		},
	}
	for _, tt := range tests {
		t.Run(`Given a ModelHeath struct,
		when Check method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			mh := ml.NewModelHealth(tt.fields.zone, tt.fields.successfulPredictions, tt.fields.failedPredictions)
			mh.Check()
			events := mh.Events()
			if tt.expectedEvent {
				require.Len(t, events, 1)
				return
			}
			require.Len(t, events, 0)
		})
	}
}
