/**
 * Provides a storage API for persisting Schedules.
 */

package msCronStorage

import (
	// Utilities.
	"fmt"
	"time"

	// Internal dependencies.
	schedule "github.com/krystalcode/go-mantis-shrimp/cron/schedule"
)

/**
 * Public API.
 */

// Storage is an interface that should be implemented by all Storage engines.
// It defines an API for storing and retrieving Schedule objects.
type Storage interface {
	// @I Split Storage API Set function into a Create and an Update function
	// @I Implement Delete function in the Storage API

	Get(int) (*schedule.Schedule, error)
	Set(*schedule.Schedule) (*int, error)
	Search(time.Duration) ([]*schedule.Schedule, error)
}

// StorageFactory is a function type that should be implemented by all Storage
// engine factories. It defines a function type that receives the required
// configuration as a map, and it returns the Storage engine object. The
// configuration should include the requested engine keyed "STORAGE_ENGINE" plus
// any configuration required by the engine itself.
type StorageFactory func(confing map[string]string) (Storage, error)

// Create creates and returns a Storage provider, given the configuration that
// includes configuration required by the provider.
func Create(config map[string]string) (Storage, error) {
	// Register providers the first time we create a storage. We may create a more
	// generic registration mechanism when we support more storage providers that
	// may also be registered independently, but for now this is sufficient.
	if len(storageFactories) == 0 {
		storageFactories["redis"] = NewRedisStorage
	}

	engine, ok := config["STORAGE_ENGINE"]
	if !ok {
		err := fmt.Errorf("no storage engine provided")
		panic(err)
	}

	factory, ok := storageFactories[engine]
	if !ok {
		err := fmt.Errorf("unknown storage engine \"%s\"", engine)
		panic(err)
	}

	return factory(config)
}

/**
 * For internal use.
 */

// storageFactories holds a map of all known Storage factories.
var storageFactories = make(map[string]StorageFactory)
