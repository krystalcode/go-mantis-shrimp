/**
 * Provides a Watch that checks the status of a webpage or a service.
 */

package msWatchHealthCheck

import (
	// Utilities.
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/watches/common"
)

/**
 * Types and their functions.
 */

// Watch implements the common.Watch interface. It provides a Watch that checks
// the health status of the defined URL. Its evaluation of whether the included
// Actions will be executed depend on the evaluation of its Conditions.
type Watch struct {
	common.WatchBase

	// The URL that will be checked.
	URL string `json:"url"`
	// The HTTP Status codes that are considered successful.
	Statuses []int `json:"statuses"`
	// How much to wait for the response before considering the URL inaccessible.
	Timeout time.Duration `json:"timeout"`
	// The Conditions that will evaluate the results to determine whether the
	// Actions should be triggered or not.
	Conditions []Condition `json:"conditions"`

	// The result of the data operation.
	result Result
}

// Do implements common.Watch.Do(). It prepares the Result of the Watch, it
// evalutes the Conditions, and returns the IDs of the Actions that should be
// triggered as a result of the Watch, if any.
func (watch Watch) Do() []int {
	watch.data()
	ok := watch.evaluate()

	if !ok {
		return []int{}
	}

	// If all conditions pass, return the IDs of the Actions that should be
	// triggered.
	// Store any Actions given in the Actions field and return their IDs as well.
	return watch.ActionsIds
}

// Makes a GET call to the URL defined in the Watch and determines the Result.
func (watch *Watch) data() {
	client := http.Client{
		Timeout: watch.Timeout,
	}
	res, err := client.Get(watch.URL)
	if err != nil {
		// @I Differentiate between lack of accessibility and timeout in health
		//    check watch
		watch.result = Result{Status: "inaccessible"}
		return
	}

	// Extract status as integer.
	resStatusInt, err := strconv.Atoi(res.Status[0:3])
	if err != nil {
		panic(err)
	}

	// Check whether the Response Status is one that is considered successful.
	statusMatch := false
	for _, status := range watch.Statuses {
		if resStatusInt == status {
			statusMatch = true
			break
		}
	}

	if !statusMatch {
		watch.result = Result{Status: "status_mismatch"}
		return
	}

	// If we got a response with one of the successful statuses, the result is
	// "success".
	watch.result = Result{Status: "success"}
}

// Go through all Conditions defined in the Watch and evaluate them. The
// Condtions are successful in their entirety when all Conditions evaluate
// successfully.
func (watch *Watch) evaluate() bool {
	// @I Support Condition operators in Watches that would allow combining
	//    Conditions in flexible ways
	// @I Consider abstracting the Watch.evalute() function so that it is reusable
	// @I Consider using Goroutines to evaluate multiple Conditions in a Watch

	allOk := true
	for _, condition := range watch.Conditions {
		ok := condition.Do(watch.result)
		if !ok {
			allOk = false
			break
		}
	}

	return allOk
}

// Result holds the result of a URL health check. It is simply a string that can
// hold one of the following values:
// - success
// - inaccessible
// - timeout
// - status_mismatch
type Result struct {
	Status string
}

// Condition is an interface that should be implemented by all Condition types
// for the Health Check Watch. It simply defines a function that, given the
// Result of a Health Check operation, it decides whether the Condition is met.
type Condition interface {
	Do(Result) bool
}

// ConditionSuccess implements the Condition interface, providing a Condition
// that is met when the Result of a Health Check is successful ("success").
type ConditionSuccess struct{}

// Do implements Condition.Do(), determining whether the Result of a Health
// Check operation is successful.
func (condition ConditionSuccess) Do(result Result) bool {
	if result.Status == "success" {
		return true
	}

	return false
}

// ConditionFailure implements the Condition interface, providing a Condition
// that is met when the Result of a Health Check is unsuccessful (any other
// result apart from "success").
type ConditionFailure struct{}

