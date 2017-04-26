/**
 * Provides a redis storage adapter for storing and retrieving Watches.
 */

package msWatchStorage

import (
	// Utilities
	"encoding/json"
	"fmt"
	"strconv"

	// Redis.
	"github.com/mediocregopher/radix.v2/redis"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/watches/common"
	wrapper "github.com/krystalcode/go-mantis-shrimp/watches/wrapper"
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

// Get implements Storage.Get(). It retrieves from Storage and returns the Watch
// for the given ID.
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

	// We store the Watches as WatchWrappers that contain the Watch type as well.
	var wrapper wrapper.WatchWrapper
	json.Unmarshal(jsonWatch, &wrapper)

	return wrapper.Watch
}

// Set implements Storage.Set(). It stores the given Watch object to the Redis
// Storage.
func (storage Redis) Set(watch common.Watch) int {
	// @I Consider using hashmaps instead of json values
	// @I Investigate risk of a Watch overriding another due to race conditions when
	//    creating them

	if storage.client == nil {
		panic("The Redis client has not been initialized yet.")
	}

	// We'll be storing a WatchWrapper which contains the Watch type as well.
	wrapper, err := wrapper.Wrapper(watch)
	if err != nil {
		panic(err)
	}
	jsonWatch, err := json.Marshal(wrapper)
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

// generateID generates an ID for a new Watch by incrementing the last known
// Watch ID.
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

// Generate a Redis key for the given Watch ID.
func redisKey(_id int) string {
	return "watch:" + strconv.Itoa(_id)
}
