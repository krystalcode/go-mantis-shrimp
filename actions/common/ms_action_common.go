package msActionCommon

// All actions should be implementing this interface.
// It simply defines a Do() function that performs the action.
type Action interface {
	Do()
}

// All actions should also be including the ActionBase as an embedded struct
// (anonymous field). It provides all fields that should be present in all
// action implementations.
type ActionBase struct {
	Name string
}
