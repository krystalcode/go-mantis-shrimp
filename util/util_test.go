/**
 * Tests for the Util module.
 */

package msUtil

import (
	// Testing packages.
	"github.com/stretchr/testify/assert"
	"testing"

	// Utilities.
	"os"
	"path"
)

/**
 * Tests.
 */

func TestStringToIntegers_Success(t *testing.T) {
	sIDs := "1, 2,3 "

	aIDsIntDesired := make(map[int]struct{})
	aIDsIntDesired[1] = struct{}{}
	aIDsIntDesired[2] = struct{}{}
	aIDsIntDesired[3] = struct{}{}

	aIDsIntResult, err := StringToIntegers(sIDs, ",")

	// There shouldn't be any errors while converting to integers.
	assert.Nil(t, err)

	// The result should be the desired one.
	assert.Equal(t, aIDsIntDesired, aIDsIntResult)
}

func TestStringToIntegers_ContainsString(t *testing.T) {
	sIDs := "1, 2,h "
	aIDsIntResult, err := StringToIntegers(sIDs, ",")

	// There should be an errors while converting to integers, while the result
	// should be nil.
	assert.NotNil(t, err)
	assert.Nil(t, aIDsIntResult)
}

func TestReadJSONFile_Success(t *testing.T) {
	structDesired := CorrectJSONStruct{"A"}
	var structResult CorrectJSONStruct

	currentDir, err := os.Getwd()
	assert.Nil(t, err)
	filename := path.Join(currentDir, "struct_test.json")
	err = ReadJSONFile(filename, &structResult)

	assert.Nil(t, err)
	assert.Equal(t, structDesired, structResult)
}

func TestReadJSONFile_Failure(t *testing.T) {
	var structResult WrongJSONStruct

	// Error while loading the file .
	err := ReadJSONFile("/file/that/does/not/exist", &structResult)
	assert.NotNil(t, err)

	// Error while converting JSON data to struct.
	currentDir, err := os.Getwd()
	assert.Nil(t, err)
	filename := path.Join(currentDir, "struct_test.json")
	err = ReadJSONFile(filename, &structResult)
	assert.NotNil(t, err)
}

/**
 * Functions/types for internal use.
 */

type CorrectJSONStruct struct {
	SomeString string `json:"some_string"`
}

type WrongJSONStruct struct {
	SomeInteger int `json:"some_string"`
}
