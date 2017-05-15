/**
 * Tests for the Util module.
 */

package msUtil

import (
	// Testing packages.
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringToIntegers_Success(t *testing.T) {
	sIDs := "1, 2,3 "

	aIDsIntDesired := make(map[int]struct{})
	aIDsIntDesired[1] = struct{}{}
	aIDsIntDesired[2] = struct{}{}
	aIDsIntDesired[3] = struct{}{}

	aIDsIntResult, err  := StringToIntegers(sIDs, ",")

	// There shouldn't be any errors while converting to integers.
	assert.Nil(t, err)

	// The result should be the desired one.
	assert.Equal(t, aIDsIntDesired, aIDsIntResult)
}

func TestStringToIntegers_ContainsString(t *testing.T) {
	sIDs := "1, 2,h "
	aIDsIntResult, err  := StringToIntegers(sIDs, ",")

	// There should be an errors while converting to integers, while the result
	// should be nil.
	assert.NotNil(t, err)
	assert.Nil(t, aIDsIntResult)
}
