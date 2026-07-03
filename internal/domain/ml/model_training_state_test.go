package ml_test

import (
	"testing"
	"time"

	"github.com/bruli/watersystem-ml/internal/domain/ml"
	"github.com/stretchr/testify/require"
)

func TestModelTrainingState_IsRecentlyTraining(t *testing.T) {
	type fields struct {
		zone       string
		trainingAt time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "and training time is recent, then it returns true",
			fields: fields{
				zone:       "zone",
				trainingAt: time.Now().Add(-1 * time.Hour),
			},
			want: true,
		},
		{
			name: "and training time is not recent, then it returns false",
			fields: fields{
				zone:       "zone",
				trainingAt: time.Now().Add(-50 * time.Hour),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(`Given a ModelTrainingState struct,
		when IsRecentlyTraining method is called `+tt.name, func(t *testing.T) {
			t.Parallel()
			m := ml.NewModelTrainingState(tt.fields.zone, tt.fields.trainingAt)
			require.Equal(t, tt.want, m.IsRecentlyTraining())
		})
	}
}
