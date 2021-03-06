package msUtil

import (
	// Utilities.
	"encoding/json"
	"io/ioutil"
	"strconv"
	"strings"
)

// StringToIntegers converts an input of comma-separated string values to
// a map where the keys are the input values converted to integers. We return
// them as the keys to keep the algorithm more efficient. Since we will be
// iterating the result with a "for" loop, it does not make any difference
// really.
func StringToIntegers(input string, delimiter string) (map[int]struct{}, error) {
	aString := strings.Split(input, delimiter)
	aInt := make(map[int]struct{})

	for _, s := range aString {
		i, err := strconv.Atoi(strings.Trim(s, " "))
		if err != nil {
			return nil, err
		}

		if _, ok := aInt[i]; !ok {
			aInt[i] = struct{}{}
		}
	}

	return aInt, nil
}

// ReadJSONFile loads a file containing JSON data into the given struct pointer.
// Note that the compiler cannot check whether the provided value is a pointer
// and not giving a pointer to a struct will throw a runtime error.
func ReadJSONFile(filepath string, object interface{}) error {
	bytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, object)
	if err != nil {
		return err
	}

	return nil
}
