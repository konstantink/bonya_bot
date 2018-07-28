package main

// BotCommand - int8 alias
type BotCommand int8

const (
	// InfoCommand - command 'info' shows information of the level
	// InfoCommand BotCommand = 1 + iota

	// SetChatIDCommand - command 'setchat' defines the chat where to send level updates
	SetChatIDCommand BotCommand = 2 + iota

	// WatchCommand - command 'watch' tells bot to monitor state of game
	WatchCommand

	// StopWatchingCommand - command 'stopwatching' tells to stop monitor game
	StopWatchingCommand

	// CodeCommand - command 'c' sends code to EN engine
	CodeCommand

	// CompositeCodeCommand - command 'cc' sends composite code to EN engine
	CompositeCodeCommand

	// SectorsLeftCommand - command 'sl' sends message with list of sectors left
	SectorsLeftCommand

	// TimeLeftCommand - command 'tl' sends message with time left for level
	TimeLeftCommand

	// ListHelpsCommand - command 'lh' sends message with helps
	ListHelpsCommand

	// HelpTimeCommand - command 'ht' sends message with time before help
	HelpTimeCommand

	// StartCommand - command 'start' starts the process of configuing bot to monitor game
	// StartCommand

	// TestHelpChange - [REMOVE] - command '' test command to send message with help
	TestHelpChange
)

// BotCommandDict - dictionary with all bot commands
var BotCommandDict = map[string]BotCommand{
	// "info":         InfoCommand,
	"setchat":      SetChatIDCommand,
	"watch":        WatchCommand,
	"stopwatching": StopWatchingCommand,
	"c":            CodeCommand,
	"с":            CodeCommand,
	"cc":           CompositeCodeCommand,
	"сс":           CompositeCodeCommand,
	"sl":           SectorsLeftCommand,
	"ос":           SectorsLeftCommand,
	"tl":           TimeLeftCommand,
	"ов":           TimeLeftCommand,
	"lh":           ListHelpsCommand,
	"ht":           HelpTimeCommand,
	// "start":        StartCommand,
	"helpchange": TestHelpChange}

// QuestDomains dictionary that stores the urls to different game engines
// Later this can be stored in the database and extended by talking to bot
var QuestDomains = map[string]string{
	"quest.ua":         "http://quest.ua",
	"kharkov.quest.ua": "http://kharkov.quest.ua",
	"kharkov.en.cx":    "http://kharkov.en.cx",
}

// ButtonsPerRow number of buttons that should be displayed in one row
const ButtonsPerRow = 2
