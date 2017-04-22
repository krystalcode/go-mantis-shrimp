package main

import (
	// Utilities.
	"net/http"

	// Gin.
	gin "gopkg.in/gin-gonic/gin.v1"
)


/**
 * Main program entry.
 */
func main() {
	router := gin.Default()

	// Version 1 of the trigger API.
	v1 := router.Group("/v1")
	{
		v1.POST("/", v1Trigger)
	}

	/**
   * @I Make the trigger API port configurable
   */
	router.Run(":8888")
}

/**
 * Endpoint functions.
 */
func v1Trigger(c *gin.Context) {
	_id  := c.PostForm("_id")
	watch := getWatchById(_id)

	/**
   * @I Implement authentication of the caller
   * @I Does the _id need any escaping?
   * @I Retrieve the record from the database
   * @I Ensure the caller has the permissions to trigger evaluation of a Watch
   * @I Trigger evaluation of the Watch
   * @I Investigate whether we need our own response status codes
   */

	// Return a Not Found response if there is no Watch with such _id.
	if watch == nil {
		c.JSON(
			http.StatusNotFound,
			gin.H {
				"status" : http.StatusNotFound,
			},
		)
		return
	}

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H {
			"status" : http.StatusOK,
		},
	)
}

// Stub function for getting a Watch by its _id until we implement Watch storage.
func getWatchById(_id string) *Watch {
	var watch = Watch{
		_id: _id,
	}
	return &watch
}

// Stub model for Watches until we design the Watch structure.
type Watch struct {
	_id string
}
