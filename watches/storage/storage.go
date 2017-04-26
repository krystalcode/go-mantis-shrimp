/**
 * Provides a storage API for persisting Watches.
 */

package msWatchStorage

import (
	// Utilities.
	"errors"
	"fmt"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/watches/common"
)

/**
 * Public API.
 */

// Storage interfaces that should be implemented by all providers.

type Storage interface {
	Get(int) common.Watch
	Set(common.Watch) int
}

type StorageFactory func(confing map[string]string) (Storage, error)

// Function that, given configuration, creates and returns a storage provider.
func Create(config map[string]string) (Storage, error) {
	// Register providers the first time we create a storage. We may create a more
	// generic registration mechanism when we support more storage providers that
	// may also be registered independently, but for now this is sufficient.
	if len(storageFactories) == 0 {
		storageFactories["redis"] = NewRedisStorage
	}

	engine, ok := config["STORAGE_ENGINE"]
	if !ok {
		err := errors.New(fmt.Sprintf("No storage engine provided."))
		panic(err)
	}

	factory, ok := storageFactories[engine]
	if !ok {
		err := errors.New(fmt.Sprintf("Unknown storage engine \"%s\".", engine))
		panic(err)
	}

	return factory(config)
}

/**
 * For internal use.
 */

var storageFactories = make(map[string]StorageFactory)
