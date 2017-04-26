package main

import (
	// Utilities.
	"fmt"
	"net/http"
	"strconv"

	// Gin.
	gin "gopkg.in/gin-gonic/gin.v1"

	// Internal dependencies.
	sdk "github.com/krystalcode/go-mantis-shrimp/actions/sdk"
	storage "github.com/krystalcode/go-mantis-shrimp/watches/storage"
	wrapper "github.com/krystalcode/go-mantis-shrimp/watches/wrapper"
)

/**
 * Constants.
 */

// @I Make the Action API base url configurable
const ActionApiBaseURL = "http://ms-action-api:8888"
const ActionApiVersion = "1"

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

	// Version 1 of the trigger API.
	v1 := router.Group("/v1")
	{
		// Create a new Watch.
		v1.POST("/", v1Create)

		// Trigger execution of the Watch via its ID.
		v1.POST("/:_id/trigger", v1Trigger)
	}

	/**
	 * @I Make the trigger API port configurable
	 */
	router.Run(":8888")
}

/**
 * Endpoint functions.
 */

// Create a new Watch.
/**
 * @I Implement authentication of the caller
 * @I Validate parameters per Watch type
 * @I Ensure the caller has the permissions to create Watches
 * @I Log errors and send a 500 response instead of panicking
 * @I Implement creating and triggering a Watch in a single request
 */
func v1Create(c *gin.Context) {
	// The parameters are provided as a JSON object in the request. Bind it to an
	// object of the corresponding type.
	var wrapper wrapper.WatchWrapper
	err := c.BindJSON(&wrapper)
	if err != nil {
		panic(err)
	}
	// @I Return 400 Bad Request if we are given no Watch type in a Create request
	if wrapper.Watch == nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": http.StatusBadRequest,
			},
		)
		return
	}

	// Get the Watch as an object of the appropriate type.
	watch := wrapper.Watch

	// Store the Watch.
	// @I Remove underscores from all _id variables
	storage := c.MustGet("storage").(storage.Storage)
	_id := storage.Set(watch)

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H{
			"status": http.StatusOK,
			"_id":    _id,
		},
	)
}

// Trigger the Watch given by its ID.
/**
 * @I Implement authentication of the caller
 * @I Does the _id need any escaping?
 * @I Ensure the caller has the permissions to trigger evaluation of a Watch
 * @I Investigate whether we need our own response status codes
 * @I Allow triggering multiple action ids in one request
 */
func v1Trigger(c *gin.Context) {
	// The _id parameter is required.
	_idString := c.Param("_id")
	if _idString == "" {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": http.StatusBadRequest,
			},
		)
		return
	}

	// IDs are stored in storage as integers.
	_id, err := strconv.Atoi(_idString)
	if err != nil {
		panic(err)
	}

	// Get the Watch with the requested ID from storage.
	storage := c.MustGet("storage").(storage.Storage)
	watch := storage.Get(_id)

	// Return a Not Found response if there is no Watch with such _id.
	if watch == nil {
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"status": http.StatusNotFound,
			},
		)
		return
	}

	// Trigger execution of the Watch.
	// We only need to acknowledge that the Watch was triggered; we don't have to
	// for the execution to finish as this can take time.
	go func() {
		actionsIds := watch.Do()
		if len(actionsIds) == 0 {
			return
		}

		sdkConfig := sdk.Config{
			BaseURL: ActionApiBaseURL,
			Version: ActionApiVersion,
		}
		// @I Trigger all Watch Actions in one request
		for _, actionId := range actionsIds {
			go func() {
				err := sdk.TriggerById(actionId, sdkConfig)
				if err != nil {
					// @I Investigate log management strategy for all services
					fmt.Println(err)
				}
			}()
		}
	}()

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H{
			"status": http.StatusOK,
		},
	)
}

/**
 * Middleware.
 */

// Middleware for making available the storate engine to the controllers.
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
