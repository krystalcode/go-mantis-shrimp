package main

import (
	// Utilities.
	"fmt"
	"time"

	// Internal dependencies.
	schedule "github.com/krystalcode/go-mantis-shrimp/cron/schedule"
	storage "github.com/krystalcode/go-mantis-shrimp/cron/storage"
	sdk "github.com/krystalcode/go-mantis-shrimp/watches/sdk"
)

/**
 * Constants.
 */

// @I Make the Watch API base url configurable

// WatchAPIBaseURL holds the base url where the Watch API should be contacted.
const WatchAPIBaseURL = "http://ms-watch-api:8888"

// WatchAPIVersion holds the version of the Watch API that client calls use.
const WatchAPIVersion = "1"

// CronSearchIntervalSeconds holds the frequency in seconds with which the cron
// will look for Schedules candidate for triggering.
// @I Make the search interval configurable
const CronSearchIntervalSeconds = 1

/**
 * Main program entry.
 *
 * The flow of the program is as follows. A goroutine searches for new Schedules
 * at regular intervals. It sends any discovered Schedules to a channel where
 * any final evaluation will be executed to make sure we don't trigger Schedules
 * that we shouldn't). The IDs of the Watches of Schedules that pass the
 * evaluation are sent to another channel that executes the triggering.
 *
 * @I Add a Cron API for accepting Schedule submissions
 */
func main() {
	// Channel that receives IDs of the Watches that are ready to be triggered.
	triggers := make(chan int)

	// Channel that receives Schedules that are candidate for triggering.
	schedules := make(chan schedule.Schedule)

	// Search for candidate Schedules; it could be from a variety of sources.
	go search(schedules)

	// Listen to candidate Schedules and send them for execution as they come. We
	// do this in a goroutine so that we don't block the program yet.
	go func() {
		for schedule := range schedules {
			go run(schedule, triggers)
		}
	}()

	// Configuration required by the Watch API SDK.
	// @I Load Watch API SDK configuration from file or command line
	config := sdk.Config{
		WatchAPIBaseURL,
		WatchAPIVersion,
	}

	// Listen for IDs of Watches that are ready for triggering, and trigger them
	// as they come. We keep the channel open and the program stays on perpetual.
	for watchID := range triggers {
		fmt.Printf("triggering Watch with ID \"%d\"\n", watchID)
		err := sdk.TriggerByID(watchID, config)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// search looks for Schedules that are candidate for triggering at regular
// intervals. It could be from a variety of sources, but for now we only
// implement search via the Cron component.
func search(schedules chan<- schedule.Schedule) {
	// @I Support different sources of candidate Schedules configurable via JSON
	//    or YAML

	// Create Redis Storage.
	config := map[string]string{
		"STORAGE_ENGINE":    "redis",
		"STORAGE_REDIS_DSN": "redis:6379",
	}
	storage, err := storage.Create(config)
	if err != nil {
		panic(err)
	}

	// The duration of the search interval.
	interval := CronSearchIntervalSeconds * time.Second

	candidateSchedules, err := storage.Search(interval)
	if err != nil {
		panic(err)
	}

	for _, schedule := range candidateSchedules {
		schedules <- *schedule
	}

	// Repeat the search after the defined search interval.
	time.Sleep(interval)
	search(schedules)
}

// run sends the IDs of the Watches to the channel where they will be queued for
// triggering.
func run(schedule schedule.Schedule, triggers chan<- int) {
	// @I Investigate throttling architecture and implementation

	watchesIDs := schedule.Do()

	// If there are Watches to trigger, it means that the Schedule was successful.
	// We update the current time to be the Schedule's last trigger time.
	if len(watchesIDs) > 0 {
		go func() {
			// Create Redis Storage.
			// @I Make Redis storage thread safe by using a connection pool instead of
			//    creating a separate Redis instance
			config := map[string]string{
				"STORAGE_ENGINE":    "redis",
				"STORAGE_REDIS_DSN": "redis:6379",
			}
			storage, err := storage.Create(config)
			if err != nil {
				panic(err)
			}

			// @I Update only the individual field instead of the full object.
			now := time.Now()
			schedule.Last = &now
			storage.Update(&schedule)
		}()
	}

	for _, ID := range watchesIDs {
		triggers <- ID
	}
}
