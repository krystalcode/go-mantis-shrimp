package main

import (
	// Utilities.
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)


/**
 * Constants.
 */

// @I Make the Trigger API base url configurable
const TriggerApiBaseURL = "http://ms-trigger-api:8888"
const TriggerApiVersion = "1"


/**
 * Main program entry.
 *
 * The flow of the program is as follows. A goroutine searches for new items at
 * regular intervals. It sends any discovered items to a channel where they will
 * be filtered (evaluating trigger conditions to make sure we don't trigger
 * items that we shouldn't). Items that pass the filter are sent to another
 * channel that executes the triggering.
 *
 * @I Add a Scheduler API for accepting scheduler item submissions
 */
func main() {
	// Channel that receives items that are ready to be triggered.
	triggers := make(chan Item)

	// Channel that receives items that are candidate for triggering.
	items := make(chan Item)

	// Search for candidate items; it could be from a variety of sources.
	go search(items)

	// Listen to candidate items and send them for filtering and execution as they
	// come. We do this in a goroutine so that we don't block the program yet.
	go func() {
		for item := range items {
			go filter(item, triggers)
		}
	}()

	// Listen for items that are ready for triggering, and do so as they come. We
	// keep the channel open and the program stays on perpetual.
	for item := range triggers {
		trigger(item._id)
	}
}

// Regularly look for items candidate for triggering.
// @I Search for candidate items in a connected Elastic Search database
// @I Support different sources of candidate items configurable via JSON or YAML
func search(items chan<- Item) {
	newItems := loadItems()
	for _, item := range newItems {
		items <- item
	}
	time.Sleep(5 * time.Second)
	search(items)
}

// Filter candidate items to ensure that we don't run expired items or that we
// don't run them prematurely.
func filter(item Item, triggers chan<- Item) {
	now := time.Now()
	afterStart := item.start == nil || now.After(*item.start)
	beforeEnd  := item.end   == nil || now.Before(*item.end)
	if afterStart && beforeEnd {
		run(item, triggers)

		// @I Set active status to true to avoid requeueing the same item
		// @I Investigate throttling architecture and implementation
	}
}

// Execute triggering of items by sending to the corresponding channel.
// @I Consider rescheduling of triggered items via the corresponding channel
func run(item Item, triggers chan<- Item) {
	triggers <- item
	time.Sleep(item.interval)
	filter(item, triggers)
}

// Trigger an item by making a call to the Trigger API.
func trigger(_id string) {
	url  := TriggerApiBaseURL + "/v" + TriggerApiVersion + "/"
	body := []byte("{\"_id\":\"" + _id + "\"}")

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Extract status as integer from "200 OK"
	resStatus, err := strconv.Atoi(res.Status[0:3])
	if err != nil {
		panic(err)
	}

	if resStatus != http.StatusOK {
		resBody, _ := ioutil.ReadAll(res.Body)
		fmt.Println("Status:"  , res.Status)
		fmt.Println("Headers:" , res.Header)
		fmt.Println("Body:"    , string(resBody))
	}
}

// Temporary function that acts as a candidate items source.
func loadItems() []Item {
	items := []Item {
		Item {
			_id      : "1",
			interval : 1 * time.Second,
			enabled  : true,
			active   : false,
		},
		Item {
			_id      : "2",
			interval : 3 * time.Second,
			enabled  : true,
			active   : false,
		},
	}

	return items
}

type Item struct {
	// Unique Item identifier.
	_id      string
	// The Item should be triggered only between its start and end times.
	start    *time.Time
	end      *time.Time
	// How frequently the Item should be triggered.
	interval time.Duration
	// Items should not be triggered if they are disabled.
	enabled  bool
	// Whether the Item is active in the scheduler queue. It can be used to
	// prevent loading and triggering the same candidate item more than once.
	active   bool
}
