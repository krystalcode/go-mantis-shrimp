/**
 * Provides an API for triggering the execution of Watches.
 */

package main

import (
	// Utilities.
	"fmt"
	"net/http"

	// Gin.
	gin "gopkg.in/gin-gonic/gin.v1"

	// Internal dependencies.
	sdk "github.com/krystalcode/go-mantis-shrimp/actions/sdk"
	util "github.com/krystalcode/go-mantis-shrimp/util"
	common "github.com/krystalcode/go-mantis-shrimp/watches/common"
	storage "github.com/krystalcode/go-mantis-shrimp/watches/storage"
	wrapper "github.com/krystalcode/go-mantis-shrimp/watches/wrapper"
)

/**
 * Constants.
 */

// @I Make the Action API base url configurable

// ActionAPIBaseURL holds the base url where the Action API should be contacted.
const ActionAPIBaseURL = "http://ms-action-api:8888"

// ActionAPIVersion holds the version of the Action API that client calls use.
const ActionAPIVersion = "1"

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

	// Version 1 of the Watch API.
	v1 := router.Group("/v1")
	{
		// Create a new Watch.
		v1.POST("/", v1Create)

		// Trigger execution of the Watch via its ID.
		v1.POST("/:ids/trigger", v1Trigger)
	}

	/**
	 * @I Make the trigger API port configurable
	 */
	router.Run(":8888")
}

/**
 * Endpoint functions.
 */

// v1Create provides an endpoint that creates a new Watch based on the JSON object
// given in the request.
func v1Create(c *gin.Context) {
	/**
	 * @I Implement authentication of the caller
	 * @I Validate parameters per Watch type
	 * @I Ensure the caller has the permissions to create Watches
	 * @I Log errors and send a 500 response instead of panicking
	 * @I Implement creating and triggering a Watch in a single request
	 */

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
	_id, err := storage.Set(watch)
	if err != nil {
		panic(err)
	}

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H{
			"status": http.StatusOK,
			"_id":    _id,
		},
	)
}

// v1Trigger provides an endpoint that triggers execution of the Action given in
// the request by its ID, by making a call to the Action API.
func v1Trigger(c *gin.Context) {
	/**
	 * @I Implement authentication of the caller
	 * @I Does the _id need any escaping?
	 * @I Ensure the caller has the permissions to trigger evaluation of a Watch
	 * @I Investigate whether we need our own response status codes
	 */

	// The "ids" parameter is required. We allow for multiple comma-separated
	// string IDs, so we need to convert them to an array of integer IDs.
	// We want to make sure that the caller makes the request they want to without
	// mistakes, so we do not trigger any Watches if there is any error, even in
	// one of the IDs.
	// @I Refactor converting a comma-separated list of string IDs to an array of
	//    integer IDs into a utility function
	sIDs := c.Param("ids")
	aIDsInt, err := util.StringToArrayOfIntegers(sIDs, ",")
	if err != nil {
		c.JSON(
			http.StatusNotFound,
			gin.H{
				"status": http.StatusNotFound,
			},
		)
		return
	}

	// Get the Watches with the requested IDs from storage.
	storage := c.MustGet("storage").(storage.Storage)

	var watches []*common.Watch
	for iID, _ := range aIDsInt {
		watch, err := storage.Get(iID)
		if err != nil {
			// Return a Not Found response if there is no Watch with such ID.
			if watch == nil {
				c.JSON(
					http.StatusNotFound,
					gin.H{
						"status": http.StatusNotFound,
					},
				)
			}
			return
		}

		// We could trigger the Watch at this point, however we prefer to check
		// that all Watches exist first.
		watches = append(watches, watch)
	}

	// Trigger execution of the Watches.
	// We only need to acknowledge that the Watches were triggered; we don't have to
	// for the execution to finish as this can take time.
	for _, pointer := range watches {
		go func() {
			watch := *pointer
			actionsIds := watch.Do()
			if len(actionsIds) == 0 {
				return
			}

			sdkConfig := sdk.Config{
				BaseURL: ActionAPIBaseURL,
				Version: ActionAPIVersion,
			}
			// @I Trigger all Watch Actions in one request
			for _, actionID := range actionsIds {
				go func() {
					err := sdk.TriggerByID(actionID, sdkConfig)
					if err != nil {
						// @I Investigate log management strategy for all services
						fmt.Println(err)
					}
				}()
			}
		}()
	}

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
