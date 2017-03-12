package main

import (
	"container/list"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/tucnak/telebot"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

//type BotCommand int8

type SendInfo struct {
	Recepient telebot.Recipient
	Text      string
	Options   *telebot.SendOptions
}

type PhotoInfo struct {
	Recepient telebot.Recipient
	Photo     *telebot.Photo
	Options   *telebot.SendOptions
}

//const (
//	InfoCommand BotCommand = 1 << iota
//	SetChatIdCommand
//	WatchCommand
//	StopWatchingCommand
//	TestHelpChange
//)

var (
	quit                  chan struct{}
	sendInfoChan          chan *SendInfo
	photoInfoChan         chan *PhotoInfo
	levelInfoChan         chan LevelInfo
	levelChangeChan       chan LevelInfo
	helpChangeChan        chan HelpInfo
	sectorChangeChan      chan ExtendedSectorInfo
	mixedActionChangeChan chan MixedActionInfo

	mainChat telebot.Chat
)

// Helpers

func NewSendInfo(recepient telebot.Recipient, text string, options *telebot.SendOptions) *SendInfo {
	return &SendInfo{
		Recepient: recepient,
		Text:      text,
		Options:   options,
	}
}

func IsBotCommand(m *telebot.Message) bool {
	for _, entity := range m.Entities {
		fmt.Println(entity)
		if entity.Type == "bot_command" {
			return true
		}
	}
	return false
}

///////////////////////////////

func processLevel(recepient telebot.Recipient, en *EnAPI) {
	var (
		level      *LevelResponse
		err        error
		images     []Image
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

	ticker = time.NewTicker(2 * time.Second)
	quit = make(chan struct{})
	go func() {
		defer func() {
			if p := recover(); p != nil {
				log.Println(fmt.Errorf("[startWatching/anonymous] внутренняя ошибка: %v", p))
			}
		}()

		for {
			select {
			case <-ticker.C:
				go func() {
					resp, err := en.GetLevelInfo()
					if err != nil {
						log.Println("Error:", err)
						return
					}
					log.Println("Send level to channel")
					levelInfoChan <- *resp.Level
				}()
			case <-quit:
				ticker.Stop()
			}
		}
	}()
}

func stopWatching() {
	close(quit)
}

func setChat(chat telebot.Chat) {
	var mu sync.RWMutex

	mu.Lock()
	defer mu.Unlock()

	mainChat = chat
}

func sendCode(en *EnAPI, codes []string) {
	defer func() {
		if p := recover(); p != nil {
			log.Println(fmt.Errorf("[sendCode] внутренняя ошибка: %v", p))
		}
	}()

	for _, code := range codes {
		log.Printf("Sending code %q to EN engine", code)
		resp, err := en.SendCode(code)
		if err != nil {
			log.Println("Failed to send code:", err)
		}
		log.Println(resp.Level.MixedActions[0].Answer)
	}
}

func ProcessBotCommand(m *telebot.Message, en *EnAPI) {
	var (
		command     []byte
		commandCode BotCommand
		ok          bool
		ent         *telebot.MessageEntity
	)

	// TODO: Change to helper function
	ent = &m.Entities[0]
	command = make([]byte, ent.Length-1, ent.Length-1)
	copy(command, m.Text[ent.Offset+1:ent.Length])
	if commandCode, ok = BotCommandDict[string(command)]; !ok {
		log.Println("Unknown command:", command)
	}
	log.Println("Command:", string(command))

	switch commandCode {
	case InfoCommand:
		processLevel(m.Chat, en)
	case WatchCommand:
		startWatching(en)
	case StopWatchingCommand:
		stopWatching()
	case SetChatIdCommand:
		setChat(m.Chat)
	case CodeCommand:
		messageTail := make([]byte, len(m.Text)-ent.Length-1, len(m.Text)-ent.Length-1)
		re := regexp.MustCompile("\\s*,\\s*")
		copy(messageTail, m.Text[ent.Length+1:])
		go sendCode(en, re.Split(strings.ToLower(string(messageTail)), -1))
	}
}

func CheckHelps(oldLevel LevelInfo, newLevel LevelInfo) {
	log.Println("Check helps state")
	for i, _ := range oldLevel.Helps {
		if oldLevel.Helps[i].Number == newLevel.Helps[i].Number {
			if oldLevel.Helps[i].HelpText != newLevel.Helps[i].HelpText {
				log.Println("New hint is available")
				helpChangeChan <- newLevel.Helps[i]
			}
		}
	}
	log.Println("Finish checking changes in Helps section")
}

func CheckSectors(oldLevel LevelInfo, newLevel LevelInfo) {
	log.Println("Start checking changes in Sectors section")
	for i, _ := range oldLevel.Sectors {
		if oldLevel.Sectors[i].Name == newLevel.Sectors[i].Name {
			if oldLevel.Sectors[i].IsAnswered != newLevel.Sectors[i].IsAnswered {
				log.Println("Sector is closed")
				sectorChangeChan <- ExtendedSectorInfo{
					sectorInfo:    &newLevel.Sectors[i],
					sectorsLeft:   newLevel.SectorsLeftToClose,
					sectorsPassed: newLevel.PassedSectorsCount,
					totalSectors:  int8(len(newLevel.Sectors))}
			}
		}
	}
	log.Println("Finish checking changes in Sectors section")
}

func CheckMixedActions(oldLevel LevelInfo, newLevel LevelInfo) {
	log.Println("Start checking changes in MixedActions section")
	sort.Sort(newLevel.MixedActions)
	fmt.Println(len(newLevel.MixedActions))
	if len(newLevel.MixedActions) > 0 {
		if len(oldLevel.MixedActions) == 0 {
			for _, item := range newLevel.MixedActions {
				mixedActionChangeChan <- item
			}
		} else {
			lastActionId := oldLevel.MixedActions[0].ActionId
			for _, item := range newLevel.MixedActions {
				fmt.Println(item.ActionId, item.Answer, lastActionId)
				if item.ActionId == lastActionId {
					break
				}
				mixedActionChangeChan <- item
			}
		}
	}
	//if len(oldLevel.MixedActions) < len(newLevel.MixedActions) {
	//	for i := len(oldLevel.MixedActions); i < len(newLevel.MixedActions); i++ {
	//		mixedActionChangeChan <- newLevel.MixedActions[i]
	//	}
	//}
	log.Println("Finish checking changes in MixedActions section")
}

func initChannels() {
	sendInfoChan = make(chan *SendInfo, 10)
	photoInfoChan = make(chan *PhotoInfo, 10)
	levelInfoChan = make(chan LevelInfo, 10)
	levelChangeChan = make(chan LevelInfo, 10)
	helpChangeChan = make(chan HelpInfo, 10)
	sectorChangeChan = make(chan ExtendedSectorInfo, 10)
	mixedActionChangeChan = make(chan MixedActionInfo, 10)
}

func main_() {
	bot, err := telebot.NewBot(os.Getenv("BONYA_BOT_TOKEN"))
	fmt.Println(os.Getenv("BONYA_BOT_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	messages := make(chan telebot.Message, 100)
	bot.Listen(messages, 1*time.Second)

	for message := range messages {
		if message.Text == "/hi" {
			bot.SendMessage(message.Chat,
				"Hello, "+message.Sender.FirstName+"!", nil)
		}
	}
}

func main() {
	var (
		envConfig EnvConfig
		bot       *telebot.Bot
		err       error
		updates   chan telebot.Message
		update    telebot.Message
		en        *EnAPI
	)

	defer func() {
		if p := recover(); p != nil {
			log.Println(fmt.Errorf("[main] внутренняя ошибка: %v", p))
		}
	}()

	err = envconfig.Process("bonya", &envConfig)
	failOnError(err, "Can't read environment variables")

	println(envConfig.BotToken)
	bot, err = telebot.NewBot(envConfig.BotToken)
	failOnError(err, "Can't connect to bot server")

	initChannels()

	go func() {
		defer func() {
			if p := recover(); p != nil {
				log.Println(fmt.Errorf("[main/anonymous] внутренняя ошибка: %v", p))
			}
		}()

		for {
			select {
			case si := <-sendInfoChan:
				bot.SendMessage(si.Recepient, si.Text, si.Options)
			case pi := <-photoInfoChan:
				bot.SendPhoto(pi.Recepient, pi.Photo, pi.Options)
				//		case li := <-levelInfoChan:
				//			log.Println("Receive level from channel")
				//			if isNewLevel(en.CurrentLevel, &li) {
				//				log.Println("Level is new")
				//				en.CurrentLevel = &li
				//				levelChangeChan <- li
				//				continue
				//			}
				//			CheckHelps(*en.CurrentLevel, li)
				//			CheckMixedActions(*en.CurrentLevel, li)
				//			CheckSectors(*en.CurrentLevel, li)
				//			en.CurrentLevel = &li
				//		case mai := <-mixedActionChangeChan:
				//			log.Println("New MixedAction is added")
				//			var text string
				//			if mai.IsCorrect {
				//				text = fmt.Sprintf(CorrectAnswerString, mai.Answer, mai.Login)
				//			} else {
				//				text = fmt.Sprintf(IncorrectAnswerString, mai.Answer, mai.Login)
				//			}
				//			sendInfoChan <- &SendInfo{Recepient: mainChat, Text: text, Options:
				//				&telebot.SendOptions{ParseMode: telebot.ModeMarkdown}}
			}
		}
	}()

	jar, _ := cookiejar.New(nil)
	en = &EnAPI{
		login:         envConfig.User,
		password:      envConfig.Password,
		Client:        &http.Client{Jar: jar},
		CurrentGameId: 25733,
		CurrentLevel:  nil,
		Levels:        list.New()}
	en.Login()

	log.Printf("Authorized on account %s", bot.Identity.Username)
	updates = make(chan telebot.Message, 50)
	bot.Listen(updates, 30*time.Second)

	////for update = range updates {
	for {
		select {
		case update = <-updates:
			log.Println("Read updates from Telegram")
			if update.Text != "" {
				if IsBotCommand(&update) {
					go ProcessBotCommand(&update, en)
				}

				log.Printf("[%s] %s", update.Sender.Username, update.Text)
			}
			bot.SendMessage(update.Chat, fmt.Sprintf("Dear %s, I can't understand you", update.Sender.Username),
				&telebot.SendOptions{ReplyTo: update, ParseMode: telebot.ModeMarkdown})
		}
	}

}

// Send photo
//file, _ := telebot.NewFile("/tmp/Screen32Shot322016-09-0632at3216.47.23.png")
//bot.SendPhoto(update.Chat, &telebot.Photo{File: file,
//Thumbnail: telebot.Thumbnail{File: file, Width: 120, Height: 120}, Caption: "Photo"}, nil)
