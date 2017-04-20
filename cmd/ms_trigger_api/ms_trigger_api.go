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
	item := getItemById(_id)

	/**
   * @I Implement authentication of the caller
   * @I Does the _id need any escaping?
   * @I Retrieve the record from the database
   * @I Ensure the caller has the permissions to trigger evaluation of an item
   * @I Trigger evaluation of the item
   * @I Find a better term than "item"
   * @I Investigate whether we need our own response status codes
   */

	// Return a Not Found response if there is no item with such _id.
	if item == nil {
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

// Stub function for getting an item by its _id until we implement item storage.
func getItemById(_id string) *Item {
	var item = Item{
		_id: _id,
	}
	return &item
}

// Stub model items until we design the item structure.
type Item struct {
	_id string
}
