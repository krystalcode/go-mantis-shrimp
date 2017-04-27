package msWatchCommon

import (
	// Internal dependencies.
	actions "github.com/krystalcode/go-mantis-shrimp/actions/common"
)

// Watch is an interface that should be implemented by all Watch types.
// It simply defines a Do() function that prepares any data and evaluates any
// conditions. It returns a list of the IDs of the Actions that should be
// triggered as a result of the Watch, if any.
type Watch interface {
	Do() []int
}

// WatchBase should be included by all Watch types as an embedded struct
// (anonymous field). It provides all fields that should be present in all
// Watch implementations.
type WatchBase struct {
	// @I Add CreatedAt and UpdatedAt fields in Watches

	Name       string           `json:"name"`
	ActionsIDs []int            `json:"actions_ids"`
	Actions    []actions.Action `json:"actions"`
}
