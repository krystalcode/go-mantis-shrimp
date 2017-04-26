/**
 * Tests for the Health Check Watch.
 */

package msWatchHealthCheck

import (
	// Testing packages.
	"testing"
	"github.com/stretchr/testify/assert"

	// Utilities.
	"time"

	// Internal dependencies.
	common  "github.com/krystalcode/go-mantis-shrimp/watches/common"
	actions "github.com/krystalcode/go-mantis-shrimp/actions/common"
)

func testWatch() Watch {
	watch := Watch{
		common.WatchBase{
			"Test Watch",
			[]int{},
			[]actions.Action{},
		},
		"https://golang.org/pkg/testing/",
		[]int{ 200 },
		30 * time.Second,
		[]Condition{},
		Result{},
	}
	return watch
}

/**
 * Test combinations of Results (success, failure, inaccessible) and Conditions
 * (ConditionSuccess, ConditionFailure).
 */

func TestEvaluateConditionsOnSuccess_Success(t *testing.T) {
	watch := testWatch()
	condition := ConditionSuccess{}
	watch.result = Result{ "success" }
	watch.Conditions = []Condition{ condition }

	ok := watch.evaluate()
	assert.True(t, ok)
}

func TestEvaluateConditionsOnSuccess_Failure(t *testing.T) {
	watch := testWatch()
	condition := ConditionFailure{}
	watch.result = Result{ "success" }
	watch.Conditions = []Condition{ condition }

	ok := watch.evaluate()
	assert.False(t, ok)
}

func TestEvaluateConditionsOnFailure_Success(t *testing.T) {
	watch := testWatch()
	condition := ConditionSuccess{}
	watch.result = Result{ "failure" }
	watch.Conditions = []Condition{ condition }

	ok := watch.evaluate()
	assert.False(t, ok)
}

func TestEvaluateConditionsOnFailure_Failure(t *testing.T) {
	watch := testWatch()
	condition := ConditionFailure{}
	watch.result = Result{ "failure" }
	watch.Conditions = []Condition{ condition }

	ok := watch.evaluate()
	assert.True(t, ok)
}

func TestEvaluateConditionsOnInaccessible_Success(t *testing.T) {
	watch := testWatch()
	condition := ConditionSuccess{}
	watch.result = Result{ "Inaccessible" }
	watch.Conditions = []Condition{ condition }

	ok := watch.evaluate()
	assert.False(t, ok)
}

func TestEvaluateConditionsOnInaccessible_Failure(t *testing.T) {
	watch := testWatch()
	condition := ConditionFailure{}
	watch.result = Result{ "Inaccessible" }
	watch.Conditions = []Condition{ condition }

	ok := watch.evaluate()
	assert.True(t, ok)
}