// Do implements Condition.Do(), determining whether the Result of a Health
// Check operation is unsuccessful.
func (condition ConditionFailure) Do(result Result) bool {
	if result.Status != "success" {
		return true
	}

	return false
}

/**
 * JSON.
 */

// MarshalJSON encodes a ConditionSuccess object into a JSON object that
// contains a single field, indicating its type. This is desired so that a
// JSON-encoded Watch object containing such a Condition can be then decoded
// based on the Condition type.
func (condition ConditionSuccess) MarshalJSON() ([]byte, error) {
	return []byte(`{"type":"success"}`), nil
}

// MarshalJSON encodes a ConditionFailure object into a JSON object that
// contains a single field, indicating its type. This is desired so that a
// JSON-encoded Watch object containing such a Condition can be then decoded
// based on the Condition type.
func (condition ConditionFailure) MarshalJSON() ([]byte, error) {
	return []byte(`{"type":"failure"}`), nil
}

// UnmarshalJSON provides decoding of a JSON-encoded Watch object so that the
// Conditions held in the "conditions" field are properly constructed based on
// their type.
func (watch *Watch) UnmarshalJSON(bytes []byte) error {
	// Deserialize everything into a map of json.RawMessage; its indices would
	// correspond to the Watch struct's fields.
	var jsonMap map[string]*json.RawMessage
	err := json.Unmarshal(bytes, &jsonMap)
	if err != nil {
		return err
	}

	// Decode all other fields first.
	// @I Find a generic way to override JSON decoding of a specific field without
	//    having to manually decode the rest of a struct's fields
	if jsonMap["name"] != nil {
		var name string
		err = json.Unmarshal(*jsonMap["name"], &name)
		if err != nil {
			return err
		}
		watch.Name = name
	}
	if jsonMap["actions_ids"] != nil {
		var actionsIds []int
		err = json.Unmarshal(*jsonMap["actions_ids"], &actionsIds)
		if err != nil {
			return err
		}
		watch.ActionsIds = actionsIds
	}
	if jsonMap["url"] != nil {
		var URL string
		err = json.Unmarshal(*jsonMap["url"], &URL)
		if err != nil {
			return err
		}
		watch.URL = URL
	}
	if jsonMap["statuses"] != nil {
		var statuses []int
		err = json.Unmarshal(*jsonMap["statuses"], &statuses)
		if err != nil {
			return err
		}
		watch.Statuses = statuses
	}
	if jsonMap["timeout"] != nil {
		var timeout time.Duration
		err = json.Unmarshal(*jsonMap["timeout"], &timeout)
		if err != nil {
			return err
		}
		watch.Timeout = timeout
	}

	// If no conditions are given, there's nothing to do; return or we'll get an
	// error.
	if jsonMap["conditions"] == nil {
		return nil
	}

	var rawConditions []*json.RawMessage
	err = json.Unmarshal(*jsonMap["conditions"], &rawConditions)
	if err != nil {
		return err
	}

	// Create a slice of the right size that will hold the Conditions.
	watch.Conditions = make([]Condition, len(rawConditions))

	// Decode the Conditions from their JSON structure and put them in the
	// corresponding field slice.
	for index, rawCondition := range rawConditions {
		var conditionInnerJSON map[string]*json.RawMessage
		err = json.Unmarshal(*rawCondition, &conditionInnerJSON)
		if err != nil {
			return err
		}

		// Get the type of the Condition.
		var conditionType string
		err = json.Unmarshal(*conditionInnerJSON["type"], &conditionType)
		if err != nil {
			return err
		}

		switch conditionType {
		case "success":
			watch.Conditions[index] = ConditionSuccess{}
			break
		case "failure":
			watch.Conditions[index] = ConditionFailure{}
			break
		default:
			return fmt.Errorf("unknown Condition type \"%s\"", conditionType)
		}
	}

	return nil
}
