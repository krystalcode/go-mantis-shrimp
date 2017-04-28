/**
 * Provides a redis storage adapter for storing and retrieving actions.
 */

package msActionStorage

import (
	// Utilities
	"encoding/json"
	"fmt"
	"strconv"

	// Redis.
	"github.com/mediocregopher/radix.v2/redis"

	// Internal dependencies.
	chat "github.com/krystalcode/go-mantis-shrimp/actions/chat"
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
)

/**
 * Redis storage provider.
 */

// Redis implements the Storage interface, allowing to use Redis as a Storage
// engine.
type Redis struct {
	dsn    string
	client *redis.Client
}

// Get implements Storage.Get(). It retrieves from Storage and returns the
// Action for the given ID.
func (storage Redis) Get(_id int) common.Action {
	if storage.client == nil {
		panic("The Redis client has not been initialized yet.")
	}

	key := redisKey(_id)

	r := storage.client.Cmd("GET", key)
	if r.Err != nil {
		panic(r.Err)
	}

	jsonAction, err := r.Bytes()
	// If an error happens here, it should be because there is no value for this
	// key. It could be the case that the data is corrupted or the wrong data is
	// stored, we should see how to handle this later.
	// @I Handle edge cases when deserializing json in Redis
	if err != nil {
		return nil
	}

	// @I Dynamically detect the Action type and convert json to struct
	//    accordingly
	action := chat.Action{}
	err = json.Unmarshal(jsonAction, &action)
	if err != nil {
		panic(err)
	}

	return action
}

// Set implements Storage.Set(). It stores the given Action object to the Redis
// Storage.
func (storage Redis) Set(action common.Action) int {
	// @I Consider using hashmaps instead of json values
	// @I Investigate risk of an Action overriding another due to race conditions
	//    when creating them

	if storage.client == nil {
		panic("The Redis client has not been initialized yet.")
	}

	jsonAction, err := json.Marshal(action)
	if err != nil {
		panic(err)
	}

	// Generate an ID, store the Action, and update the Actions index set.
	_id := storage.generateID()
	key := redisKey(_id)
	err = storage.client.Cmd("SET", key, jsonAction).Err
	if err != nil {
		panic(err)
	}
	err = storage.client.Cmd("ZADD", "actions", _id, key).Err
	if err != nil {
		panic(err)
	}

	return _id
}

// generateID generates an ID for a new Action by incrementing the last known
// Action ID.
func (storage Redis) generateID() int {
	// Get the last ID that exists on the Actions index set, so that we can generate
	// the next one.
	r, err := storage.client.Cmd("ZREVRANGE", "actions", 0, 0, "WITHSCORES").List()
	if err != nil {
		panic(err)
	}

	// If there are no actions yet, start with ID 1.
	if len(r) == 0 {
		return 1
	}

	_id, err := strconv.Atoi(r[1])
	if err != nil {
		panic(err)
	}

	return _id + 1
}

// NewRedisStorage implements the StorageFactory function type. It initiates a
// connection to the Redis database defined in the given configuration, and it
// returns the Storage engine object.
var NewRedisStorage = func(config map[string]string) (Storage, error) {
	dsn, ok := config["STORAGE_REDIS_DSN"]
	if !ok {
		err := fmt.Errorf(
			"the \"%s\" configuration option is required for the Redis storage",
			"STORAGE_REDIS_DSN",
		)
		return nil, err
	}

	client, err := redis.Dial("tcp", dsn)
	if err != nil {
		err := fmt.Errorf("failed to connect to Redis: %s", err.Error())
		return nil, err
	}

	storage := Redis{
		dsn:    dsn,
		client: client,
	}

	return storage, nil
}

/**
 * For internal use.
 */

// Generate a Redis key for the given Action ID.
func redisKey(_id int) string {
	return "action:" + strconv.Itoa(_id)
}
