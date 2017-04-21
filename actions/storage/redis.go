/**
 * Provides a redis storage adapter for storing and retrieving actions.
 */

package msActionStorage

import (
	// Utilities
	"encoding/json"
	"errors"
	"fmt"

	// Redis.
  "github.com/mediocregopher/radix.v2/redis"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
	chat   "github.com/krystalcode/go-mantis-shrimp/actions/chat"
)

/**
 * Redis storage provider.
 */

// Implements the Storage interface.
type Redis struct {
	dsn     string
	client *redis.Client
}

// Get an Action by its ID.
func (storage Redis) Get(_id string) common.Action {
	if storage.client == nil {
		panic("The Redis client has not been initialized yet.")
	}

	jsonAction, err := storage.client.Cmd("GET", _id).Bytes()
	if err != nil {
		panic(err)
	}

	// @I Dynamically detect the Action type and convert json to struct
	//    accordingly
	action := chat.Action{}
	json.Unmarshal(jsonAction, &action)

	return action
}

// Set an Action.
// @I Consider automatically generate the ID and return it
func (storage Redis) Set(action common.Action) {
	if storage.client == nil {
		panic("The Redis client has not been initialized yet.")
	}

	jsonAction, err := json.Marshal(action)
	if err != nil {
		panic(err)
	}

	err = storage.client.Cmd("SET", action.GetName(), jsonAction).Err
	if err != nil {
		panic(err)
	}
}

// Implements the StorageFactory interface.
// Connects to the Redis database defined in the given configuration and returns
// the storage.
func NewRedisStorage(config map[string]string) (Storage, error) {
	dsn, ok := config["STORAGE_REDIS_DSN"]
	if !ok {
		err := errors.New(
			fmt.Sprintf(
				"The \"%s\" configuration option is required for the Redis storage",
				"STORAGE_REDIS_DSN",
			),
		)
		return nil, err
	}

	client, err := redis.Dial("tcp", dsn)
	if err != nil {
		err := errors.New(fmt.Sprintf("Failed to connect to Redis: %s", err.Error()))
		return nil, err
	}

	storage := Redis {
		dsn    : dsn,
		client : client,
	}

	return storage, nil
}
