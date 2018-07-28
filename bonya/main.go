package main

import (
	"container/list"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bonya_bot/en"
	"github.com/kelseyhightower/envconfig"
	tb "github.com/tucnak/telebot"
)

// DEBUG flag to dump some additional info
var DEBUG = true

//type BotCommand int8

// type SendInfo struct {
// 	Recepient telebot.Recipient
// 	Text      string
// 	Options   *telebot.SendOptions
// }
//
// type PhotoInfo struct {
// 	Recepient telebot.Recipient
// 	Photo     *telebot.Photo
// 	Options   *telebot.SendOptions
// }
//
// type CoordInfo struct {
// 	Recepient telebot.Recipient
// 	Location  *telebot.Venue
// 	Options   *telebot.SendOptions
// }

var (
	quit chan struct{}

	// sendInfoChan   chan en.ToChat
	// photoInfoChan  chan *PhotoInfo
	// coordsInfoChan chan *CoordInfo
	levelInfoChan chan *en.Level
	messageChan   chan MessageSender

	mainChat tb.Chat
)

// Helpers

func SendImageFromUrl(recipient tb.Recipient, images en.Images) {
	var (
		file *os.File
	)
	for _, img := range images {

		resp, err := http.Get(img.URL)
		if err != nil {
			log.Println("Can't download image:", err)
		}

		filename := fmt.Sprintf("/tmp/%s", path.Base(img.URL))
		fileInfo, err := os.Stat(filename)
		if err == nil && fileInfo.Size() > 0 {
			file, err = os.Open(filename)
		} else {
			log.Println("Image is not downloaded yet:", err)
			file, err = os.Create(filename)
			if err != nil {
				log.Fatal("Cannot create file:", err)
				return
			}
			// Use io.Copy to just dump the response body to the file. This supports huge files
			_, err = io.Copy(file, resp.Body)
			resp.Body.Close()
			file.Close()
			if err != nil {
				log.Fatal(err)
			}
		}
		telebotFile, _ := tb.NewFile(file.Name())
		log.Printf("Sending photo to the channel (%s)", file.Name())
		thumbnail := tb.Thumbnail{File: telebotFile, Width: 120, Height: 120}
		messageChan <- PhotoMessage{Message{Recipient: recipient, Options: nil},
			&tb.Photo{File: telebotFile, Thumbnail: thumbnail, Caption: img.Caption}}
		// photoInfoChan <- &PhotoInfo{Recepient: recepient, Photo: &telebot.Photo{File: telebotFile,
		// 	Thumbnail: thumbnail, Caption: img.Caption}, Options: nil}
	}
}

func SendCoords(recipient tb.Recipient, coords en.Coordinates) {
	for _, coord := range coords {
		messageChan <- LocationMessage{Message{recipient, nil},
			&tb.Venue{Location: tb.Location{Latitude: float32(coord.Lat), Longitude: float32(coord.Lon)},
				Title: coord.OriginalString}}
		// coordsInfoChan <- &CoordInfo{Recepient: recepient,
		// 	Location: &telebot.Venue{Location: telebot.Location{Latitude: float32(coord.Lat), Longitude: float32(coord.Lon)},
		// 		Title: coord.OriginalString},
		// 	Options: nil}
	}
}

// IsBotCommand returns true if the message is a bot command or false otherwise
func IsBotCommand(m *tb.Message) bool {
	for _, entity := range m.Entities {
		if DEBUG {
			log.Printf("[DEBUG] Entity type: %s; offset: %d; length: %d", entity.Type, entity.Offset, entity.Length)
		}
		if entity.Type == "bot_command" {
			return true
		}
	}
	if m.Text[0] == '/' {
		return true
	}
	return false
}

///////////////////////////////

