/**
 * Provides an API for triggering the execution of Actions.
 */

package main

import (
	// Utilities.
	"net/http"
	"strconv"
	"strings"

	// Gin.
	gin "gopkg.in/gin-gonic/gin.v1"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
	storage "github.com/krystalcode/go-mantis-shrimp/actions/storage"
	wrapper "github.com/krystalcode/go-mantis-shrimp/actions/wrapper"
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

	// Version 1 of the Action API.
	v1 := router.Group("/v1")
	{
		// Create a new Action.
		v1.POST("/", v1Create)

		// Trigger execution of the action via its ID.
		v1.POST("/:ids/trigger", v1Trigger)
	}

	/**
	 * @I Make the Action API port configurable
	 */
	router.Run(":8888")
}

/**
 * Endpoint controllers.
 */

// v1Create provides an endpoint that creates a new Action based on the JSON
// object given in the request.
func v1Create(c *gin.Context) {
	/**
	 * @I Implement authentication of the caller
	 * @I Validate parameters per Action type
	 * @I Ensure the caller has the permissions to create Actions
	 * @I Log errors and send a 500 response instead of panicking
	 * @I Implement creating and triggering an Action in a single request
	 */

	// The parameters are provided as a JSON object in the request. Bind it to an
	// object of the corresponding type.
	var wrapper wrapper.ActionWrapper
	err := c.BindJSON(&wrapper)
	if err != nil {
		panic(err)
	}
	// @I Return 400 Bad Request if we are given no Action type in a Create request
	if wrapper.Action == nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H{
				"status": http.StatusBadRequest,
			},
		)
		return
	}

	// Get the Action as an object of the appropriate type.
	action := wrapper.Action

	// Store the Action.
	storage := c.MustGet("storage").(storage.Storage)
	_id := storage.Set(action)

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H{
			"status": http.StatusOK,
			"_id":    _id,
		},
	)
}

// v1Trigger provides an endpoint that triggers the Actions given in the request
// by their ID.
func v1Trigger(c *gin.Context) {
	/**
	 * @I Implement authentication of the caller
	 * @I Does the _id need any escaping?
	 * @I Ensure the caller has the permissions to trigger actions
	 * @I Consider allowing the caller to pass on the actions as well for being
	 *    able to avoid the extra database call
	 * @I Investigate whether we need our own response status codes
	 */

	// The "ids" parameter is required. We allow for multiple comma-separated
	// string IDs, so we need to convert them to an array of integer IDs.
	// We want to make sure that the caller makes the request they want to without
	// mistakes, so we do not trigger any Actions if there is any error, even in
	// one of the IDs.
	sIDs := c.Param("ids")
	aIDsString := strings.Split(sIDs, ",")

	aIDsInt := make(map[int]struct{})
	for _, sID := range aIDsString {
		iID, err := strconv.Atoi(sID)
		if err != nil {
			c.JSON(
				http.StatusNotFound,
				gin.H{
					"status": http.StatusNotFound,
				},
			)
			return
		}

		if _, ok := aIDsInt[iID]; !ok {
			aIDsInt[iID] = struct{}{}
		}
	}

	// Get the Actions with the requested IDs from storage.
	storage := c.MustGet("storage").(storage.Storage)

	var actions []*common.Action
	for iID, _ := range aIDsInt {
		action, err := storage.Get(iID)
		if err != nil {
			panic(err)
		}

		// Return a Not Found response if there is no Action with such ID.
		if action == nil {
			c.JSON(
				http.StatusNotFound,
				gin.H{
					"status": http.StatusNotFound,
				},
			)
			return
		}

		// We could trigger the Action at this point, however we prefer to check
		// that all Actions exist first.
		actions = append(actions, action)
	}

	// Trigger executions of the Actions.
	// We only need to acknowledge that the Actions were triggered; we don't have
	// to for the execution to finish as this can take time.
	for _, pointer := range actions {
		go func() {
			action := *pointer
			action.Do()
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
