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
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
)

/**
 * Constants.
 */

// chatMessageActionTimeout defines the HTTP Client's timeout duration in
// seconds for all Chat Message Actions.
// @I Make the timeout for Chat Message actions configurable per Action
const chatMessageActionTimeout = 30

/**
 * Types and their methods.
 */

// HTTPClient is an interface that is used to allow dependency injection of the
// HTTP client that makes the request to the Action's URL. Dependency injection
// is necessary for testing purposes.
type HTTPClient interface {
	Post(string, string, io.Reader) (*http.Response, error)
}

// Action implements the common.Action interface. It provides an Action that
// posts a message to a chat application such as Rocket.Chat, Mattermost, Slack
// or HipChat. Rocket.Chat-style message payload via a webhook is supported at
// the moment.
type Action struct {
	// Common fields and functions for all Actions.
	common.ActionBase

	// Webhook where the message will be posted. Provided by the chat application.
	URL string `json:"url"`
	// The message that will be posted.
	Message Message `json:"message"`

	// The HTTP client used to make the request to the URL.
	httpClient HTTPClient
}

// Do Implements common.Action.Do().
// It executes the Chat Action by posting the message to the chat application.
func (action Action) Do() {
	// Convert the message to JSON.
	body, err := json.Marshal(action.Message)
	if err != nil {
		panic(err)
	}

	// Create and send the request.
	res, err := action.httpClient.Post(action.URL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		resBody, _ := ioutil.ReadAll(res.Body)
		fmt.Println("Status:", res.Status)
		fmt.Println("Headers:", res.Header)
		fmt.Println("Body:", string(resBody))
	}
}

// SetHTTPClient allows to inject an HTTP client into the corresponding field.
func (action *Action) SetHTTPClient(client HTTPClient) {
	action.httpClient = client
}

// Message holds a chat message. Implements the structure required by
// Rocket.Chat.
// @see https://rocket.chat/docs/developer-guides/rest-api/chat/postmessage
type Message struct {
	Text        *string       `json:"text,omitempty"`
	Alias       *string       `json:"alias,omitempty"`
	Emoji       *string       `json:"emoji,omitempty"`
	Avatar      *string       `json:"avatar,omitempty"`
	Attachments *[]Attachment `json:"attachments,omitempty"`
}

// Attachment holds chat message attachments. Implements the structure required
// by Rocket.Chat.
// @see https://rocket.chat/docs/developer-guides/rest-api/chat/postmessage
type Attachment struct {
	Color             *string  `json:"color,omitempty"`
	Text              *string  `json:"text,omitempty"`
	Ts                *string  `json:"ts,omitempty"`
	ThumbURL          *string  `json:"thumb_url,omitempty"`
	MessageLink       *string  `json:"message_link,omitempty"`
	Collapsed         *bool    `json:"collapsed,omitempty"`
	AuthorName        *string  `json:"author_name,omitempty"`
	AuthorLink        *string  `json:"author_link,omitempty"`
	AuthorIcon        *string  `json:"author_icon,omitempty"`
	Title             *string  `json:"title,omitempty"`
	TitleLink         *string  `json:"title_link,omitempty"`
	TitleLinkDownload *string  `json:"title_link_download,omitempty"`
	ImageURL          *string  `json:"image_url,omitempty"`
	AudioURL          *string  `json:"audio_url,omitempty"`
	Fields            *[]Field `json:"fields,omitempty"`
}

// Field holds chat message attachment fields. Implements the structure required
// by Rocket.Chat.
// @see https://rocket.chat/docs/developer-guides/rest-api/chat/postmessage
type Field struct {
	Short *bool   `json:"short,omitempty"`
	Title *string `json:"title,omitempty"`
	Value *string `json:"value,omitempty"`
}

// NewChatMessageAction implements the ActionFactory function type. It creates a
// Chat Message Action based on the given JSON-object, and initializes it by
// injecting the required HTTP client.
var NewChatMessageAction = func(jsonAction *json.RawMessage) (common.Action, error) {
	// Create an Action object from JSON.
	var action Action
	err := json.Unmarshal(*jsonAction, &action)
	if err != nil {
		return nil, err
	}

	// Inject an HTTP client with the Action's timeout.
	client := &http.Client{
		Timeout: chatMessageActionTimeout * time.Second,
	}
	action.SetHTTPClient(client)

	return action, nil
}
