/**
 * Provides a type that holds Schedules for triggering Watches.
 */

package msCronSchedule

import (
	// Utilities.
	"time"
)

/**
 * Public API.
 */

// Schedule holds the data required for defining when a set of Watches should be
// triggered e.g. how frequently and between which start and end times.
type Schedule struct {
	// Triggering the Watches is limited only between the start and end times.
	Start *time.Time `json:"start"`
	Stop  *time.Time `json:"stop"`

	// How frequently the Watches should be triggered.
	Interval time.Duration `json:"interval"`

	// A list of the Watches that are scheduled, given by their IDs.
	WatchesIDs []int `json:"watches_ids"`

	// Boolean field that allows the ability to disable a Schedule.
	Enabled bool `json:"enabled"`
}

// Do ensures that any additional conditions are met before triggering the
// Watches. It returns the IDs of the Watches that should be triggered.
func (schedule Schedule) Do() []int {
	// Make sure that we are within the time frame of the schedule.
	// This filtering is also done by the search function that fetches the
	// Schedules, but let's keep an additional check here since it doesn't cost a
	// lot until the search functionality is battled tested.
	now := time.Now()
	afterStart := schedule.Start == nil || now.After(*schedule.Start)
	beforeEnd := schedule.Stop == nil || now.Before(*schedule.Stop)
	if afterStart && beforeEnd {
		return schedule.WatchesIDs
	}

	return nil
}
