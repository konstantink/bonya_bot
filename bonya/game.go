package main

import (
	"fmt"

	"github.com/tucnak/telebot"
)

// GameSettings structure to store some settings for the game
type GameSettings struct {
	// ChatID id of the chat where to send notifications about current game
	ChatID int64
	// Domain where game is running
	Domain string
	// GameID id of the game
	GameID int32
	// UserName active user that can login to engine and get level information
	UserName string
	// Password for the user
	Password string
}

func (gs GameSettings) String() string {
	return fmt.Sprintf("<GameSettings: %d %s %d %s>", gs.ChatID, gs.Domain, gs.GameID, gs.UserName)
}

//func (gs *GameSettings)

// NewGameSettings constructor for GameSettings
func NewGameSettings() *GameSettings {
	// return &GameSettings{0, "", 0, "", ""
	return &GameSettings{}
}

type DomainChecker struct {
	Channel  *chan MessageSender
	Settings *GameSettings
	Text     string
}

// getKeyboard returns keyboard with all available buttons.
// TODO: move all available to DB
func (gsc DomainChecker) getKeyboard() (keyboard [][]telebot.KeyboardButton) {
	var (
		kbQuest        = telebot.KeyboardButton{Text: "quest.ua", Data: "http://quest.ua"}
		kbKharkovQuest = telebot.KeyboardButton{Text: "kharkov.quest.ua", Data: "http://kharkov.quest.ua"}
		kbKharkovEn    = telebot.KeyboardButton{Text: "kharkov.en.cx", Data: "http://kharkov.en.cx"}
	)
	keyboard = [][]telebot.KeyboardButton{{kbKharkovQuest, kbKharkovEn}, {kbQuest}}
	return
}

func (gsc DomainChecker) String() string {
	return fmt.Sprintf("<DomainChecker %s>", "a")
}

func (gsc DomainChecker) Prepare() {
	*gsc.Channel <- NewTextInlineMessage(telebot.Chat{ID: gsc.Settings.ChatID}, gsc.Text, gsc.getKeyboard())
}

// Process checks that
func (gsc DomainChecker) Process(args ...interface{}) bool {
	if gsc.Settings.Domain != "" {
		*gsc.Channel <- NewTextMessage(telebot.Chat{ID: gsc.Settings.ChatID},
			fmt.Sprintf("%s", gsc.Settings.Domain), telebot.Message{})
		return true
	}
	return false
}

// GameChecker helps to check that all the settings are set correctly
// and if not then promt the user to set them
type GameChecker struct {
	Channel  *chan MessageSender
	Settings *GameSettings
	Text     string
}

func (gc GameChecker) String() string {
	return fmt.Sprintf("<GameChecker %p>", gc.Settings)
}

type GameSettingsCheckingMachine struct {
	CheckingMachine
}

func (fsm *GameSettingsCheckingMachine) ResetState() {
	// fsm.SetState(fsm.machine.rules[0].Origin())
}
