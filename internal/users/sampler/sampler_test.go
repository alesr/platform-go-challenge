package sampler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSampleUsers(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		amount int
	}{
		{
			name:   "Sample 0 users",
			amount: 0,
		},
		{
			name:   "Sample 10 users",
			amount: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := SampleUsers(tc.amount)
			require.NoError(t, err)

			assert.Equal(t, tc.amount, len(got))

		})
	}
}
