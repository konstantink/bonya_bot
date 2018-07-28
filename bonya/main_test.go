package main

import (
	"errors"
	"testing"

	tb "github.com/tucnak/telebot"
)

type testBotSender struct {
	recipient tb.Recipient
	options   *tb.SendOptions
	text      string
	photo     *tb.Photo
	venue     *tb.Venue
}

// var _ package.BotSender = (*testBotSender)(nil)

func (tbs *testBotSender) SendMessage(recipient tb.Recipient, text string, options *tb.SendOptions) error {
	tbs.recipient = recipient
	tbs.text = text
	tbs.options = options
	if tbs.text == "Error" {
		return errors.New("Test error")
	}
	return nil
}

func (tbs *testBotSender) SendPhoto(recipient tb.Recipient, photo *tb.Photo, options *tb.SendOptions) error {
	return nil
}

func (tbs *testBotSender) SendVenue(recipient tb.Recipient, venue *tb.Venue, options *tb.SendOptions) error {
	return nil
}

type testRecipient struct {
	name string
}

func (tr testRecipient) Destination() string {
	return tr.name
}

///////////////////////////////////////////////////////////////////////////////////

func TestSendMessage(t *testing.T) {
	var (
		sender    = &testBotSender{}
		options   = &tb.SendOptions{ParseMode: tb.ModeMarkdown, DisableWebPagePreview: true}
		recipient = testRecipient{name: "Test Chat"}
		message   = TextMessage{
			Message: Message{
				Recipient: recipient,
				Options:   options},
			Text: "Test string"}
		errMessage = TextMessage{
			Message: Message{
				Recipient: recipient,
				Options:   options},
			Text: "Error"}
	)

	err := message.Send(sender)
	if err != nil {
		t.Errorf("Not expected errors, got %s", err)
	}
	if sender.recipient.Destination() != message.Recipient.Destination() {
		t.Errorf("Expected recipient \"%s\", got \"%s\"", message.Recipient.Destination(), sender.recipient.Destination())
	}

	err = errMessage.Send(sender)
	if err == nil {
		t.Errorf("Expected error %s, got nil", "Test error")
	}
	if sender.recipient.Destination() != message.Recipient.Destination() {
		t.Errorf("Expected recipient \"%s\", got \"%s\"", message.Recipient.Destination(), sender.recipient.Destination())
	}
}
