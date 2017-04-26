/**
 * Provides an SDK for communicating with the Action API.
 */

package msActionSDK

import (
	// Utilities.
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// Structure that holds any configuration required.
type Config struct {
	BaseURL string
	Version string
}

// Trigger an individual Action, given its ID.
func TriggerById(_id int, config Config) error {
	// Prepare the URL and the request body.
	idString := strconv.Itoa(_id)
	url  := config.BaseURL + "/v" + config.Version + "/" + idString + "/trigger"
	body := []byte{}

	// Make the request.
	client := &http.Client{}
	res, err := client.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Extract status as integer from "200 OK"
	resStatus, err := strconv.Atoi(res.Status[0:3])
	if err != nil {
		return err
	}

	// Response status should always be 200.
	if resStatus != http.StatusOK {
		resBody, ioErr := ioutil.ReadAll(res.Body)
		if ioErr != nil {
			err = errors.New(
				fmt.Sprintf(
					"Response Status not \"200 OK\" when triggering an Action by its ID. Status: \"%s\", Headers: \"%s\", Body: An error occurred while decoding the body: \"%s\".",
					res.Status,
					res.Header,
					ioErr,
				),
			)
			return err
		}
		err = errors.New(
			fmt.Sprintf(
				"Response Status not \"200 OK\" when triggering an Action by its ID. Status: \"%s\", Headers: \"%s\", Body: \"%s\".",
				res.Status,
				res.Header,
				resBody,
			),
		)
		return err
	}

	return nil
}
