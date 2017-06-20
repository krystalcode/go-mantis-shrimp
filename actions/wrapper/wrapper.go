/**
 * Provides a ActionWrapper that holds both an Action and its type.
 *
 * It helps manage Actions of different types in the following two cases:
 * - Decoding/encoding a JSON object, such as from an incoming request, that
 *   contains an Action and its type.
 * - Storing/retrieving Actions in a key-value store where information about the
 *   Action struct we need to use is part of the stored value.
 */

package msActionWrapper

import (
	// Utilities.
	"encoding/json"
	"fmt"
	"reflect"

	// Internal dependencies.
	chat "github.com/krystalcode/go-mantis-shrimp/actions/chat"
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
	mailgun "github.com/krystalcode/go-mantis-shrimp/actions/mailgun"
)

// ActionWrapper provides a structure that holds an Action together with its type.
type ActionWrapper struct {
	Type   string        `json:"type"`
	Action common.Action `json:"action"`
}

// UnmarshalJSON properly decodes an ActionWrapper JSON object by decoding the
// contained Action depending on the value of the "type" field.
func (wrapper *ActionWrapper) UnmarshalJSON(bytes []byte) error {
	// Get the first-level fields as json.RawMessage data.
	var jsonMap map[string]*json.RawMessage
	err := json.Unmarshal(bytes, &jsonMap)
	if err != nil {
		return err
	}

	// We need the type of the Action to be given, otherwise we cannot decode the
	// JSON data.
	if jsonMap["type"] == nil && jsonMap["action"] != nil {
		return fmt.Errorf("cannot decode ActionWrapper JSON object without given the Action's type")
	}

	// Get the Action type from the corresponding field.
	var actionType string
	err = json.Unmarshal(*jsonMap["type"], &actionType)
	if err != nil {
		return err
	}
	wrapper.Type = actionType

	if jsonMap["action"] == nil {
		return nil
	}

	// Decode the JSON data in the "action" field to a Action struct of the
	// appropriate type.
	switch actionType {
	case "chat_message":
		var action chat.Action
		err = json.Unmarshal(*jsonMap["action"], &action)
		if err != nil {
			return err
		}
		wrapper.Action = action
		break
	case "mailgun_message":
		var action mailgun.Action
		err = json.Unmarshal(*jsonMap["action"], &action)
		if err != nil {
			return err
		}
		wrapper.Action = action
		break
	default:
		return fmt.Errorf(
			"unknown Action type \"%s\" while trying to decode an ActionWrapper JSON object",
			actionType,
		)
	}

	return nil
}

// Wrapper creates an ActionWrapper for the given Action, based on its type that
// is detected via reflection.
func Wrapper(action common.Action) (*ActionWrapper, error) {
	var actionType string

	structType := reflect.TypeOf(action)
	switch structType.PkgPath() {
	case "github.com/krystalcode/go-mantis-shrimp/actions/chat":
		actionType = "chat_message"
		break
	case "github.com/krystalcode/go-mantis-shrimp/actions/mailgun":
		actionType = "mailgun_message"
		break
	default:
		err := fmt.Errorf(
			"unknown Action struct \"%s\" when trying to wrap an Action in a wrapper",
			structType,
		)
		return nil, err
	}

	wrapper := ActionWrapper{
		Type:   actionType,
		Action: action,
	}
	return &wrapper, nil
}

// ActionFactory is a function type that should be implemented by all Action
// factories. It defines a function that receives the JSON-encoded Action object
// (in raw bytes), and it returs the corresponding Action object, including any
// initializations required such as dependency injection.
type ActionFactory func(jsonAction *json.RawMessage) (common.Action, error)

// Create creates and returns an initialized Action object that corresponds to
// the type and holds the data defined to the given JSON-object byte slice.
func Create(jsonAction []byte) (common.Action, error) {
	// Register factories the first time we create an Action. We may create a more
	// generic registration mechanism when we support more Action factories that
	// may also be registered independently, but for now this is sufficient.
	if len(actionFactories) == 0 {
		actionFactories["chat_message"] = chat.NewChatMessageAction
		actionFactories["mailgun_message"] = mailgun.NewMailgunMessageAction
	}

	var jsonMap map[string]*json.RawMessage
	err := json.Unmarshal(jsonAction, &jsonMap)
	if err != nil {
		return nil, err
	}

	// We need to be given the type as well, otherwise we can't know what type of
	// Action we need to create.
	if jsonMap["type"] == nil {
		err = fmt.Errorf("trying to create an Action from a JSON object that does not contain the Action's type")
		return nil, err
	}

	var actionType string
	err = json.Unmarshal(*jsonMap["type"], &actionType)
	if err != nil {
		return nil, err
	}

	factory, ok := actionFactories[actionType]
	if !ok {
		err := fmt.Errorf("unknown Action factory for type \"%s\"", actionType)
		return nil, err
	}

	return factory(jsonMap["action"])
}

/**
 * For internal use.
 */

// actionFactories holds a map of all known Action factories.
var actionFactories = make(map[string]ActionFactory)
