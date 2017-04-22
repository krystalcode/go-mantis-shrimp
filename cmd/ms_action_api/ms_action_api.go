/**
 * Provides an API for triggering the execution of Actions.
 */

package main

import (
	// Utilities.
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
 * Endpoint functions.
 */

// Create an Action based on the given parameters.
func v1Create(c *gin.Context) {
	storage := c.MustGet("storage").(storage.Storage)
	// Get a mock action until we implement constructing one from the parameters.
	action := getActionById("mock-id")
	_id := storage.Set(action)

	/**
   * @I Implement authentication of the caller
   * @I Accept and validate parameters per action type
   * @I Ensure the caller has the permissions to create actions
   * @I Check if we'd rather send a 500 response in case of errors
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

// Stub function for getting an action by its _id until we implement storage.
func getActionById(_id string) common.Action {
	text := "There is an update available for the chat."
	alias := "mantis-shrimp"
	emoji := ":smirk:"
	aTs := "2016-12-09T16:53:06.761Z"
	aText := "An attachment to the message"
	aAuthorName := "Mantis Shrimp"
	aAuthorLink := "https://github.com/krystalcode/go-mantis-shrimp"
	attachment := chat.Attachment {
		Ts : &aTs,
		Text : &aText,
		AuthorName : &aAuthorName,
		AuthorLink : &aAuthorLink,
	}
	attachments := []chat.Attachment {attachment}
	action := chat.Action {
		common.ActionBase { "1" },
		"http://chat:3000/hooks/kFQCMu8tSGfCQw9i4/YpzBK9Y3eDHCDBf4wABKTwNjkd7hrLtGofPNGXzYJjuX3rKq",
		chat.Message {
			Text : &text,
			Alias : &alias,
			Emoji : &emoji,
			Attachments : &attachments,
		},
	}
	return action
}