func processLevel(recepient tb.Recipient, engine *en.API) {
	var (
		retries   = 0
		levelInfo *en.Level
		err       error
	)

	for {
		// retries = 0
		levelInfo, err = engine.GetLevelInfo()
		if err != nil {
			if err.Error() == "No level info" {
				return
			}
			if retries > 3 {
				break
			}
			retries++
			time.Sleep(time.Second)
			log.Printf("Attempt #%d. Can't get level info: %s", retries, err)
			engine.Login()
			continue
		}
		break
	}
	if levelInfo == nil {
		log.Println("Can't find level information")
		return
	}
	engine.CurrentLevel = levelInfo

	engine.CurrentLevel.Tasks[0].TaskText, engine.CurrentLevel.Coords =
		en.ExtractCoordinates(engine.CurrentLevel.Tasks[0].TaskText)

	engine.CurrentLevel.Tasks[0].TaskText, engine.CurrentLevel.Images =
		en.ExtractImages(engine.CurrentLevel.Tasks[0].TaskText, "Картинка")

	engine.CurrentLevel.Tasks[0].TaskText = en.ReplaceCommonTags(engine.CurrentLevel.Tasks[0].TaskText)

	messageChan <- NewTextMessage(mainChat, engine.CurrentLevel.ToText(), tb.Message{})
	// sendInfoChan <- engine.CurrentLevel
	SendImageFromUrl(mainChat, engine.CurrentLevel.Images)
	SendCoords(mainChat, engine.CurrentLevel.Coords)
	//log.Printf("In func %p", &en.CurrentLevel.Coords)
	//SendLevelInfo(recepient, en.CurrentLevel)
}

