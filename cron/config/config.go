/**
 * Provides a type that holds the configuration for the Cron component.
 */

package msCronConfig

import (
	// Internal dependencies.
	schedule "github.com/krystalcode/go-mantis-shrimp/cron/schedule"
)

// Config holds the configuration required for the Cron component.
type Config struct {
	// Configuration required for the Watch API SDK.
	WatchAPI ConfigWatchAPI `json:"watch_api"`
	// The search interval.
	SearchInterval string `json:"search_interval"`
	// The Storage configuration.
	Storage map[string]interface{} `json:"storage"`
	// Schedules to be loaded in the case of using ephemeral storage.
	Schedules []schedule.Schedule `json:"schedules"`
}

// ConfigWatchAPI holds the configuration required for making calls to the Watch
// API.
type ConfigWatchAPI struct {
	// The base url without a trailing slash.
	BaseURL string `json:"base_url"`
	// The API version.
	Version string `json:"version"`
}
