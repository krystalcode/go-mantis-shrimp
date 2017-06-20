/**
 * Tests for the Mailgun Message Action.
 */

package msActionMailgun

import (
	// Testing packages.
	"testing"

	// Mailgun.
	mailgun "gopkg.in/mailgun/mailgun-go.v1"

	// Utilities.
	"fmt"

	// Internal dependencies.
	common "github.com/krystalcode/go-mantis-shrimp/actions/common"
)

/**
 * Helper types and functions reused in various tests.
 */

// testAction generates an Action object with some defaults.
func testAction() Action {
	action := Action{
		common.ActionBase{
			"Test Message",
		},
		"example.com",
		"test-api-key",
		"test-public-api-key",
		"sender@example.com",
		"recipient@example.com",
		"Test email subject",
		"Test email body",
		nil,
	}
	return action
}

// MockMailgunClientSuccess simulates successfully sending an email message.
type MockMailgunClientSuccess struct{}

func (client MockMailgunClientSuccess) Send(m *mailgun.Message) (string, string, error) {
	return "Email message successfully send", "id.example.mailgun.com", nil
}

// MockMailgunClientFailure simulates failure while sending an email message.
type MockMailgunClientFailure struct{}

func (client MockMailgunClientFailure) Send(m *mailgun.Message) (string, string, error) {
	return "", "", fmt.Errorf("there has been an error while sending the email message")
}

/**
 * Tests.
 */

func TestMailgunMessage_Success(t *testing.T) {
	action := testAction()
	client := MockMailgunClientSuccess{}
	action.SetMailgunClient(client)
	action.Do()
}

// @I Write tests for failure in MailgunMessageAction after having Action.Do()
//    functions return errors
