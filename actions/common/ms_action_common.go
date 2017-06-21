package msActionCommon

// Action is an interface that should be implemented by all Watch types.
// It simply defines a Do() function that does whatever the Action is meant to
// do.
type Action interface {
	Do() error
}

// ActionBase should be included by all Action types as an embedded struct
// (anonymous field). It provides all fields that should be present in all
// Action implementations.
type ActionBase struct {
	// @I Add CreatedAt and UpdatedAt fields in Actions

	Name string `json:"name"`
}
