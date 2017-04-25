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
	"errors"
	"fmt"
	"reflect"

	// Internal dependencies.
	common  "github.com/krystalcode/go-mantis-shrimp/watches/common"
	health  "github.com/krystalcode/go-mantis-shrimp/watches/health_check"
)

// Wrapper structure that holds the type of the Watch as well.
type WatchWrapper struct {
	Type  string        `json:"type"`
	Watch common.Watch  `json:"watch"`
}

// Properly decode a Wrapper JSON object and the contained Watch depending on
// the value of the "type" field.
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
		return errors.New("Cannot decode WatchWrapper JSON object without given the Watch's type.")
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
		return errors.New(
			fmt.Sprintf(
				"Unknown Watch type \"%s\" while trying to decode a WatchWrapper JSON object",
				watchType,
			),
		)
	}

	return nil
}

// Given a Watch, create a WatchWrapper based on its type detected via
// reflection.
func Wrapper(watch common.Watch) (*WatchWrapper, error) {
	var watchType string

	structType := reflect.TypeOf(watch)
	switch structType.PkgPath() {
	case "github.com/krystalcode/go-mantis-shrimp/watches/health_check":
		watchType = "health_check"
		break
	default:
		err := errors.New(
			fmt.Sprintf(
				"Unknown Watch struct \"%s\" when trying to wrap a Watch in a wrapper.",
				structType,
			),
		)
		return nil, err
	}

	wrapper := WatchWrapper{
		Type  : watchType,
		Watch : watch,
	}
	return &wrapper, nil
}
