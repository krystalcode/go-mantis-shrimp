package msCronCommon

// Schedule is an interface that should be implemented by all Schedule types.
// It simply defines a Do() function that decides whether the Schedule's Watches
// should be triggered, and return the Watches' IDs.
type Schedule interface {
	Do() []int
}

// ScheduleBase should be included by all Schedule types as an embedded struct
// (anonymous field). It provides all fields that should be present in all
// Schedule implementations.
type ScheduleBase struct {
	// @I Add CreatedAt and UpdatedAt fields in Schedules.

	// A list of the Watches that are scheduled, given by their IDs.
	WatchesIDs []int `json:"watches_ids"`
}
