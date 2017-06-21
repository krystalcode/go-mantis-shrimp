/**
 * Provides an action for sending an email message via Mailgun.
 */

package msActionMailgun

import (
	// Utilities.
	"encoding/json"

	// Mailgun.
	mailgun "gopkg.in/mailgun/mailgun-go.v1"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
)

/**
 * Types and their methods.
 */

// MailgunClient is an interface that is used to allow dependency injection of
// the client that makes the request to the Mailgun API. Dependency injection is
// necessary for testing purposes.
type MailgunClient interface {
	Send(m *mailgun.Message) (string, string, error)
}

// Action implements the common.Action interface. It provides an Action that
// sends an email message via Mailgun. The details of the message are hardcoded
// per action at the moment, but template capabilities will be provided in the
// future.
type Action struct {
	// Common fields and functions for all Actions.
	common.ActionBase

	// Mailgun configuration.
	MailgunDomain       string `json:"mailgun_domain"`
	MailgunAPIKey       string `json:"mailgun_api_key"`
	MailgunPublicAPIKey string `json:"mailgun_public_api_key"`

	// Message details.
	MessageFrom    string `json:"message_from"`
	MessageTo      string `json:"message_to"`
	MessageSubject string `json:"message_subject"`
	MessageBody    string `json:"message_body"`

	// The client used to make the request to the Mailgun API.
	mailgunClient MailgunClient
}

// Do Implements common.Action.Do().
// It executes the Mailgun Action by sending the email message via Mailgun.
func (action Action) Do() error {
	message := mailgun.NewMessage(
		action.MessageFrom,
		action.MessageSubject,
		action.MessageBody,
		action.MessageTo,
	)
	_, _, err := action.mailgunClient.Send(message)
	if err != nil {
		return err
	}

	return nil
}

// SetMailgunClient allows to inject a Mailgun client into the corresponding
// field.
func (action *Action) SetMailgunClient(client MailgunClient) {
	action.mailgunClient = client
}

// NewMailgunMessageAction implements the ActionFactory function type. It creates
// a Mailgun Message Action based on the given JSON-object, and initializes it
// by injecting the required Mailgun client.
var NewMailgunMessageAction = func(jsonAction *json.RawMessage) (common.Action, error) {
	// Create an Action object from JSON.
	var action Action
	err := json.Unmarshal(*jsonAction, &action)
	if err != nil {
		return nil, err
	}

	// Inject a Mailgun client with the Action's timeout.
	client := mailgun.NewMailgun(
		action.MailgunDomain,
		action.MailgunAPIKey,
		action.MailgunPublicAPIKey,
	)
	action.SetMailgunClient(client)

	return action, nil
}
