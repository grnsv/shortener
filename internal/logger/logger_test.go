package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	envs := []string{"testing", "production", "development"}
	for _, env := range envs {
		t.Run(env, func(t *testing.T) {
			logger, err := New(env)
			assert.NoError(t, err)
			assert.NotNil(t, logger)

			err = logger.Sync()
			if err != nil && err.Error() == "sync /dev/stderr: invalid argument" {
				t.Skipf("Skipping test due to known issue: %v", err)
			}
			assert.NoError(t, err)
		})
	}
}
