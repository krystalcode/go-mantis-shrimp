/**
 * Provides an action for posting a message to a chat application. Currently
 * supporting posting to Rocket Chat via a webhook.
 *
 * @I Architect a generic chat plugin that supports Slack and HipChat as well
 */

package msActionChat

import (
	// Utilities.
	"bytes"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
)

/**
 * Types and their methods.
 */

// Action for posting a message to the chat application.
type Action struct {
	// Common fields for all actions.
	common.ActionBase
	// Webhook where the message will be posted. Provided by the chat application.
	URL     string
	// The message that will be posted.
	Message Message
}

// Implements Action interface.
// Executes the action by posting the message to the chat application.
func (action Action) Do() {
	url  := action.URL

	// Convert the message to JSON.
	body, err := json.Marshal(action.Message)
	if err != nil {
		panic(err)
	}

	// Create and send the request.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	// Extract status as integer from "200 OK"
	resStatus, err := strconv.Atoi(res.Status[0:3])
	if err != nil {
		panic(err)
	}

	if resStatus != http.StatusOK {
		resBody, _ := ioutil.ReadAll(res.Body)
		fmt.Println("Status:"  , res.Status)
		fmt.Println("Headers:" , res.Header)
		fmt.Println("Body:"    , string(resBody))
	}
}

// Holds a message. Implements the structure required by Rocket Chat
// https://rocket.chat/docs/developer-guides/rest-api/chat/postmessage
type Message struct {
	Text        *string `json:"text,omitempty"`
	Alias       *string `json:"alias,omitempty"`
	Emoji       *string `json:"emoji,omitempty"`
	Avatar      *string `json:"avatar,omitempty"`
	Attachments *[]Attachment `json:"attachments,omitempty"`
}

type Attachment struct {
	Color             *string  `json:"color,omitempty"`
	Text              *string  `json:"text,omitempty"`
	Ts                *string  `json:"ts,omitempty"`
	ThumbUrl          *string  `json:"thumb_url,omitempty"`
	MessageLink       *string  `json:"message_link,omitempty"`
	Collapsed         *bool    `json:"collapsed,omitempty"`
	AuthorName        *string  `json:"author_name,omitempty"`
	AuthorLink        *string  `json:"author_link,omitempty"`
	AuthorIcon        *string  `json:"author_icon,omitempty"`
	Title             *string  `json:"title,omitempty"`
	TitleLink         *string  `json:"title_link,omitempty"`
	TitleLinkDownload *string  `json:title_link_download,omitempty`
	ImageURL          *string  `json:"image_url,omitempty"`
	AudioURL          *string  `json:"audio_url,omitempty"`
	Fields            *[]Field `json:"fields,omitempty"`
}

type Field struct {
	Short *bool   `json:"short,omitempty"`
	Title *string `json:"title,omitempty"`
	Value *string `json:"value,omitempty"`
}
