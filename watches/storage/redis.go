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

// Create implements Storage.Create(). It stores the given Watch object as a new
// value in the Redis Storage and it returns an automatically generated ID.
func (storage Redis) Create(watchPointer *common.Watch) (*int, error) {
	// Generate an ID and store the Watch.
	watchID, err := storage.generateID()
	if err != nil {
		return nil, err
	}

	err = storage.set(*watchID, watchPointer)
	if err != nil {
		return nil, err
	}

	return watchID, nil
}

// Get implements Storage.Get(). It retrieves from Storage and returns the Watch
// for the given ID.
func (storage Redis) Get(id int) (*common.Watch, error) {
	// @I Delegate error handling to the caller in Storage API functions

	if storage.client == nil {
		return nil, fmt.Errorf("the Redis client has not been initialized yet")
	}

	key := redisKey(id)

	r := storage.client.Cmd("GET", key)
	if r.Err != nil {
		return nil, r.Err
	}

	jsonWatch, err := r.Bytes()
	// If an error happens here, it should be because there is no value for this
	// key. It could be the case that the data is corrupted or the wrong data is
	// stored, we should see how to handle this later.
	// @I Handle edge cases when deserializing json in Redis
	if err != nil {
		return nil, err
	}

	// Create and initialize a Watch object based on the given JSON object.
	watch, err := wrapper.Create(jsonWatch)
	if err != nil {
		return nil, err
	}

	return &watch, nil
}

// Update implements Storage.Update(). It stores the given Watch object as a
// value in the Redis Storage, overriding the existing value with the given ID.
func (storage Redis) Update(watchID int, watchPointer *common.Watch) error {
	return storage.set(watchID, watchPointer)
}

// set stores a Watch object as a Redis value at the key corresponding to the
// given ID.
func (storage Redis) set(watchID int, watchPointer *common.Watch) error {
	// @I Consider using hashmaps instead of json values when storing Watches

	if storage.client == nil {
		return fmt.Errorf("the Redis client has not been initialized yet")
	}

	// We'll be storing a WatchWrapper which contains the Watch type as well.
	watch := *watchPointer
	wrapper, err := wrapper.Wrapper(watch)
	if err != nil {
		return err
	}
	jsonWatch, err := json.Marshal(wrapper)
	if err != nil {
		return err
	}

	// Store the Watch, and update the Watches index set.
	key := redisKey(watchID)
	err = storage.client.Cmd("SET", key, jsonWatch).Err
	if err != nil {
		return err
	}
	// @I Add the Watch's ID to the index only when creating it
	err = storage.client.Cmd("ZADD", "watches", watchID, key).Err
	if err != nil {
		return err
	}

	return nil
}

// generateID generates an ID for a new Watch by incrementing the last known
// Watch ID.
func (storage Redis) generateID() (*int, error) {
	// @I Investigate risk of a Watch overriding another due to race conditions when
	//    creating them

	if storage.client == nil {
		return nil, fmt.Errorf("the Redis client has not been initialized yet")
	}

	// Get the last ID that exists on the Watches index set, so that we can generate
	// the next one.
	r, err := storage.client.Cmd("ZREVRANGE", "watches", 0, 0, "WITHSCORES").List()
	if err != nil {
		return nil, err
	}

	// If there are no Watches yet, start with ID 1.
	if len(r) == 0 {
		newWatchID := 1
		return &newWatchID, nil
	}

	latestWatchID, err := strconv.Atoi(r[1])
	if err != nil {
		return nil, err
	}

	newWatchID := latestWatchID + 1
	return &newWatchID, nil
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

// Generate a Redis key for the given Watch ID.
func redisKey(id int) string {
	return "watch:" + strconv.Itoa(id)
}
