/**
 * Provides a type that holds the configuration for the Action API.
 */

package msActionConfig

import (
	// Internal dependencies.
	wrapper "github.com/krystalcode/go-mantis-shrimp/actions/wrapper"
)

// Config holds the configuration required for the Action API.
type Config struct {
	// The Storage configuration.
	Storage map[string]interface{} `json:"storage"`
	// Actions to be loaded in the case of using ephemeral storage.
	ActionWrappers []wrapper.ActionWrapper `json:"actions"`
}
