/**
 * Tests for the msWatchStorage module.
 */

package msWatchStorage

import (
	// Testing packages.
	"github.com/stretchr/testify/assert"
	"testing"
)

var NewTestStorage = func(config map[string]interface{}) (Storage, error) {
	return nil, nil
}

/**
 * Tests.
 */

func TestCreate_Success(t *testing.T) {
	storageFactories["test"] = NewTestStorage

	config := make(map[string]interface{})
	config["type"] = "test"
	_, err := Create(config)
	assert.Nil(t, err)
}

func TestCreate_WrongConfig_MissingType(t *testing.T) {
	config := make(map[string]interface{})
	_, err := Create(config)
	assert.NotNil(t, err)
}

func TestCreate_WrongConfig_EmptyType(t *testing.T) {
	config := make(map[string]interface{})
	config["type"] = ""
	_, err := Create(config)
	assert.NotNil(t, err)
}

func TestCreate_WrongConfig_WrongType(t *testing.T) {
	config := make(map[string]interface{})
	config["type"] = "mysql"
	_, err := Create(config)
	assert.NotNil(t, err)
}
