package msWatchCommon

import (
	// Internal dependencies.
	actions "github.com/krystalcode/go-mantis-shrimp/actions/common"
)

// All Watches should be implementing this interface.
// It simply defines a Do() function that prepares any data and evaluates any
// conditions.
type Watch interface {
	Do()
}

// All Watches should also be including the WatchBase as an embedded struct
// (anonymous field). It provides all fields that should be present in all
// Watch implementations.
// @I Add CreatedAt and UpdatedAt fields in Watches
type WatchBase struct {
	Name         string         `json:"name"`
	ActionsIds []int            `json:"actions_ids"`
	Actions    []actions.Action `json:"actions"`
}
