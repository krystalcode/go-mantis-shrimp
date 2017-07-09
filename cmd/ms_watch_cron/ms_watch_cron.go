package main

import (
	// Utilities.
	"fmt"
	"time"

	// Internal dependencies.
	config "github.com/krystalcode/go-mantis-shrimp/cron/config"
	schedule "github.com/krystalcode/go-mantis-shrimp/cron/schedule"
	storage "github.com/krystalcode/go-mantis-shrimp/cron/storage"
	util "github.com/krystalcode/go-mantis-shrimp/util"
	sdk "github.com/krystalcode/go-mantis-shrimp/watches/sdk"
)

/**
 * Constants.
 */

// CronConfigFile holds the full path to the file containing the configuration
// for the Cron component.
const CronConfigFile = "/etc/mantis-shrimp/cron.config.json"

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
	// Load configuration.
	// @I Support providing configuration file for Cron component via cli options
	// @I Validate Cron component configuration when loading from JSON file
	var cronConfig config.Config
	err := util.ReadJSONFile(CronConfigFile, &cronConfig)
	if err != nil {
		panic(err)
	}

	// Load Schedules provided in the config, if we run on ephemeral storage mode.
	loadEphemeralSchedules(&cronConfig)

	// Channel that receives IDs of the Watches that are ready to be triggered.
	triggers := make(chan int)

	// Channel that receives Schedules that are candidate for triggering.
	schedules := make(chan schedule.Schedule)

	// Search for candidate Schedules; it could be from a variety of sources.
	go search(schedules, &cronConfig)

	// Listen to candidate Schedules and send them for execution as they come. We
	// do this in a goroutine so that we don't block the program yet.
	go func() {
		for schedule := range schedules {
			go run(schedule, triggers, &cronConfig)
		}
	}()

	// Configuration required by the Watch API SDK.
	// @I Load Watch API SDK configuration from file or command line
	sdkConfig := sdk.Config{
		cronConfig.WatchAPI.BaseURL,
		cronConfig.WatchAPI.Version,
	}

	// Listen for IDs of Watches that are ready for triggering, and trigger them
	// as they come. We keep the channel open and the program stays on perpetual.
	for watchID := range triggers {
		fmt.Printf("triggering Watch with ID \"%d\"\n", watchID)
		err := sdk.TriggerByID(watchID, sdkConfig)
		if err != nil {
			fmt.Println(err)
		}
	}
}

// search looks for Schedules that are candidate for triggering at regular
// intervals. It could be from a variety of sources, but for now we only
// implement search via the Cron component.
func search(schedules chan<- schedule.Schedule, cronConfig *config.Config) {
	// @I Support different sources of candidate Schedules configurable via JSON
	//    or YAML

	// Create Redis Storage.
	storage, err := storage.Create(cronConfig.Storage)
	if err != nil {
		panic(err)
	}

	// The duration of the search interval.
	interval, err := time.ParseDuration(cronConfig.SearchInterval)
	if err != nil {
		panic(err)
	}

	candidateSchedules, err := storage.Search(interval)
	if err != nil {
		panic(err)
	}

	for _, schedule := range candidateSchedules {
		schedules <- *schedule
	}

	// Repeat the search after the defined search interval.
	time.Sleep(interval)
	search(schedules, cronConfig)
}

// run sends the IDs of the Watches to the channel where they will be queued for
// triggering.
func run(schedule schedule.Schedule, triggers chan<- int, cronConfig *config.Config) {
	// @I Investigate throttling architecture and implementation

	watchesIDs := schedule.Do()

	// If there are Watches to trigger, it means that the Schedule was successful.
	// We update the current time to be the Schedule's last trigger time.
	if len(watchesIDs) > 0 {
		go func() {
			// Create Redis Storage.
			// @I Make Redis storage thread safe by using a connection pool instead of
			//    creating a separate Redis instance
			storage, err := storage.Create(cronConfig.Storage)
			if err != nil {
				panic(err)
			}

			// @I Update only the individual field instead of the full object.
			now := time.Now()
			schedule.Last = &now
			storage.Update(&schedule, false)
		}()
	}

	for _, ID := range watchesIDs {
		triggers <- ID
	}
}

// loadEphemeralSchedules checks if the storage engine is configured to run in
// "ephemeral" mode, and if so, it loads into it any Schedules contained in the
// configuration file.
func loadEphemeralSchedules(cronConfig *config.Config) {
	// @I Load init Schedules directly in Redis via a script so that services
	//    don't have to be restarted together
	mode, ok := cronConfig.Storage["mode"]
	if !ok || mode.(string) != "ephemeral" || cronConfig.Schedules == nil {
		return
	}

	storage, err := storage.Create(cronConfig.Storage)
	if err != nil {
		panic(err)
	}

	for _, schedule := range cronConfig.Schedules {
		_, err := storage.Create(&schedule)
		if err != nil {
			panic(err)
		}
	}
}
