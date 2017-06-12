/**
 * Tests for the Redis storage engine of the msWatchStorage module.
 */

package msWatchStorage

import (
	// Testing packages.
	"github.com/stretchr/testify/assert"
	"testing"
)

/**
 * Tests.
 */

func TestNewRedisStorage_MissingDSN(t *testing.T) {
	config := make(map[string]interface{})
	_, err := Create(config)
	assert.NotNil(t, err)
}

/**
 * Tests for functions/types for internal use.
 */

func TestRedisKey(t *testing.T) {
	sIDResult := redisKey(1)
	sIDDesired := "watch:1"
	assert.Equal(t, sIDDesired, sIDResult)
}
