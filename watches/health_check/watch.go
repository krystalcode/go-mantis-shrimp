/**
 * Provides a Watch that checks the status of a webpage or a service.
 */

package msWatchHealthCheck

import (
	// Utilities.
	"encoding/json"
	"errors"
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
	Conditions  []Condition `json:"conditions"`

	// The result of the data operation.
	result Result
}

// Implements common.Watch.Do().
func (watch Watch) Do() {
	watch.data()
	ok := watch.evaluate()

	if !ok {
		return
	}

	// If all conditions pass, trigger the Actions.
	// @I Implement triggering Actions when evaluating a Watch's conditions
	// succeeds.
	fmt.Println("All conditions pass, the Actions should be triggered now.")
}

// Makes a GET call to the URL defined in the Watch and determines the Result.
func (watch *Watch) data() {
	client := http.Client {
		Timeout : watch.Timeout,
	}
	res, err := client.Get(watch.URL)
	if err != nil {
		// @I Differentiate between lack of accessibility and timeout in health
		//    check watch
		watch.result = Result { Status : "inaccessible" }
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
		watch.result = Result { Status : "status_mismatch"}
		return
	}

	// If we got a response with one of the successful statuses, the result is
	// "success".
	watch.result = Result { Status : "success" }
}

// Go through all Conditions defined in the Watch and evaluate them. The
// Condtions are successful in their entirety when all Conditions evaluate
// successfully.
// @I Support Condition operators in Watches that would allow combining
//    Conditions in flexible ways
// @I Consider abstracting the Watch.evalute() function so that it is reusable
func (watch *Watch) evaluate() bool {
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

// The result of a URL health check. Simply a string that can hold one of the
// following values:
// - success
// - inaccessible
// - timeout
// - status_mismatch
type Result struct {
	Status string
}

// All Conditions should implement this interface.
// It simply defines a function that, given the Result of a health check
// operation, decides whether it passes the Condition.
type Condition interface {
	Do(Result) bool
}

// Condition that succeeds when the health check is successful.
type ConditionSuccess struct {}
func (condition ConditionSuccess) Do(result Result) bool {
	if result.Status == "success" {
		return true
	}

	return false
}

// Condition that succeeds when the health check has failed for whatever reason.
type ConditionFailure struct {}
func (condition ConditionFailure) Do(result Result) bool {
	if result.Status != "success" {
		return true
	}

	return false
}

/**
 * JSON.
 */

func (condition ConditionSuccess) MarshalJSON() ([]byte, error) {
	return []byte(`{"type":"success"}`), nil
}

func (condition ConditionFailure) MarshalJSON() ([]byte, error) {
	return []byte(`{"type":"failure"}`), nil
}

// We need some special handling for decoding JSON since the Conditions can be
// of different types.
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
			return errors.New(fmt.Sprintf("Unknown Condition type \"%s\"", conditionType))
		}
	}

	return nil
}
