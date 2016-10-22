package main


import (
	"github.com/kelseyhightower/envconfig"
	"github.com/tucnak/telebot"
	"time"
	"fmt"
	"log"
	"net/http"
	"container/list"
	"net/http/cookiejar"
)

//type BotCommand int8

type SendInfo struct {
	Recepient telebot.Recipient
	Text string
	Options *telebot.SendOptions
}

type PhotoInfo struct {
	Recepient telebot.Recipient
	Photo *telebot.Photo
	Options *telebot.SendOptions
}

//const (
//	InfoCommand BotCommand = 1 << iota
//	SetChatIdCommand
//	WatchCommand
//	StopWatchingCommand
//	TestHelpChange
//)

var (
	quit chan struct{}
	sendInfoChan chan *SendInfo
	photoInfoChan chan *PhotoInfo
	levelInfoChan chan *LevelInfo
)

// Helpers

func NewSendInfo(recepient telebot.Recipient, text string, options *telebot.SendOptions) *SendInfo {
	return &SendInfo{
		Recepient: recepient,
		Text: text,
		Options: options,
	}
}

func IsBotCommand(m *telebot.Message) bool {
	for _, entity := range m.Entities {
		if entity.Type == "bot_command" {
			return true
		}
	}
	return false
}

///////////////////////////////

func processLevel(recepient telebot.Recipient, en *EnAPI) {
	var (
		level *LevelResponse
		err error
		images []Image
		text, task string
	)

	for {
		level, err = en.GetLevelInfo()
		if err != nil {
			log.Println("Can't get level info:", err)
			en.Login()
			continue
		}
		break
	}
	en.CurrentLevel = level.Level
	task = ReplaceCoordinates(en.CurrentLevel.Tasks[0].TaskText)
	task, images = ReplaceImages(task)
	task = ReplaceCommonTags(task)

	go SendImageFromUrl(recepient, images)

	text = fmt.Sprintf(LevelInfoString,
		en.CurrentLevel.Number,
		en.CurrentLevel.Name,
		PrettyTimePrint(en.CurrentLevel.Timeout),
		PrettyTimePrint(en.CurrentLevel.TimeoutSecondsRemain),
		task)

	sendInfoChan <- NewSendInfo(recepient, text,
		&telebot.SendOptions{ParseMode: telebot.ModeMarkdown,
			DisableWebPagePreview: true})
}

func startWatching(en *EnAPI) {
	var (
		ticker *time.Ticker
	)

	ticker = time.NewTicker(2*time.Second)
	quit = make(chan struct{})
	go func() {
		for {
			select{
			case <-ticker.C:
				resp, err := en.GetLevelInfo()
				if err != nil {
					log.Println("Error:", err)
					continue
				}
				log.Println("Send level to channel")
				levelInfoChan <- resp.Level
				fmt.Println("After writing:", len(levelInfoChan))
			case <-quit:
				ticker.Stop()
			}
		}
	}()
}

func stopWatching() {
	close(quit)
}

func ProcessBotCommand(m *telebot.Message, en *EnAPI) {
	var (
		command []byte
		commandCode BotCommand
		ok bool
		ent *telebot.MessageEntity
	)

	// TODO: Change to helper function
	ent = &m.Entities[0]
	command = make([]byte, ent.Length-1, ent.Length-1)
	copy(command, m.Text[ent.Offset+1:ent.Length])
	if commandCode, ok = BotCommandDict[string(command)]; !ok {
		log.Println("Unknown command:", command)
	}

	switch commandCode {
	case InfoCommand:
		processLevel(m.Chat, en)
	case WatchCommand:
		startWatching(en)
	case StopWatchingCommand:
		stopWatching()
	}
}

func initChannels() {
	sendInfoChan = make(chan *SendInfo, 10)
	photoInfoChan = make(chan *PhotoInfo, 10)
	levelInfoChan = make(chan *LevelInfo, 10)
}

func main() {
	var (
		envConfig EnvConfig
		bot *telebot.Bot
		err error
		updates chan telebot.Message
		update telebot.Message
		en *EnAPI
	)

	err = envconfig.Process("bonya", &envConfig)
	if err != nil {
		log.Fatal("Can't read environment variables:", err)
	}

	bot, err = telebot.NewBot(envConfig.BotToken)
	if err != nil {
		log.Panic(err)
	}

	initChannels()

	go func() {
		for {
			select {
			case si := <-sendInfoChan:
				bot.SendMessage(si.Recepient, si.Text, si.Options)
			case pi := <-photoInfoChan:
				bot.SendPhoto(pi.Recepient, pi.Photo, pi.Options)
			case <-levelInfoChan:

			}
		}
	}()

	jar, _ := cookiejar.New(nil)
	en = &EnAPI{
		login: envConfig.User,
		password: envConfig.Password,
		Client: &http.Client{Jar: jar},
		CurrentGameId: 25733,
		CurrentLevel: nil,
		Levels: list.New()}
	en.Login()

	log.Printf("Authorized on account %s", bot.Identity.Username)
	updates = make(chan telebot.Message, 50)
	bot.Listen(updates, 30*time.Second)

	//for update := range updates {
	for {
		select{
		case update = <-updates:
			log.Println("Read updates from Telegram")
			if update.Text != "" {
				if IsBotCommand(&update) {
					go ProcessBotCommand(&update, en)
				}

				log.Printf("[%s] %s", update.Sender.Username, update.Text)
			}
			//bot.SendMessage(update.Chat, fmt.Sprintf("Dear %s, I can't understand you", update.Sender.Username),
			//	&telebot.SendOptions{ReplyTo: update, ParseMode: telebot.ModeMarkdown})
		}
	}
}

	// Send photo
	//file, _ := telebot.NewFile("/tmp/Screen32Shot322016-09-0632at3216.47.23.png")
	//bot.SendPhoto(update.Chat, &telebot.Photo{File: file,
	//Thumbnail: telebot.Thumbnail{File: file, Width: 120, Height: 120}, Caption: "Photo"}, nil)

