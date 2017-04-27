/**
 * Provides a WatchWrapper that holds both a Watch and its type.
 *
 * It helps manage Watches of different types in the following two cases:
 * - Decoding/encoding a JSON object, such as from an incoming request, that
 *   contains a Watch and its type.
 * - Storing/retrieving Watches in a key-value store where information about the
 *   Watch struct we need to use is part of the stored value.
 */

package msWatchWrapper

import (
	// Utilities.
	"encoding/json"
	"fmt"
	"reflect"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/watches/common"
	health "github.com/krystalcode/go-mantis-shrimp/watches/health_check"
)

// WatchWrapper provides a structure that holds a Watch together with its type.
type WatchWrapper struct {
	Type  string       `json:"type"`
	Watch common.Watch `json:"watch"`
}

// UnmarshalJSON properly decodes a WatchWrapper JSON object by decoding the
// contained Watch depending on the value of the "type" field.
func (wrapper *WatchWrapper) UnmarshalJSON(bytes []byte) error {
	// Get the first-level fields as json.RawMessage data.
	var jsonMap map[string]*json.RawMessage
	err := json.Unmarshal(bytes, &jsonMap)
	if err != nil {
		return err
	}

	// We need the type of the Watch to be given, otherwise we cannot decode the
	// JSON data.
	if jsonMap["type"] == nil && jsonMap["watch"] != nil {
		return fmt.Errorf("cannot decode WatchWrapper JSON object without given the Watch's type")
	}

	// Get the Watch type from the corresponding field.
	var watchType string
	err = json.Unmarshal(*jsonMap["type"], &watchType)
	if err != nil {
		return err
	}
	wrapper.Type = watchType

	if jsonMap["watch"] == nil {
		return nil
	}

	// Decode the JSON data in the "watch" field to a Watch struct of the
	// appropriate type.
	switch watchType {
	case "health_check":
		var watch health.Watch
		err = json.Unmarshal(*jsonMap["watch"], &watch)
		if err != nil {
			return err
		}
		wrapper.Watch = watch
		break
	default:
		return fmt.Errorf(
			"unknown Watch type \"%s\" while trying to decode a WatchWrapper JSON object",
			watchType,
		)
	}

	return nil
}

// Wrapper creates a WatchWrapper for the given Watch, based on its type that is
// detected via reflection.
func Wrapper(watch common.Watch) (*WatchWrapper, error) {
	var watchType string

	structType := reflect.TypeOf(watch)
	switch structType.PkgPath() {
	case "github.com/krystalcode/go-mantis-shrimp/watches/health_check":
		watchType = "health_check"
		break
	default:
		err := fmt.Errorf(
			"unknown Watch struct \"%s\" when trying to wrap a Watch in a wrapper",
			structType,
		)
		return nil, err
	}

	wrapper := WatchWrapper{
		Type:  watchType,
		Watch: watch,
	}
	return &wrapper, nil
}

// WatchFactory is a function type that should be implemented by all Watch
// factories. It defines a function that receives the JSON-encoded Watch object
// (in raw bytes), and it returs the corresponding Watch object, including any
// initializations required such as dependency injection.
type WatchFactory func(jsonWatch *json.RawMessage) (common.Watch, error)

// Create creates and returns an initialized Watch object that corresponds to
// the type and holds the data defined to the given JSON-object byte slice.
func Create(jsonWatch []byte) (common.Watch, error) {
	// Register factories the first time we create a Watch. We may create a more
	// generic registration mechanism when we support more Watch factories that
	// may also be registered independently, but for now this is sufficient.
	if len(watchFactories) == 0 {
		watchFactories["health_check"] = health.NewHealthCheckWatch
	}

	var jsonMap map[string]*json.RawMessage
	err := json.Unmarshal(jsonWatch, &jsonMap)
	if err != nil {
		return nil, err
	}

	// We need to be given the type as well, otherwise we can't know what type of
	// Watch we need to create.
	if jsonMap["type"] == nil {
		err = fmt.Errorf("trying to create a Watch from a JSON object that does not contain the Watch's type")
		return nil, err
	}

	var watchType string
	err = json.Unmarshal(*jsonMap["type"], &watchType)
	if err != nil {
		return nil, err
	}

	factory, ok := watchFactories[watchType]
	if !ok {
		err := fmt.Errorf("unknown Watch factory for type \"%s\"", watchType)
		return nil, err
	}

	return factory(jsonMap["watch"])
}

/**
 * For internal use.
 */

// watchFactories holds a map of all known Watch factories.
var watchFactories = make(map[string]WatchFactory)
