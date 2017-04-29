package msCronBasic

import (
	// Utilities.
	"time"
)

import (
	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/cron/common"
)

/**
 * Types and their functions.
 */

// Schedule implements the common.Schedule interface. It provides a basic
// Schedule that defines when a set of Watches should be triggered e.g. how
// frequently and between which start and end times.
type Schedule struct {
	common.ScheduleBase

	// Triggering the Watches is limited only between the start and end times.
	start time.Time `json:"start"`
	stop  time.Time `json:"stop"`
	// How frequently the Watches should be triggered.
	interval time.Duration `json:"interval"`
	// Boolean field that allows the ability to disable a Schedule.
	enabled bool `json:"enabled"`
	// Whether the Schedule is active in the Scheduler queue. It can be used to
	// prevent loading and triggering the same candidate Watches more than once.
	active bool `json:"active"`
}

// Do implements common.Schedule.Do().
func (schedule *Schedule) Do() []int {
	now := time.Now()
	afterStart := &schedule.start == nil || now.After(schedule.start)
	beforeEnd := &schedule.stop == nil || now.Before(schedule.stop)
	if afterStart && beforeEnd {
		return schedule.WatchesIDs
	}

	return nil
}
