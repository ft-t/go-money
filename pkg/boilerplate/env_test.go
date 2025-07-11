package boilerplate_test

import (
	"github.com/ft-t/go-money/pkg/boilerplate"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetCurrentEnv(t *testing.T) {
	testCases := []struct {
		val      string
		expected boilerplate.Environment
	}{
		{"dev", boilerplate.Dev},
		{"ci", boilerplate.Ci},
		{"prod", boilerplate.Prod},
	}

	for _, tc := range testCases {
		t.Run(tc.val, func(t *testing.T) {
			t.Setenv("ENVIRONMENT", tc.val)
			assert.EqualValues(t, tc.expected, boilerplate.GetCurrentEnvironment())
		})
	}
}
