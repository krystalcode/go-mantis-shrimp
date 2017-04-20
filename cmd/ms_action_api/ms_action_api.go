/**
 * Provides an API for triggering the execution of Actions.
 */

package main

import (
	// Utilities.
	"net/http"

	// Gin.
	gin "gopkg.in/gin-gonic/gin.v1"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
	chat   "github.com/krystalcode/go-mantis-shrimp/actions/chat"
)


/**
 * Main program entry.
 */
func main() {
	router := gin.Default()

	// Version 1 of the Action API.
	v1 := router.Group("/v1")
	{
		// Trigger execution of the action via its ID.
		v1.POST("/:_id", v1Action)
	}

	/**
   * @I Make the Action API port configurable
   */
	router.Run(":8888")
}


/**
 * Endpoint functions.
 */

func v1Action(c *gin.Context) {
	_id    := c.PostForm("_id")
	action := getActionById(_id)

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
	action.Do()

	// All good.
	c.JSON(
		http.StatusOK,
		gin.H {
			"status" : http.StatusOK,
		},
	)
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
