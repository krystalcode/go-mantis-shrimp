package main

import (
	// Utilities.
	"net/http"

	// Gin
	gin "gopkg.in/gin-gonic/gin.v1"

	// Internal dependencies.
	schedule "github.com/krystalcode/go-mantis-shrimp/cron/schedule"
	storage "github.com/krystalcode/go-mantis-shrimp/cron/storage"
)

/**
 * Main program entry.
 */
func main() {
	router := gin.Default()

	// Make storage available to the controllers.
	// @I Load storage configuration from file or cli options
	config := map[string]string{
		"STORAGE_ENGINE":    "redis",
		"STORAGE_REDIS_DSN": "redis:6379",
	}
	router.Use(Storage(config))

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
	scheduleID, err := storage.Set(&schedule)
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
func Storage(config map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		storage, err := storage.Create(config)
		if err != nil {
			panic(err)
		}
		c.Set("storage", storage)
		c.Next()
	}
}
