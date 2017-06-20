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
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
	wrapper "github.com/krystalcode/go-mantis-shrimp/actions/wrapper"
)

/**
 * Redis storage provider.
 */

// RedisClient is an interface that is used to allow dependency injection of the
// Redis client that makes the requests to the Redis datastore. Dependency
// injection is necessary for testing purposes.
type RedisClient interface {
	Cmd(string, ...interface{}) *redis.Resp
}

// Redis implements the Storage interface, allowing to use Redis as a Storage
// engine.
type Redis struct {
	dsn    string
	client RedisClient
}

// Get implements Storage.Get(). It retrieves from Storage and returns the
// Action for the given ID.
func (storage Redis) Get(id int) (*common.Action, error) {
	if storage.client == nil {
		return nil, fmt.Errorf("the Redis client has not been initialized yet")
	}

	key := redisKey(id)

	r := storage.client.Cmd("GET", key)
	if r.Err != nil {
		return nil, r.Err
	}

	jsonAction, err := r.Bytes()
	// If an error happens here, it should be because there is no value for this
	// key. It could be the case that the data is corrupted or the wrong data is
	// stored, we should see how to handle this later.
	// @I Handle edge cases when deserializing json in Redis
	if err != nil {
		return nil, err
	}

	// Create and initialize an Action object based on the given JSON object.
	action, err := wrapper.Create(jsonAction)
	if err != nil {
		return nil, err
	}

	return &action, nil
}

// Set implements Storage.Set(). It stores the given Action object to the Redis
// Storage.
func (storage Redis) Set(action common.Action) (*int, error) {
	// @I Consider using hashmaps instead of json values
	// @I Investigate risk of an Action overriding another due to race conditions
	//    when creating them

	if storage.client == nil {
		return nil, fmt.Errorf("the Redis client has not been initialized yet")
	}

	// We'll be storing an ActionWrapper which contains the Action type as well.
	wrapper, err := wrapper.Wrapper(action)
	if err != nil {
		return nil, err
	}
	jsonAction, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}

	// Generate an ID, store the Action, and update the Actions index set.
	id := storage.generateID()
	key := redisKey(id)
	err = storage.client.Cmd("SET", key, jsonAction).Err
	if err != nil {
		return nil, err
	}
	err = storage.client.Cmd("ZADD", "actions", id, key).Err
	if err != nil {
		return nil, err
	}

	return &id, nil
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

	id, err := strconv.Atoi(r[1])
	if err != nil {
		panic(err)
	}

	return id + 1
}

// NewRedisStorage implements the StorageFactory function type. It initiates a
// connection to the Redis database defined in the given configuration, and it
// returns the Storage engine object.
var NewRedisStorage = func(config map[string]interface{}) (Storage, error) {
	dsn, ok := config["dsn"]
	if !ok {
		err := fmt.Errorf("the DSN configuration option is required for the Redis storage")
		return nil, err
	}

	sDSN := dsn.(string)

	client, err := redis.Dial("tcp", sDSN)
	if err != nil {
		err := fmt.Errorf("failed to connect to Redis: %s", err.Error())
		return nil, err
	}

	storage := Redis{
		dsn:    sDSN,
		client: client,
	}

	return storage, nil
}

/**
 * For internal use.
 */

// Generate a Redis key for the given Action ID.
func redisKey(id int) string {
	return "action:" + strconv.Itoa(id)
}
