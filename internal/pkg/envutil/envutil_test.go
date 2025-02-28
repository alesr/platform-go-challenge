package envutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	testCases := []struct {
		name          string
		givenKey      string
		givenValue    string
		givenFallback string
		expect        string
		isEnvSet      bool
	}{
		{
			name:          "return envar when it exists",
			givenKey:      "FOO_KEY",
			givenValue:    "foo_value",
			givenFallback: "fallback_value",
			expect:        "foo_value",
			isEnvSet:      true,
		},
		{
			name:          "return fallback when envar doesn't exist",
			givenKey:      "BAR_KEY",
			givenValue:    "",
			givenFallback: "fallback_value",
			expect:        "fallback_value",
			isEnvSet:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isEnvSet {
				t.Setenv(tc.givenKey, tc.givenValue)
			}

			got := GetEnv(tc.givenKey, tc.givenFallback)
			assert.Equal(t, tc.expect, got)
		})
	}
}
