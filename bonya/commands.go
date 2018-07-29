package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bonya_bot/en"
	tb "github.com/tucnak/telebot"
)

// BaseCommand is a base struct for all user-defined command handlers
type BaseCommand struct {
	output  chan MessageSender
	message tb.Message
	level   *en.Level
}

// Command is an interface that all user-defined command handlers should
// implement
type Command interface {
	Process(args ...string)
}

// UnknownCommand is a default handler in case user entered unregistered command.
// Used to send text message about incorrect command
type UnknownCommand struct {
	BaseCommand
}

// Process is required to implement Command interface
func (uc UnknownCommand) Process(args ...string) {
	if DEBUG {
		log.Printf("UnknownCommand is executed")
	}

	outputMessage := NewTextMessage(
		uc.message.Chat,
		"Unknown command",
		uc.message,
	)

	uc.output <- outputMessage
}

// NewUnknownCommand - constructor for the InfoCommand
func NewUnknownCommand(output chan MessageSender, message tb.Message, level *en.Level) (Command, error) {
	return UnknownCommand{BaseCommand{output, message, level}}, nil
}

// var CommandRegister = make(map[string]Command)

// func createCommandHandler(command string, commandsStore CommandStore) (Command, error) {
// 	// command, arguments := extractCommandAndArguments(message)
// 	commandHandler := commandsStore.Get(command)
// 	return commandHandler()
// }

// InfoCommand command used to send information about current level (or some specific, if
// argument is provided) to the chat
type InfoCommand struct {
	BaseCommand
}

// Process is required to implement Command interface
func (ic InfoCommand) Process(args ...string) {
	var (
		taskText string
		messages []string
	)
	if DEBUG {
		log.Printf("InfoCommand is executed")
	}

	messages = append(messages, ic.level.GetLevelDetails())

	taskText = ic.level.GetLevelTask()
	for _, coordinate := range ic.level.Coords {
		messages = append(messages, coordinate.String())
	}
	// TODO: use constant
	messages = append(messages, SplitText(taskText, 4096)...)

	for _, message := range messages {
		ic.output <- NewTextMessage(
			ic.message.Chat,
			message,
			tb.Message{},
		)
		time.Sleep(2 * time.Millisecond)
	}
	for _, coordinate := range ic.level.Coords {
		ic.output <- NewLocationMessage(
			ic.message.Chat,
			&tb.Venue{
				Location: tb.Location{
					Latitude:  float32(coordinate.Lat),
					Longitude: float32(coordinate.Lon)},
				Title: coordinate.String()},
			tb.Message{},
		)
		time.Sleep(2 * time.Millisecond)
	}
	for _, image := range ic.level.Images {
		log.Printf("[INFO] Reading file %s", image.Filepath)
		file, _ := tb.NewFile(image.Filepath)
		thumbnail := tb.Thumbnail{
			File:   file,
			Width:  60,
			Height: 60,
		}
		ic.output <- NewPhotoMessage(
			ic.message.Chat,
			&tb.Photo{
				File:      file,
				Thumbnail: thumbnail,
				Caption:   image.Caption,
			},
			tb.Message{},
		)
		time.Sleep(2 * time.Millisecond)
	}
}

// NewInfoCommand - constructor for the InfoCommand
func NewInfoCommand(output chan MessageSender, message tb.Message, level *en.Level) (Command, error) {
	return InfoCommand{BaseCommand{output, message, level}}, nil
}

// StartCommand handler for 'start' command, that is used for basic bot configuration
// for the chat from where it was called
type StartCommand struct {
	BaseCommand
}

// Process is required to implement Command interface
func (sc StartCommand) Process(args ...string) {
	var (
		buttons = make([][]tb.KeyboardButton, len(QuestDomains)/ButtonsPerRow+len(QuestDomains)%ButtonsPerRow)
		i       int8
	)
	for domain := range QuestDomains {
		if len(buttons[i]) == ButtonsPerRow {
			i++
		}
		buttons[i] = append(buttons[i], tb.KeyboardButton{Text: domain, Data: QuestDomains[domain]})
	}
	if DEBUG {
		log.Printf("[INFO] buttons %d %s", len(buttons), buttons)
	}
	message := NewTextInlineMessage(sc.message.Chat, "Выберите домен:", buttons)

	sc.output <- message
}

// NewStartCommand - constructor for the StartCommand
func NewStartCommand(output chan MessageSender, message tb.Message, level *en.Level) (Command, error) {
	return StartCommand{BaseCommand{output, message, level}}, nil
}

// CommandFactory factory type that is stored in the CommandStore. User-defined commands
// should implement this method
type CommandFactory func(chan MessageSender, tb.Message, *en.Level) (Command, error)

// CommandStore structure to store user-defined command factories
type CommandStore struct {
	*sync.RWMutex
	register map[string]CommandFactory
}

// NewCommandStore creates a new store and return a reference to it
func NewCommandStore() *CommandStore {
	return &CommandStore{
		RWMutex:  &sync.RWMutex{},
		register: make(map[string]CommandFactory),
	}
}

// Get tries to find the command handler in store for the provided key. If command
// is not found then returns error that command is not registered in store
func (cr CommandStore) Get(key string) (CommandFactory, error) {
	cr.RLock()
	defer cr.RUnlock()
	handler, exist := cr.register[key]
	if exist {
		return handler, nil
	}

	return NewUnknownCommand, fmt.Errorf("Command '%s' is not registered in store", key)
}

// Register is used to add user-defined command factory to the store
func (cr CommandStore) Register(command string, commandFactory CommandFactory) {
	cr.Lock()
	defer cr.Unlock()
	cr.register[command] = commandFactory
}

func (cr CommandStore) init() {
	cr.Register("info", NewInfoCommand)
	cr.Register("start", NewStartCommand)
}
