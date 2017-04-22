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

// @I Make the Watch API base url configurable
const WatchApiBaseURL = "http://ms-watch-api:8888"
const WatchApiVersion = "1"


/**
 * Main program entry.
 *
 * The flow of the program is as follows. A goroutine searches for new Watches at
 * regular intervals. It sends any discovered Watches to a channel where they will
 * be filtered (evaluating trigger conditions to make sure we don't trigger
 * Watches that we shouldn't). Watches that pass the filter are sent to another
 * channel that executes the triggering.
 *
 * @I Add a Scheduler API for accepting scheduler Watch submissions
 */
func main() {
	// Channel that receives Watches that are ready to be triggered.
	triggers := make(chan Watch)

	// Channel that receives Watches that are candidate for triggering.
	watches := make(chan Watch)

	// Search for candidate Watches; it could be from a variety of sources.
	go search(watches)

	// Listen to candidate Watches and send them for filtering and execution as they
	// come. We do this in a goroutine so that we don't block the program yet.
	go func() {
		for watch := range watches {
			go filter(watch, triggers)
		}
	}()

	// Listen for Watches that are ready for triggering, and do so as they come. We
	// keep the channel open and the program stays on perpetual.
	for watch := range triggers {
		trigger(watch._id)
	}
}

// Regularly look for Watch candidate for triggering.
// @I Search for candidate Watches in a connected Elastic Search database
// @I Support different sources of candidate Watches configurable via JSON or YAML
func search(watches chan<- Watch) {
	newWatches := loadWatches()
	for _, watch := range newWatches {
		watches <- watch
	}
	time.Sleep(5 * time.Second)
	search(watches)
}

// Filter candidate Watches to ensure that we don't run expired Watches or that we
// don't run them prematurely.
func filter(watch Watch, triggers chan<- Watch) {
	now := time.Now()
	afterStart := watch.start == nil || now.After(*watch.start)
	beforeEnd  := watch.end   == nil || now.Before(*watch.end)
	if afterStart && beforeEnd {
		run(watch, triggers)

		// @I Set active status to true to avoid requeueing the same Watch
		// @I Investigate throttling architecture and implementation
	}
}

// Execute triggering of Watches by sending to the corresponding channel.
// @I Consider rescheduling of triggered Watches via the corresponding channel
func run(watch Watch, triggers chan<- Watch) {
	triggers <- watch
	time.Sleep(watch.interval)
	filter(watch, triggers)
}

// Trigger an Watch by making a call to the Trigger API.
func trigger(_id string) {
	url  := WatchApiBaseURL + "/v" + WatchApiVersion + "/"
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

// Temporary function that acts as a candidate Watches source.
func loadWatches() []Watch {
	watches := []Watch {
		Watch {
			_id      : "1",
			interval : 1 * time.Second,
			enabled  : true,
			active   : false,
		},
		Watch {
			_id      : "2",
			interval : 3 * time.Second,
			enabled  : true,
			active   : false,
		},
	}

	return watches
}

type Watch struct {
	// Unique Watch identifier.
	_id      string
	// The Watch should be triggered only between its start and end times.
	start    *time.Time
	end      *time.Time
	// How frequently the Watch should be triggered.
	interval time.Duration
	// Watches should not be triggered if they are disabled.
	enabled  bool
	// Whether the Watch is active in the scheduler queue. It can be used to
	// prevent loading and triggering the same candidate Watch more than once.
	active   bool
}
