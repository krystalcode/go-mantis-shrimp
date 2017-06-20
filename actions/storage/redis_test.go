/**
 * Tests for the Redis Action Storage API.
 */

package msActionStorage

import (
	// Testing packages.
	"github.com/stretchr/testify/assert"
	"testing"

	// Redis.
	"github.com/mediocregopher/radix.v2/redis"

	// Utilities.
	"fmt"
	"reflect"

	// Internal dependencies.
	chat "github.com/krystalcode/go-mantis-shrimp/actions/chat"
)

/**
 * Tests.
 */

func TestRedisKey(t *testing.T) {
	iID := 1
	sDesired := "action:1"
	sResult := redisKey(iID)

	assert.Equal(t, sDesired, sResult)
}

func TestGet_NoClient(t *testing.T) {
	storage := Redis{}
	id := 1
	_, err := storage.Get(id)
	assert.NotNil(t, err)
}

func TestGet_RedisError(t *testing.T) {
	client := &TestRedisClient_ErrorResponse{}
	storage := Redis{
		client: client,
	}
	id := 1
	_, err := storage.Get(id)
	assert.NotNil(t, err)
}

func TestGet_EmptyResponse(t *testing.T) {
	client := &TestRedisClient_EmptyResponse{}
	storage := Redis{
		client: client,
	}
	id := 1
	_, err := storage.Get(id)
	assert.NotNil(t, err)
}

func TestGet_JSONError(t *testing.T) {
	client := &TestRedisClient_WrongValueResponse{}
	storage := Redis{
		client: client,
	}
	id := 1
	_, err := storage.Get(id)
	assert.NotNil(t, err)
}

func TestGet_Success(t *testing.T) {
	// Create the expected Action object that matches the JSON returned by the
	// Redis Client.
	id := 1
	messageText := "Chat message text"
	message := chat.Message{
		Text: &messageText,
	}
	pExpectedAction := chat.NewAction(
		"Action name",
		"Chat webhook",
		message,
	)
	oExpectedAction := *pExpectedAction

	// Make a stub request to get an Action from Redis.
	client := &TestRedisClient_RightValueResponse{}
	storage := Redis{
		client: client,
	}
	pAction, err := storage.Get(id)
	assert.Nil(t, err)

	// We'll be checking if the JSON data is properly converted to an Action
	// object i.e. if the object is of the right type and if the object's struct
	// fields have the expected values.
	oActualAction := *pAction
	oExpectedActionType := reflect.TypeOf(oExpectedAction)
	oActualActionType := reflect.TypeOf(oActualAction)
	assert.Equal(t, oExpectedActionType.PkgPath(), oActualActionType.PkgPath())
	assert.Equal(t, oExpectedActionType.Name(), oActualActionType.Name())

	oExpectedActionValue := reflect.ValueOf(oExpectedAction)
	oActualActionValue := reflect.ValueOf(oActualAction)

	// Check the Name field.
	assert.Equal(
		t,
		oExpectedActionValue.FieldByName("Name").String(),
		oActualActionValue.FieldByName("Name").String(),
	)

	// Check the URL field.
	assert.Equal(
		t,
		oExpectedActionValue.FieldByName("URL").String(),
		oActualActionValue.FieldByName("URL").String(),
	)

	// Check the Message, only the Text field was set.
	vExpectedActionMessage := oExpectedActionValue.FieldByName("Message")
	iExpectedActionMessage := vExpectedActionMessage.Interface()
	oExpectedActionMessage := iExpectedActionMessage.(chat.Message)
	vActualActionMessage := oActualActionValue.FieldByName("Message")
	iActualActionMessage := vActualActionMessage.Interface()
	oActualActionMessage := iActualActionMessage.(chat.Message)

	assert.Equal(
		t,
		*oExpectedActionMessage.Text,
		*oActualActionMessage.Text,
	)
}

/**
 * Functions/types for internal use.
 */

type TestRedisClient_EmptyResponse struct{}

func (c *TestRedisClient_EmptyResponse) Cmd(cmd string, args ...interface{}) *redis.Resp {
	return &redis.Resp{}
}

type TestRedisClient_ErrorResponse struct{}

func (c *TestRedisClient_ErrorResponse) Cmd(cmd string, args ...interface{}) *redis.Resp {
	err := fmt.Errorf("an error has occurred while executing the Redis command")
	return &redis.Resp{
		Err: err,
	}
}

type TestRedisClient_WrongValueResponse struct{}

func (c *TestRedisClient_WrongValueResponse) Cmd(cmd string, args ...interface{}) *redis.Resp {
	return redis.NewResp("{}")
}

type TestRedisClient_RightValueResponse struct{}

func (c *TestRedisClient_RightValueResponse) Cmd(cmd string, args ...interface{}) *redis.Resp {
	return redis.NewResp("{\"type\":\"chat_message\",\"action\":{\"name\":\"Action name\",\"url\":\"Chat webhook\",\"message\":{\"text\":\"Chat message text\"}}}")
}
