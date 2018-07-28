package main

import (
	"log"

	tb "github.com/tucnak/telebot"
)

// MessageSender interface that all message types should implement, so that
// each message type knows how it should be sent
type MessageSender interface {
	// Send function that should be implemented in all message types.
	// Returns error in case something goes wrong or nil
	Send(bot BotSender) error
}

// BotSender interface mostly for testing purposes but it defines the interface
// to messsenger API
type BotSender interface {
	// SendMessage function to send text messages to recipient (chat, user)
	SendMessage(recipient tb.Recipient, text string, options *tb.SendOptions) error
	// SendMessage function to send photo messages to recipient (chat, user)
	SendPhoto(recipient tb.Recipient, photo *tb.Photo, options *tb.SendOptions) error
	// SendVenue function to send location messages to recipient (chat, user)
	SendVenue(recipient tb.Recipient, venue *tb.Venue, options *tb.SendOptions) error
}

// Message structure that represents basic fields required to send message
type Message struct {
	//Recipient who should receive message
	Recipient tb.Recipient
	// Options some options required by messenger
	Options *tb.SendOptions
}

// TextMessage represents text message type
type TextMessage struct {
	Message

	// Text string to send, usually it is information about level, hint or bonus.
	// Also it is used for additional messages like time left for level or sectors to close
	Text string
}

// Send implementation of Sender interface for TextMessage type
func (tm TextMessage) Send(bot BotSender) error {
	log.Print("[INFO] Send message to chat")
	err := bot.SendMessage(tm.Recipient, tm.Text, tm.Options)
	if err != nil {
		log.Printf("ERROR: Cannot send message: %s", err)
	}
	return err
}

// NewTextMessage constructor for the TextMessage type
func NewTextMessage(recipient tb.Recipient, message string, replyTo tb.Message) *TextMessage {
	textMessage := new(TextMessage)
	textMessage.Options = &tb.SendOptions{ParseMode: tb.ModeMarkdown,
		DisableWebPagePreview: true,
		ReplyTo:               replyTo}
	textMessage.Recipient = recipient
	textMessage.Text = message
	return textMessage
}

// TextInlineMessage the same as TextMessage, but with inline keyboard
type TextInlineMessage struct {
	TextMessage
}

// NewTextInlineMessage constructor for the TextInlineMessage
func NewTextInlineMessage(recipient tb.Recipient, message string, keyboard [][]tb.KeyboardButton) *TextInlineMessage {
	textInlineMessage := new(TextInlineMessage)
	textInlineMessage.Options = &tb.SendOptions{
		ReplyMarkup: tb.ReplyMarkup{
			Selective:       true,
			ForceReply:      true,
			ResizeKeyboard:  true,
			OneTimeKeyboard: true,
			InlineKeyboard:  keyboard,
		}}
	textInlineMessage.Recipient = recipient
	textInlineMessage.Text = message
	return textInlineMessage
}

// PhotoMessage represents photo message type
type PhotoMessage struct {
	Message

	// Photo structure that is accepted by Telegram to send photo message. Photos are
	// extracted from levels, hints, bonuses
	Photo *tb.Photo
}

// Send implementation of Sender interface for PhotoMessage type
func (pm PhotoMessage) Send(bot BotSender) error {
	log.Print("Send photo to chat")
	err := bot.SendPhoto(pm.Recipient, pm.Photo, pm.Options)
	if err != nil {
		log.Printf("WARNING: Cannot send message: %s", err)
	}
	return err
}

// NewPhotoMessage constructor for the TextMessage type
func NewPhotoMessage(recipient tb.Recipient, photo *tb.Photo, replyTo tb.Message) *PhotoMessage {
	photoMessage := new(PhotoMessage)
	photoMessage.Options = &tb.SendOptions{ParseMode: tb.ModeMarkdown,
		DisableWebPagePreview: true,
		ReplyTo:               replyTo}
	photoMessage.Recipient = recipient
	photoMessage.Photo = photo
	return photoMessage
}

// LocationMessage represents location message type
type LocationMessage struct {
	Message

	// Location is extracted from level information, hints or bonuses
	Location *tb.Venue
}

// Send implementation of Sender interface for LocationMessage type
func (lm LocationMessage) Send(bot BotSender) error {
	log.Print("[INFO] Send location to chat")
	err := bot.SendVenue(lm.Recipient, lm.Location, lm.Options)
	if err != nil {
		log.Printf("WARNING: Cannot send message: %s", err)
	}
	return err
}

// NewLocationMessage constructor for the TextMessage type
func NewLocationMessage(recipient tb.Recipient, venue *tb.Venue, replyTo tb.Message) *LocationMessage {
	locationMessage := new(LocationMessage)
	locationMessage.Options = &tb.SendOptions{ParseMode: tb.ModeMarkdown,
		DisableWebPagePreview: true,
		ReplyTo:               replyTo}
	locationMessage.Recipient = recipient
	locationMessage.Location = venue
	return locationMessage
}
