/**
 * Provides a type that holds the configuration for the Watch API.
 */

package msWatchConfig

import (
	// Internal dependencies.
	wrapper "github.com/krystalcode/go-mantis-shrimp/watches/wrapper"
)

// Config holds the configuration required for the Watch API.
type Config struct {
	// Configuration required for the Action API SDK.
	ActionAPI ConfigActionAPI `json:"action_api"`
	// The Storage configuration.
	Storage map[string]interface{} `json:"storage"`
	// Watches to be loaded in the case of using ephemeral storage.
	WatchWrappers []wrapper.WatchWrapper `json:"watches"`
}

// ConfigActionAPI holds the configuration required for making calls to the
// Action API.
type ConfigActionAPI struct {
	// The base url without a trailing slash.
	BaseURL string `json:"base_url"`
	// The API version.
	Version string `json:"version"`
}
