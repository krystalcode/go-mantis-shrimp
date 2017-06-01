/**
 * Provides a storage API for persisting Watches.
 */

package msWatchStorage

import (
	// Utilities.
	"fmt"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/watches/common"
)

/**
 * Public API.
 */

// Storage is an interface that should be implemented by all Storage engines.
// It defines an API for storing and retrieving Watch objects.
type Storage interface {
	Create(*common.Watch) (*int, error)
	Get(int) (*common.Watch, error)
	Update(int, *common.Watch) error
}

// StorageFactory is a function type that should be implemented by all Storage
// engine factories. It defines a function type that receives the required
// configuration as a map, and it returns the Storage engine object. The
// configuration should include the requested engine keyed "STORAGE_ENGINE" plus
// any configuration required by the engine itself.
type StorageFactory func(confing map[string]interface{}) (Storage, error)

// Create creates and returns a Storage provider, given the configuration that
// includes configuration required by the provider.
func Create(config map[string]interface{}) (Storage, error) {
	// Register providers the first time we create a storage. We may create a more
	// generic registration mechanism when we support more storage providers that
	// may also be registered independently, but for now this is sufficient.
	if len(storageFactories) == 0 {
		storageFactories["redis"] = NewRedisStorage
	}

	storageType, ok := config["type"]
	if !ok {
		err := fmt.Errorf("the \"type\" configuration option is required for defining the storage engine")
		return nil, err
	}

	sStorageType := storageType.(string)
	if sStorageType == "" {
		err := fmt.Errorf("no storage engine provided")
		return nil, err
	}

	factory, ok := storageFactories[sStorageType]
	if !ok {
		err := fmt.Errorf("unknown storage engine \"%s\"", sStorageType)
		return nil, err
	}

	storage, err := factory(config)
	if err != nil {
		return nil, err
	}

	return storage, nil
}

/**
 * For internal use.
 */

// storageFactories holds a map of all known Storage factories.
var storageFactories = make(map[string]StorageFactory)
