/**
 * Provides an SDK for communicating with the Action API.
 */

package msActionSDK

import (
	// Utilities.
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Config holds any configuration required to perform calls to the Action API.
type Config struct {
	BaseURL string
	Version string
}

// TriggerByID makes a POST request that triggers the Action that corresponds to
// the given ID.
func TriggerByID(id int, config Config) error {
	// Prepare the URL and the request body.
	idString := strconv.Itoa(id)
	url := config.BaseURL + "/v" + config.Version + "/" + idString + "/trigger"
	body := []byte{}

	// Make the request.
	client := &http.Client{}
	res, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Response status should always be 200.
	if res.StatusCode != http.StatusOK {
		resBody, ioErr := ioutil.ReadAll(res.Body)
		if ioErr != nil {
			err = fmt.Errorf(
				"response Status not \"200 OK\" when triggering an Action by its ID; Status: \"%d\", Headers: \"%s\", Body: An error occurred while decoding the body: \"%s\"",
				res.StatusCode,
				res.Header,
				ioErr,
			)
			return err
		}
		err = fmt.Errorf(
			"response Status not \"200 OK\" when triggering an Action by its ID; Status: \"%d\", Headers: \"%s\", Body: \"%s\"",
			res.StatusCode,
			res.Header,
			resBody,
		)
		return err
	}

	return nil
}
