/**
 * Provides a redis storage adapter for storing and retrieving Watches.
 */

package msWatchStorage

import (
	// Utilities
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	// Redis.
  "github.com/mediocregopher/radix.v2/redis"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/watches/common"
	health "github.com/krystalcode/go-mantis-shrimp/watches/health_check"
)

/**
 * Redis storage provider.
 */

// Implements the Storage interface.
type Redis struct {
	dsn     string
	client *redis.Client
}

// Get a Watch by its ID.
func (storage Redis) Get(_id int) common.Watch {
	if storage.client == nil {
		panic("The Redis client has not been initialized yet.")
	}

	key := redisKey(_id)

	r := storage.client.Cmd("GET", key)
	if r.Err != nil {
		panic(r.Err)
	}

	jsonWatch, err := r.Bytes()
	// If an error happens here, it should be because there is no value for this
	// key. It could be the case that the data is corrupted or the wrong data is
	// stored, we should see how to handle this later.
	// @I Handle edge cases when deserializing json in Redis
	if err != nil {
		return nil
	}

	// @I Dynamically detect the Watch type and convert json to struct
	//    accordingly
	watch := health.Watch{}
	json.Unmarshal(jsonWatch, &watch)

	return watch
}

// Set an Watch.
// @I Consider using hashmaps instead of json values
// @I Investigate risk of a Watch overriding another due to race conditions when
//    creating them
func (storage Redis) Set(watch common.Watch) int {
	if storage.client == nil {
		panic("The Redis client has not been initialized yet.")
	}

	jsonWatch, err := json.Marshal(watch)
	if err != nil {
		panic(err)
	}

	// Generate an ID, store the Watch, and update the Watches index set.
	_id := storage.generateID()
	key := redisKey(_id)
	err = storage.client.Cmd("SET", key, jsonWatch).Err
	if err != nil {
		panic(err)
	}
	err = storage.client.Cmd("ZADD", "watches", _id, key).Err
	if err != nil {
		panic(err)
	}

	return _id
}

func (storage Redis) generateID() int {
	// Get the last ID that exists on the Watches index set, so that we can generate
	// the next one.
	r, err := storage.client.Cmd("ZREVRANGE", "watches", 0, 0, "WITHSCORES").List()
	if err != nil {
		panic(err)
	}

	// If there are no watches yet, start with ID 1.
	if len(r) == 0 {
		return 1
	}

	_id, err := strconv.Atoi(r[1])
	if err != nil {
		panic(err)
	}

	return _id + 1
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

/**
 * For internal use.
 */

// Generate a Redis key for the given Watch ID.
func redisKey(_id int) string {
	return "watch:" + string(_id)
}
