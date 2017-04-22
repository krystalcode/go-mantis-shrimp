/**
 * Provides an API for triggering the execution of Actions.
 */

package main

import (
	// Utilities.
	"encoding/json"
	"net/http"
	"strconv"

	// Gin.
	gin "gopkg.in/gin-gonic/gin.v1"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
	chat   "github.com/krystalcode/go-mantis-shrimp/actions/chat"
	storage "github.com/krystalcode/go-mantis-shrimp/actions/storage"
)


/**
 * Main program entry.
 */
func main() {
	router := gin.Default()

	// Make storage available to the controllers.
	// @I Load storage configuration from file or cli options
	config := map[string]string{
		"STORAGE_ENGINE": "redis",
		"STORAGE_REDIS_DSN": "redis:6379",
	}
	router.Use(Storage(config))

	// Version 1 of the Action API.
	v1 := router.Group("/v1")
	{
		// Create a new Action.
		v1.POST("/", v1Create)

		// Trigger execution of the action via its ID.
		v1.POST("/:_id/trigger", v1Trigger)
	}

	/**
   * @I Make the Action API port configurable
   */
	router.Run(":8888")
}


/**
 * Endpoint controllers.
 */

// Create an Action based on the given parameters.
func v1Create(c *gin.Context) {
	// The parameters are provided as a JSON object in the request. Bind it to an
	// object of the corresponding type.
	var JSONBody requestJSON_Create
	err := c.BindJSON(&JSONBody)
	if err != nil {
		panic(err)
	}
	if JSONBody.Type == "" || len(JSONBody.Action) == 0 {
		c.JSON(
			http.StatusBadRequest,
			gin.H {
				"status" : http.StatusBadRequest,
			},
		)
		return
	}

	// Get the Action as an object of the appropriate type.
	action := JSONBody.actionByType()
	err = json.Unmarshal(JSONBody.Action, &action)
	if err != nil {
		panic(err)
	}
	if action == nil {
		c.JSON(
			http.StatusBadRequest,
			gin.H {
				"status" : http.StatusBadRequest,
			},
		)
		return
	}

	// Store the Action.
	storage := c.MustGet("storage").(storage.Storage)
	_id := storage.Set(action)

	/**
   * @I Implement authentication of the caller
   * @I Validate parameters per action type
   * @I Ensure the caller has the permissions to create actions
   * @I Log errors and send a 500 response instead of panicking
   */

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H {
			"status" : http.StatusOK,
			"_id"    : _id,
		},
	)
}

// Trigger the Action given by its ID.
/**
 * @I Implement authentication of the caller
 * @I Does the _id need any escaping?
 * @I Retrieve the record from the database
 * @I Ensure the caller has the permissions to trigger actions
 * @I Trigger actions
 * @I Consider allowing the caller to pass on the actions as well for being
 *    able to avoid the extra database call
 * @I Investigate whether we need our own response status codes
 * @I Allow triggering multiple action ids in one request
 */
func v1Trigger(c *gin.Context) {
	// The _id parameter is required.
	_idString := c.Param("_id")
	if _idString == "" {
		c.JSON(
			http.StatusBadRequest,
			gin.H {
				"status" : http.StatusBadRequest,
			},
		)
		return
	}

	// IDs are stored in storage as integers.
	_id, err := strconv.Atoi(_idString)
	if err != nil {
		panic(err)
	}

	// Get the Action with the requested ID from storage.
	storage := c.MustGet("storage").(storage.Storage)
	action := storage.Get(_id)

	// Return a Not Found response if there is no action with such _id.
	if action == nil {
		c.JSON(
			http.StatusNotFound,
			gin.H {
				"status" : http.StatusNotFound,
			},
		)
		return
	}

	// Trigger execution of the action.
	// We only need to acknowledge that the Action was triggered; we don't have to
	// for the execution to finish as this can take time.
	go action.Do()

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H {
			"status" : http.StatusOK,
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


/**
 * Endpoint helper types/functions.
 */

// Struct for holding the request data for the Create endpoint.
type requestJSON_Create struct {
	Type   string `json="type"`
	Action json.RawMessage `json="action"`
}

// Get the right Action type based on the "type" parameter included in the
// request, so that we can properly convert the JSON parameters into an Action
// object.
func (requestData requestJSON_Create) actionByType() common.Action {
	switch requestData.Type {
	case "chat_message":
		var action chat.Action
		return &action
	}

	return nil
}
