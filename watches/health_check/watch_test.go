/**
 * Tests for the Health Check Watch.
 */

package msWatchHealthCheck

import (
	// Testing packages.
	"github.com/stretchr/testify/assert"
	"testing"

	// Utilities.
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	// Internal dependencies.
	actions "github.com/krystalcode/go-mantis-shrimp/actions/common"
	common "github.com/krystalcode/go-mantis-shrimp/watches/common"
)

/**
 * Helper types and functions reused in various tests.
 */

// testWatch generates a Watch object with some defaults.
func testWatch() Watch {
	watch := Watch{
		common.WatchBase{
			"Test Watch",
			[]int{},
			[]actions.Action{},
		},
		"https://golang.org/pkg/testing/",
		[]int{200},
		30 * time.Second,
		[]Condition{},
		nil,
		Result{},
	}
	return watch
}

// An HTTP client that returns a response with status 200.
type MockHTTPClient200 struct{}

func (client MockHTTPClient200) Get(url string) (*http.Response, error) {
	response := &http.Response{
		Status: "200 OK",
		Body:   ioutil.NopCloser(bytes.NewBuffer([]byte{})),
	}

	return response, nil
}

// An HTTP client that returns a response with status 400.
type MockHTTPClient400 struct{}

func (client MockHTTPClient400) Get(url string) (*http.Response, error) {
	response := &http.Response{
		Status: "400 BadRequest",
		Body:   ioutil.NopCloser(bytes.NewBuffer([]byte{})),
	}

	return response, nil
}

// An HTTP client that returns an error, simulating an unresponsive URL or a
// network error.
type MockHTTPClientError struct{}

func (client MockHTTPClientError) Get(url string) (*http.Response, error) {
	return nil, fmt.Errorf("cannot reach the given URL within the given timeout")
}

/**
 * Test Result preparation depending on the HTTP Response.
 */

func TestResultPreparation_Success(t *testing.T) {
	watch := testWatch()
	client := MockHTTPClient200{}
	watch.SetHTTPClient(client)
	watch.data()

	assert.Equal(t, "success", watch.result.Status)
}

func TestResultPreparation_StatusMismatch(t *testing.T) {
	watch := testWatch()
	client := MockHTTPClient400{}
	watch.SetHTTPClient(client)
	watch.data()

	assert.Equal(t, "status_mismatch", watch.result.Status)
}

func TestResultPreparation_Inaccessible(t *testing.T) {
	watch := testWatch()
	client := MockHTTPClientError{}
	watch.SetHTTPClient(client)
	watch.data()

	assert.Equal(t, "inaccessible", watch.result.Status)
}

/**
 * Test combinations of Results (success, failure, inaccessible) and Conditions
 * (ConditionSuccess, ConditionFailure).
 */

func TestEvaluateConditionsOnSuccess_Success(t *testing.T) {
	watch := testWatch()
	condition := ConditionSuccess{}
	watch.result = Result{"success"}
	watch.Conditions = []Condition{condition}

	ok := watch.evaluate()
	assert.True(t, ok)
}

func TestEvaluateConditionsOnSuccess_Failure(t *testing.T) {
	watch := testWatch()
	condition := ConditionFailure{}
	watch.result = Result{"success"}
	watch.Conditions = []Condition{condition}

	ok := watch.evaluate()
	assert.False(t, ok)
}

func TestEvaluateConditionsOnFailure_Success(t *testing.T) {
	watch := testWatch()
	condition := ConditionSuccess{}
	watch.result = Result{"failure"}
	watch.Conditions = []Condition{condition}

	ok := watch.evaluate()
	assert.False(t, ok)
}

func TestEvaluateConditionsOnFailure_Failure(t *testing.T) {
	watch := testWatch()
	condition := ConditionFailure{}
	watch.result = Result{"failure"}
	watch.Conditions = []Condition{condition}

	ok := watch.evaluate()
	assert.True(t, ok)
}

func TestEvaluateConditionsOnInaccessible_Success(t *testing.T) {
	watch := testWatch()
	condition := ConditionSuccess{}
	watch.result = Result{"Inaccessible"}
	watch.Conditions = []Condition{condition}

	ok := watch.evaluate()
	assert.False(t, ok)
}

func TestEvaluateConditionsOnInaccessible_Failure(t *testing.T) {
	watch := testWatch()
	condition := ConditionFailure{}
	watch.result = Result{"Inaccessible"}
	watch.Conditions = []Condition{condition}

	ok := watch.evaluate()
	assert.True(t, ok)
}