func startWatching(engine *en.API) {
	var (
		ticker *time.Ticker
	)

	log.Print("Start monitoring game")
	ticker = time.NewTicker(1000 * time.Millisecond)
	quit = make(chan struct{})
	go func() {
		defer func() {
			if p := recover(); p != nil {
				log.Println(fmt.Errorf("внутренняя ошибка: %v", p))
			}
		}()

		for {
			select {
			case <-ticker.C:
				go func() {
					levelInfo, err := engine.GetLevelInfo()
					if err != nil {
						log.Println("Error:", err)
						return
					}
					levelInfoChan <- levelInfo
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

func setChat(chat tb.Chat) {
	var mu sync.RWMutex

	mu.Lock()
	defer mu.Unlock()

	mainChat = chat
}

func sendCode(engine *en.API, codesToSend []string, replyTo tb.Message) {
	var (
		mu    sync.RWMutex
		codes = en.Codes{Message: replyTo}
	)

	mu.Lock()
	defer mu.Unlock()

	defer func() {
		if p := recover(); p != nil {
			log.Println(fmt.Errorf("[sendCode] внутренняя ошибка: %v", p))
		}
	}()

	for _, code := range codesToSend {
		log.Printf("Sending code %q to EN engine", code)
		// TODO: 3) Do we need to send codes that were blocked ???
		if engine.CurrentLevel.IsPassed || engine.CurrentLevel.Dismissed ||
			(engine.CurrentLevel.BlockDuration > 0 && engine.CurrentLevel.HasAnswerBlockRule) {
			log.Printf("Level is blocked, can't send code %q", code)
			codes.NotSent = append(codes.NotSent, code)
			continue
		}

		lvl, err := engine.SendCode(code)
		//log.Println(lvl.HasAnswerBlockRule)
		if err != nil {
			log.Println("Failed to send code:", err)
		}
		if lvl.MixedActions[0].IsCorrect {
			codes.Correct = append(codes.Correct, code)
		} else {
			codes.Incorrect = append(codes.Incorrect, code)
		}
		levelInfoChan <- lvl
		time.Sleep(500 * time.Millisecond)
	}
	// sendInfoChan <- &codes
	messageChan <- TextMessage{Message: Message{Recipient: mainChat,
		Options: &tb.SendOptions{ParseMode: tb.ModeMarkdown,
			DisableWebPagePreview: true,
			ReplyTo:               codes.ReplyTo()}},
		Text: codes.ToText()}
}

func extractCommandAndArguments(m tb.Message) (command string, args string) {
	if len(m.Entities) > 0 {
		ent := m.Entities[0]
		command = m.Text[ent.Offset+1 : ent.Length]
		if len(command)+1 == len(m.Text) {
			args = ""
		} else {
			args = m.Text[ent.Length+1:]
		}
		if idx := strings.Index(command, "@"); idx != -1 {
			command = command[:idx]
		}
	} else {
		re := regexp.MustCompile("/([А-я]+)\\s*(.*)")
		result := re.FindStringSubmatch(m.Text)
		command, args = result[1], result[2]
	}
	return
}

func sectorsLeft(levelInfo *en.Level) {
	var sectors = en.NewExtendedLevelSectors(levelInfo)
	// sendInfoChan <- sectors
	messageChan <- TextMessage{Message: Message{Recipient: mainChat,
		Options: &tb.SendOptions{ParseMode: tb.ModeMarkdown,
			DisableWebPagePreview: true,
			ReplyTo:               sectors.ReplyTo()}},
		Text: sectors.ToText()}
}

func timeLeft(levelInfo *en.Level) {
	var msg = fmt.Sprintf(en.TimeLeftString, en.PrettyTimePrint(levelInfo.TimeoutSecondsRemain, true))
	// sendInfoChan <- NewBotMessage(msg)
	messageChan <- TextMessage{Message: Message{Recipient: mainChat,
		Options: &tb.SendOptions{ParseMode: tb.ModeMarkdown,
			DisableWebPagePreview: true}},
		Text: msg}
}

func listHelps(levelInfo *en.Level) {
	for _, helpInfo := range levelInfo.Helps {
		//log.Printf("==========================================: %s", helpInfo.HelpText)
		helpInfo.ProcessText()
		// sendInfoChan <- &helpInfo
		messageChan <- TextMessage{Message: Message{Recipient: mainChat,
			Options: &tb.SendOptions{ParseMode: tb.ModeMarkdown,
				DisableWebPagePreview: true,
				ReplyTo:               helpInfo.ReplyTo()}},
			Text: helpInfo.ToText()}
		//SendImageFromUrl(mainChat, helpInfo.images)
		//SendCoords(mainChat, helpInfo.coords)
	}
}

func timeHelpLeft(levelInfo *en.Level) {
	for _, help := range levelInfo.Helps {
		if help.RemainSeconds > 0 {
			var msg = fmt.Sprintf(en.HelpTimeLeft, help.Number, en.PrettyTimePrint(help.RemainSeconds, false))
			// sendInfoChan <- NewBotMessage(msg)
			messageChan <- TextMessage{Message: Message{Recipient: mainChat,
				Options: &tb.SendOptions{ParseMode: tb.ModeMarkdown,
					DisableWebPagePreview: true}},
				Text: msg}
			return
		}
	}
	// sendInfoChan <- NewBotMessage("Подсказок на уровне больше нет")
	messageChan <- NewTextMessage(mainChat, "Подсказок на уровне больше нет", tb.Message{})
}

func ProcessBotCommand(m tb.Message, en *en.API, bot *tb.Bot) {
	var (
		command     string
		args        string
		commandCode BotCommand
		ok          bool
		//ent         *telebot.MessageEntity
	)

	command, args = extractCommandAndArguments(m)
	if commandCode, ok = BotCommandDict[command]; !ok {
		log.Println("Unknown command:", command)
	}
	log.Printf("Command: %s, args: %s", command, args)

	switch commandCode {
	// case StartCommand:
	// 	var (
	// 		kbQuest        = tb.KeyboardButton{Text: "quest.ua", Data: "http://quest.ua"}
	// 		kbKharkovQuest = tb.KeyboardButton{Text: "kharkov.quest.ua", Data: "http://kharkov.quest.ua"}
	// 		kbKharkovEn    = tb.KeyboardButton{Text: "kharkov.en.cx", Data: "http://kharkov.en.cx"}
	// 	)
	// 	// var keyboard = tb.InlineKeyboardMarkup{}
	// 	// var reply = tb.ReplyMarkup{InlineKeyboard: }
	// 	var kbMarkup = [][]tb.KeyboardButton{{kbKharkovQuest, kbKharkovEn}, {kbQuest}}
	// 	err := bot.SendMessage(m.Chat, "Выберите домен",
	// 		&tb.SendOptions{ReplyTo: m,
	// 			ReplyMarkup: tb.ReplyMarkup{
	// 				Selective:       true,
	// 				ForceReply:      true,
	// 				ResizeKeyboard:  true,
	// 				OneTimeKeyboard: true,
	// 				InlineKeyboard:  kbMarkup,
	// 				// CustomKeyboard:  [][]string{[]string{"kharkov.en.cx", "kharkov.quest.ua"}, []string{"quest.ua"}},
	// 			}})
	// 	if err != nil {
	// 		log.Printf("ERROR: %s", err)
	// 	}
	// case InfoCommand:
	// 	//if m.Sender.Username == "kkolesnikov" && m.Chat == nil {
	// 	//	processLevel(m.Sender, en)
	// 	//} else {
	// 	processLevel(m.Chat, en)
	// 	//log.Printf("After func %p", &en.CurrentLevel.Coords)
	// 	//}
	case WatchCommand:
		startWatching(en)
	case StopWatchingCommand:
		stopWatching()
	case SetChatIDCommand:
		setChat(m.Chat)
	case CodeCommand:
		re := regexp.MustCompile("\\s+")
		sendCode(en, re.Split(args, -1), m)
	case CompositeCodeCommand:
		sendCode(en, []string{args}, m)
	case SectorsLeftCommand:
		sectorsLeft(en.CurrentLevel)
	case TimeLeftCommand:
		timeLeft(en.CurrentLevel)
	case ListHelpsCommand:
		listHelps(en.CurrentLevel)
	case HelpTimeCommand:
		timeHelpLeft(en.CurrentLevel)
	}
}

func CheckHelps(oldLevel *en.Level, newLevel *en.Level) {
	//log.Println("Check helps state")
	for i := range oldLevel.Helps {
		if oldLevel.Helps[i].Number == newLevel.Helps[i].Number {
			if oldLevel.Helps[i].HelpText == "" && newLevel.Helps[i].HelpText != "" {
				log.Println("New hint is available")
				newLevel.Helps[i].ProcessText()
				// sendInfoChan <- &newLevel.Helps[i]
				messageChan <- TextMessage{Message: Message{Recipient: mainChat,
					Options: &tb.SendOptions{ParseMode: tb.ModeMarkdown,
						DisableWebPagePreview: true,
						ReplyTo:               newLevel.Helps[i].ReplyTo()}},
					Text: newLevel.Helps[i].ToText()}
				SendCoords(mainChat, newLevel.Helps[i].Coords)
				SendImageFromUrl(mainChat, newLevel.Helps[i].Images)
			}
		}
	}
	//log.Println("Finish checking changes in Helps section")
}

func CheckSectors(oldLevel *en.Level, newLevel *en.Level) {
	//log.Println("Start checking changes in Sectors section")
	for i := range oldLevel.Sectors {
		if oldLevel.Sectors[i].Name == newLevel.Sectors[i].Name {
			if oldLevel.Sectors[i].IsAnswered != newLevel.Sectors[i].IsAnswered {
				log.Printf("Sector %q is closed, %d sectors left to close",
					newLevel.Sectors[i].Name, newLevel.SectorsLeftToClose)
				// TODO: Replace with constant or parameter from configuration
				if newLevel.SectorsLeftToClose <= 3 {
					//sectorChangeChan <- ExtendedSectorInfo{
					// sendInfoChan <- en.NewExtendedLevelSectors(newLevel)
					messageChan <- TextMessage{Message: Message{Recipient: mainChat,
						Options: &tb.SendOptions{ParseMode: tb.ModeMarkdown,
							DisableWebPagePreview: true,
							ReplyTo:               en.NewExtendedLevelSectors(newLevel).ReplyTo()}},
						Text: en.NewExtendedLevelSectors(newLevel).ToText()}
				}
			}
		}
	}
	//log.Println("Finish checking changes in Sectors section")
}

func CheckBonuses(oldLevel *en.Level, newLevel *en.Level) {
	for i := range oldLevel.Bonuses {
		if oldLevel.Bonuses[i].Name == newLevel.Bonuses[i].Name {
			if oldLevel.Bonuses[i].IsAnswered != newLevel.Bonuses[i].IsAnswered {
				log.Printf("Bonus %q is available, code %q", newLevel.Bonuses[i].Name,
					newLevel.Bonuses[i].Answer["Answer"])
				if newLevel.Bonuses[i].Help != "" {
					newLevel.Bonuses[i].ProcessText()
					// sendInfoChan <- &newLevel.Bonuses[i]
					messageChan <- TextMessage{Message: Message{Recipient: mainChat,
						Options: &tb.SendOptions{ParseMode: tb.ModeMarkdown,
							DisableWebPagePreview: true,
							ReplyTo:               newLevel.Bonuses[i].ReplyTo()}},
						Text: newLevel.Bonuses[i].ToText()}
					SendCoords(mainChat, newLevel.Bonuses[i].Coords)
					SendImageFromUrl(mainChat, newLevel.Bonuses[i].Images)
				}
			}
		}
	}
}

func CheckLevelTimeLeft(fsm *LevelTimeCheckingMachine, li *en.Level) {
	//log.Printf("FUNC fsm: %d", fsm.CurrentState().(TimeChecker).compareTime)
	if fsm.Process(li.TimeoutSecondsRemain * time.Second) {
		timeLeft(li)
		//log.Printf(TimeLeftString, PrettyTimePrint(li.TimeoutSecondsRemain, true))
	}
}

func CheckMixedActions(oldLevel *en.Level, newLevel *en.Level) {
	log.Println("Start checking changes in MixedActions section")
	sort.Sort(newLevel.MixedActions)
	//fmt.Println(len(newLevel.MixedActions))
	if len(newLevel.MixedActions) > 0 {
		if len(oldLevel.MixedActions) == 0 {
			for _, item := range newLevel.MixedActions {
				if item.IsCorrect {
					// sendInfoChan <- item
					messageChan <- NewTextMessage(mainChat, item.ToText(), item.ReplyTo())
				}
			}
		} else {
			lastActionID := oldLevel.MixedActions[0].ActionID
			for _, item := range newLevel.MixedActions {
				fmt.Println(item.ActionID, item.Answer, lastActionID)
				if item.ActionID == lastActionID {
					break
				}
				if item.IsCorrect {
					// sendInfoChan <- item
					messageChan <- NewTextMessage(mainChat, item.ToText(), item.ReplyTo())
				}
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
	// sendInfoChan = make(chan en.ToChat, 10)
	// photoInfoChan = make(chan *PhotoInfo, 10)
	// coordsInfoChan = make(chan *CoordInfo, 10)
	levelInfoChan = make(chan *en.Level, 10)
	messageChan = make(chan MessageSender, 10)
}

func initChat(bot *tb.Bot, chatID int64) tb.Chat {
	var chat = tb.Chat{ID: chatID}
	chat, err := bot.GetChat(chat)
	if err != nil {
		log.Print("Failed to update chat info")
	}
	return chat
}

func sendLevelInfo(info en.ToChat, channel chan en.ToChat, callback func() string, args ...string) {
	channel <- info

	if callback != nil {
		callback()
	}
}

func initTimeLevelChecking() *LevelTimeCheckingMachine {
	var (
		hourTimeChecker           = TimeChecker{60 * 60 * time.Second}
		halfHourTimeChecker       = TimeChecker{30 * 60 * time.Second}
		fifteenMinutesTimeChecker = TimeChecker{15 * 60 * time.Second}
		fiveMinutesTimeChecker    = TimeChecker{5 * 60 * time.Second}
		oneMinuteTimeChecker      = TimeChecker{60 * time.Second}
		zeroTimeChecker           = TimeChecker{-1 * time.Second}

		fsm   LevelTimeCheckingMachine
		rules = Ruleset{}
	)

	rules.AddTransition(hourTimeChecker, halfHourTimeChecker)
	rules.AddTransition(halfHourTimeChecker, fifteenMinutesTimeChecker)
	rules.AddTransition(fifteenMinutesTimeChecker, fiveMinutesTimeChecker)
	rules.AddTransition(fiveMinutesTimeChecker, oneMinuteTimeChecker)
	rules.AddTransition(oneMinuteTimeChecker, zeroTimeChecker)

	fsm = NewLevelTimeCheckingMachine(hourTimeChecker, &rules)

	return &fsm
}

func main() {
	var (
		envConfig     EnvConfig
		bot           *tb.Bot
		err           error
		updates       chan tb.Message
		update        tb.Message
		engine        en.API
		commandsStore *CommandStore
		fsm           *LevelTimeCheckingMachine
	)

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err = envconfig.Process("bonya", &envConfig)
	FailOnError(err, "Can't read environment variables")

	bot, err = tb.NewBot(envConfig.BotToken)
	FailOnError(err, "Can't connect to bot server")

	initChannels()
	fsm = initTimeLevelChecking()

	go func() {
		defer func() {
			if p := recover(); p != nil {
				log.Println(fmt.Errorf("[main/anonymous] внутренняя ошибка: %v", p))
			}
		}()

		for {
			select {
			// case ts := <-telegramSenderChan:
			// mi.Message.Send(ts.Recipient, ts.Message.Body(), ts.Options)
			// recepient := message.GetRecepient()
			// options := message.GetOptions()
			// body := message.GetBody()
			//
			// switch si.Type {
			case message := <-messageChan:
				message.Send(bot)
			// case PhotoMessage:
			// 	bot.SendPhoto(recipient, body, options)
			// case CoordinatesMessage:
			// 	bot.SendVenue(recipient, venue, options)
			// }
			// case nsi := <-sendInfoChan:
			// 	log.Print("Send text to Telegram chat")
			// 	//log.Println(nsi.ToText())
			// 	//log.Println(mainChat)
			// 	text := nsi.ToText()
			// 	//if len(text) > 4096 {
			// 	//	for
			// 	//}
			// 	err := bot.SendMessage(mainChat, text,
			// 		&tb.SendOptions{ParseMode: tb.ModeMarkdown,
			// 			DisableWebPagePreview: true,
			// 			ReplyTo:               nsi.ReplyTo()})
			// 	if err != nil {
			// 		log.Println(err)
			// 	}
			// case ci := <-coordsInfoChan:
			// 	log.Print("Send coordinates to chat")
			// 	bot.SendVenue(ci.Recepient, ci.Location, ci.Options)
			// case pi := <-photoInfoChan:
			// 	log.Print("Send images to Telegram chat")
			// 	bot.SendPhoto(pi.Recepient, pi.Photo, pi.Options)
			case li := <-levelInfoChan:
				//log.Println("Receive level from channel")
				if isNewLevel(engine.CurrentLevel, li) {
					log.Printf("New level #%d", li.Number)
					engine.CurrentLevel = li
					engine.CurrentLevel.ProcessText()
					fsm.ResetState(li.Timeout * time.Second)

					// sendLevelInfo(engine.CurrentLevel, sendInfoChan, nil)
					SendImageFromUrl(mainChat, engine.CurrentLevel.Images)
					SendCoords(mainChat, engine.CurrentLevel.Coords)
				}
				go CheckHelps(engine.CurrentLevel, li)
				//go CheckMixedActions(en.CurrentLevel, li)
				go CheckBonuses(engine.CurrentLevel, li)
				go CheckSectors(engine.CurrentLevel, li)
				go CheckLevelTimeLeft(fsm, li)
				engine.CurrentLevel = li
			}
		}
	}()

	jar, _ := cookiejar.New(nil)
	engine = en.API{
		Username:      envConfig.User,
		Password:      envConfig.Password,
		Client:        &http.Client{Jar: jar},
		CurrentGameID: envConfig.GameID,
		CurrentLevel:  nil,
		Domain:        envConfig.EngineDomain,
		Levels:        list.New()}
	engine.Login2(envConfig.User, envConfig.Password)
	// engine.Login()

	log.Printf("Authorized on account %s", bot.Identity.Username)
	updates = make(chan tb.Message, 50)
	callbacks := make(chan tb.Callback, 50)
	bot.Messages = updates
	bot.Callbacks = callbacks
	go bot.Start(30 * time.Second)
	// bot.Listen(updates, 30*time.Second)

	// setChat(initChat(bot, envConfig.MainChat))
	// engine.CurrentLevel, _ = engine.GetLevelInfo()
	// if engine.CurrentLevel != nil {
	// 	fsm.ResetState(engine.CurrentLevel.TimeoutSecondsRemain * time.Second)
	// 	log.Printf("MAIN fsm: %.0f minute(s)", fsm.CurrentState().(TimeChecker).compareTime.Minutes())
	// }
	// startWatching(&engine)

	go startServer(&engine)

	commandsStore = NewCommandStore()
	commandsStore.init()

	for {
		select {
		case update = <-bot.Messages:
			//log.Printf("Read updates from Telegram: %s", update.Text)
			if update.Text != "" {
				log.Printf("[INFO] [%s@%s(%d)] %s", update.Sender.Username, update.Chat.Title, update.Chat.ID,
					update.Text)
				if IsBotCommand(&update) {
					commandName, arguments := extractCommandAndArguments(update)
					commandHandler, err := commandsStore.Get(commandName)
					if err != nil {
						log.Printf("[WARNING] %s", err)
					}
					// TODO: according to the sender/chat need to find corresponding game and get current level for it
					// TODO: pass Game struct rather than just levelInfo, because there is no levelInfo at the start
					levelInfo, err := engine.GetLevelInfo()
					command, err := commandHandler(messageChan, update, levelInfo)
					if err != nil {
						log.Printf("[ERROR] Something bad happened when constructing command handler: %s", err)
					}
					go command.Process(arguments)
					// go ProcessBotCommand(&update, &engine, bot)
				}

			}
		// case callback := <-callbacks:
		case callback := <-bot.Callbacks:
			log.Printf("CALLBACK: %s %s", callback.Sender.Username, callback.Data)
			bot.AnswerCallbackQuery(&callback, &tb.CallbackResponse{CallbackID: callback.ID})
			// TODO: maybe store fsm for separate chat in redis, or in memory. and when
			//       get update for certain chat retrieve object with correct state

			//bot.SendMessage(update.Chat, fmt.Sprintf("Dear %s, I can't understand you", update.Sender.Username),
			//	&telebot.SendOptions{ReplyTo: update, ParseMode: telebot.ModeMarkdown})
		}
	}

}
