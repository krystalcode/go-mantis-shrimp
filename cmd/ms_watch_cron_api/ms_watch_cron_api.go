package main

import (
	// Utilities.
	"net/http"

	// Gin
	gin "gopkg.in/gin-gonic/gin.v1"

	// Internal dependencies.
	config "github.com/krystalcode/go-mantis-shrimp/cron/config"
	schedule "github.com/krystalcode/go-mantis-shrimp/cron/schedule"
	storage "github.com/krystalcode/go-mantis-shrimp/cron/storage"
	util "github.com/krystalcode/go-mantis-shrimp/util"
)

/**
 * Constants.
 */

// CronConfigFile holds the default path to the file containing the
// configuration for the Cron component.
const CronConfigFile = "/etc/mantis-shrimp/cron.config.json"

/**
 * Main program entry.
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

	router := gin.Default()

	// Make storage available to the controllers.
	router.Use(Storage(cronConfig.Storage))

	// Version 1 of the Cron API.
	v1 := router.Group("/v1")
	{
		// Create a new Schedule.
		v1.POST("/", v1Create)
	}

	/**
	 * @I Make the trigger API port configurable
	 */
	router.Run(":8888")
}

/**
 * Endpoint functions.
 */

// v1Create provides an endpoint that creates a new Schedule based on the JSON
// object given in the request.
func v1Create(c *gin.Context) {
	/**
	 * @I Implement authentication of the caller
	 * @I Validate parameters
	 * @I Ensure the caller has the permissions to create Schedules
	 * @I Log errors and send a 500 response instead of panicking
	 */

	// The parameters are provided as a JSON object in the request. Bind it to an
	// object of the corresponding type.
	var schedule schedule.Schedule
	err := c.BindJSON(&schedule)
	if err != nil {
		panic(err)
	}

	// Store the Watch.
	storage := c.MustGet("storage").(storage.Storage)
	scheduleID, err := storage.Create(&schedule)
	if err != nil {
		panic(err)
	}

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H{
			"status": http.StatusOK,
			"id":     scheduleID,
		},
	)
}

/**
 * Middleware.
 */

// Storage is a Gin middleware that makes available the Storage engine to the
// endpoint controllers.
func Storage(config map[string]interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		storage, err := storage.Create(config)
		if err != nil {
			panic(err)
		}
		c.Set("storage", storage)
		c.Next()
	}
}
