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

	"github.com/kelseyhightower/envconfig"
	"github.com/tucnak/telebot"
)

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

type CoordInfo struct {
	Recepient telebot.Recipient
	Location  *telebot.Venue
	Options   *telebot.SendOptions
}

var (
	quit           chan struct{}
	sendInfoChan   chan ToChat
	photoInfoChan  chan *PhotoInfo
	coordsInfoChan chan *CoordInfo
	levelInfoChan  chan *LevelInfo

	mainChat telebot.Chat
)

// Helpers

func SendImageFromUrl(recepient telebot.Recipient, images []Image) {
	var (
		file *os.File
	)
	for _, img := range images {

		resp, err := http.Get(img.url)
		if err != nil {
			log.Println("Can't download image:", err)
		}

		filename := fmt.Sprintf("/tmp/%s", path.Base(img.url))
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
		telebotFile, _ := telebot.NewFile(file.Name())
		log.Printf("Sending photo to the channel (%s)", file.Name())
		thumbnail := telebot.Thumbnail{File: telebotFile, Width: 120, Height: 120}
		photoInfoChan <- &PhotoInfo{Recepient: recepient, Photo: &telebot.Photo{File: telebotFile,
			Thumbnail: thumbnail, Caption: img.caption}, Options: nil}
	}
}

func SendCoords(recepient telebot.Recipient, coords Coordinates) {
	for _, coord := range coords {
		coordsInfoChan <- &CoordInfo{Recepient: recepient,
			Location: &telebot.Venue{Location: telebot.Location{Latitude: float32(coord.Lat), Longitude: float32(coord.Lon)},
				Title: coord.OriginalString},
			Options: nil}
	}
}

func IsBotCommand(m *telebot.Message) bool {
	for _, entity := range m.Entities {
		log.Print("Entity:", entity)
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

func processLevel(recepient telebot.Recipient, en *EnAPI) {
	var (
		retries   int
		levelInfo *LevelInfo
		err       error
	)

	for {
		retries = 0
		levelInfo, err = en.GetLevelInfo()
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
			en.Login()
			continue
		}
		break
	}
	if levelInfo == nil {
		log.Println("Can't find level information")
		return
	}
	en.CurrentLevel = levelInfo

	en.CurrentLevel.Tasks[0].TaskText, en.CurrentLevel.Coords =
		ReplaceCoordinates(en.CurrentLevel.Tasks[0].TaskText)

	en.CurrentLevel.Tasks[0].TaskText, en.CurrentLevel.Images =
		ReplaceImages(en.CurrentLevel.Tasks[0].TaskText, "Картинка")

	en.CurrentLevel.Tasks[0].TaskText = ReplaceCommonTags(en.CurrentLevel.Tasks[0].TaskText)

	sendInfoChan <- en.CurrentLevel
	SendImageFromUrl(mainChat, en.CurrentLevel.Images)
	SendCoords(mainChat, en.CurrentLevel.Coords)
	//log.Printf("In func %p", &en.CurrentLevel.Coords)
	//SendLevelInfo(recepient, en.CurrentLevel)
}

func startWatching(en *EnAPI) {
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
					levelInfo, err := en.GetLevelInfo()
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

func setChat(chat telebot.Chat) {
	var mu sync.RWMutex

	mu.Lock()
	defer mu.Unlock()

	mainChat = chat
}

func sendCode(en *EnAPI, codesToSend []string, replyTo telebot.Message) {
	var (
		mu    sync.RWMutex
		codes Codes = Codes{replyTo: replyTo}
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
		if en.CurrentLevel.IsPassed || en.CurrentLevel.Dismissed ||
			(en.CurrentLevel.BlockDuration > 0 && en.CurrentLevel.HasAnswerBlockRule) {
			log.Printf("Level is blocked, can't send code %q", code)
			codes.notSent = append(codes.notSent, code)
			continue
		}

		lvl, err := en.SendCode(code)
		//log.Println(lvl.HasAnswerBlockRule)
		if err != nil {
			log.Println("Failed to send code:", err)
		}
		if lvl.MixedActions[0].IsCorrect {
			codes.correct = append(codes.correct, code)
		} else {
			codes.incorrect = append(codes.incorrect, code)
		}
		levelInfoChan <- lvl
		time.Sleep(500 * time.Millisecond)
	}
	sendInfoChan <- &codes
}

func extractCommandAndArguments(m *telebot.Message) (command string, args string) {
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
		re := regexp.MustCompile("^/([А-я]+)\\s*(.*)")
		log.Printf("Analyzing command: %s", m.Text)
		result := re.FindStringSubmatch(m.Text)
		if len(result) == 0{
			log.Printf("[WARNING] Can't extract command and arguments from %s", m.Text)
			return
		}
		if len(result) > 2 {
			command, args = result[1], result[2]
		} else {
			command = result[1]
			args = ""
		}
	}
	return
}

func sectorsLeft(levelInfo *LevelInfo) {
	var sectors = NewExtendedLevelSectors(levelInfo)
	sendInfoChan <- sectors
}

func timeLeft(levelInfo *LevelInfo) {
	var msg = fmt.Sprintf(TimeLeftString, PrettyTimePrint(levelInfo.TimeoutSecondsRemain, true))
	sendInfoChan <- NewBotMessage(msg)
}

func listHelps(levelInfo *LevelInfo) {
	for _, helpInfo := range levelInfo.Helps {
		//log.Printf("==========================================: %s", helpInfo.HelpText)
		helpInfo.ProcessText()
		sendInfoChan <- &helpInfo
		SendImageFromUrl(mainChat, helpInfo.Images)
		SendCoords(mainChat, helpInfo.Coords)
	}
}

func timeHelpLeft(levelInfo *LevelInfo) {
	for _, help := range levelInfo.Helps {
		if help.RemainSeconds > 0 {
			var msg = fmt.Sprintf(HelpTimeLeft, help.Number, PrettyTimePrint(help.RemainSeconds, false))
			sendInfoChan <- NewBotMessage(msg)
			return
		}
	}
	sendInfoChan <- NewBotMessage("Подсказок на уровне больше нет")
}

func ProcessBotCommand(m *telebot.Message, en *EnAPI) {
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
	case InfoCommand:
		//if m.Sender.Username == "kkolesnikov" && m.Chat == nil {
		//	processLevel(m.Sender, en)
		//} else {
		processLevel(m.Chat, en)
		//log.Printf("After func %p", &en.CurrentLevel.Coords)
		//}
	case WatchCommand:
		startWatching(en)
	case StopWatchingCommand:
		stopWatching()
	case SetChatIdCommand:
		setChat(m.Chat)
	case CodeCommand:
		re := regexp.MustCompile("\\s+")
		sendCode(en, re.Split(args, -1), *m)
	case CompositeCodeCommand:
		sendCode(en, []string{args}, *m)
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

func CheckHelps(oldLevel *LevelInfo, newLevel *LevelInfo) {
	//log.Println("Check helps state")
	for i, _ := range oldLevel.Helps {
		if oldLevel.Helps[i].Number == newLevel.Helps[i].Number {
			if oldLevel.Helps[i].HelpText == "" && newLevel.Helps[i].HelpText != "" {
				log.Println("New hint is available")
				newLevel.Helps[i].ProcessText()
				sendInfoChan <- &newLevel.Helps[i]
				SendCoords(mainChat, newLevel.Helps[i].Coords)
				SendImageFromUrl(mainChat, newLevel.Helps[i].Images)
			}
		}
	}
	//log.Println("Finish checking changes in Helps section")
}

func CheckSectors(oldLevel *LevelInfo, newLevel *LevelInfo) {
	//log.Println("Start checking changes in Sectors section")
	for i, _ := range oldLevel.Sectors {
		if oldLevel.Sectors[i].Name == newLevel.Sectors[i].Name {
			if oldLevel.Sectors[i].IsAnswered != newLevel.Sectors[i].IsAnswered {
				log.Printf("Sector %q is closed, %d sectors left to close",
					newLevel.Sectors[i].Name, newLevel.SectorsLeftToClose)
				// TODO: Replace with constant or parameter from configuration
				if newLevel.SectorsLeftToClose <= 3 {
					//sectorChangeChan <- ExtendedSectorInfo{
					sendInfoChan <- NewExtendedLevelSectors(newLevel)
				}
			}
		}
	}
	//log.Println("Finish checking changes in Sectors section")
}

func CheckBonuses(oldLevel *LevelInfo, newLevel *LevelInfo) {
	for i, _ := range oldLevel.Bonuses {
		if oldLevel.Bonuses[i].Name == newLevel.Bonuses[i].Name {
			if oldLevel.Bonuses[i].IsAnswered != newLevel.Bonuses[i].IsAnswered || (oldLevel.Bonuses[i].SecondsToStart != newLevel.Bonuses[i].SecondsToStart && newLevel.Bonuses[i].SecondsToStart == 0) {
				log.Printf("Bonus %q is available, code %q", newLevel.Bonuses[i].Name,
					newLevel.Bonuses[i].Answer["Answer"])
				if newLevel.Bonuses[i].Help != "" || newLevel.Bonuses[i].Task != "" {
					newLevel.Bonuses[i].ProcessText()
					sendInfoChan <- &newLevel.Bonuses[i]
					SendCoords(mainChat, newLevel.Bonuses[i].Coords)
					SendImageFromUrl(mainChat, newLevel.Bonuses[i].Images)
				}
			}
		}
	}
}

func CheckLevelTimeLeft(fsm *LevelTimeCheckingMachine, li *LevelInfo) {
	//log.Printf("FUNC fsm: %d", fsm.CurrentState().(TimeChecker).compareTime)
	if fsm.CheckTime(li.TimeoutSecondsRemain * time.Second) {
		timeLeft(li)
		//log.Printf(TimeLeftString, PrettyTimePrint(li.TimeoutSecondsRemain, true))
	}
}

func CheckMixedActions(oldLevel *LevelInfo, newLevel *LevelInfo) {
	log.Println("Start checking changes in MixedActions section")
	sort.Sort(newLevel.MixedActions)
	//fmt.Println(len(newLevel.MixedActions))
	if len(newLevel.MixedActions) > 0 {
		if len(oldLevel.MixedActions) == 0 {
			for _, item := range newLevel.MixedActions {
				if item.IsCorrect {
					sendInfoChan <- item
				}
			}
		} else {
			lastActionId := oldLevel.MixedActions[0].ActionId
			for _, item := range newLevel.MixedActions {
				fmt.Println(item.ActionId, item.Answer, lastActionId)
				if item.ActionId == lastActionId {
					break
				}
				if item.IsCorrect {
					sendInfoChan <- item
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
	sendInfoChan = make(chan ToChat, 10)
	photoInfoChan = make(chan *PhotoInfo, 10)
	coordsInfoChan = make(chan *CoordInfo, 10)
	levelInfoChan = make(chan *LevelInfo, 10)
}

func initChat(bot *telebot.Bot, chatId int64) telebot.Chat {
	var chat = telebot.Chat{ID: chatId}
	chat, err := bot.GetChat(chat)
	if err != nil {
		log.Print("Failed to update chat info")
	}
	return chat
}

func sendLevelInfo(info ToChat, channel chan ToChat, callback func() string, args ...string) {
	channel <- info

	if callback != nil {
		callback()
	}
}

func initTimeLevelChecking() *LevelTimeCheckingMachine {
	var (
		hourTimeChecker           TimeChecker = TimeChecker{60 * 60 * time.Second}
		halfHourTimeChecker       TimeChecker = TimeChecker{30 * 60 * time.Second}
		fifteenMinutesTimeChecker TimeChecker = TimeChecker{15 * 60 * time.Second}
		fiveMinutesTimeChecker    TimeChecker = TimeChecker{5 * 60 * time.Second}
		oneMinuteTimeChecker      TimeChecker = TimeChecker{60 * time.Second}
		zeroTimeChecker           TimeChecker = TimeChecker{-1 * time.Second}

		fsm   LevelTimeCheckingMachine
		rules Ruleset = Ruleset{}
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
		envConfig EnvConfig
		bot       *telebot.Bot
		err       error
		updates   chan telebot.Message
		update    telebot.Message
		en        EnAPI
		fsm       *LevelTimeCheckingMachine
	)

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	err = envconfig.Process("bonya", &envConfig)
	FailOnError(err, "Can't read environment variables")

	bot, err = telebot.NewBot(envConfig.BotToken)
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
			case nsi := <-sendInfoChan:
				log.Print("Send text to Telegram chat")
				//log.Println(nsi.ToText())
				//log.Println(mainChat)
				text := nsi.ToText()
				//if len(text) > 4096 {
				//	for
				//}
				err := bot.SendMessage(mainChat, text,
					&telebot.SendOptions{ParseMode: telebot.ModeMarkdown,
						DisableWebPagePreview: true,
						ReplyTo:               nsi.ReplyTo()})
				if err != nil {
					log.Println(err)
				}
			case ci := <-coordsInfoChan:
				log.Print("Send coordinates to chat")
				bot.SendVenue(ci.Recepient, ci.Location, ci.Options)
			case pi := <-photoInfoChan:
				log.Print("Send images to Telegram chat")
				bot.SendPhoto(pi.Recepient, pi.Photo, pi.Options)
			case li := <-levelInfoChan:
				//log.Println("Receive level from channel")
				if isNewLevel(en.CurrentLevel, li) {
					log.Printf("New level #%d", li.Number)
					en.CurrentLevel = li
					en.CurrentLevel.ProcessText()
					fsm.ResetState(li.Timeout * time.Second)

					sendLevelInfo(en.CurrentLevel, sendInfoChan, nil)
					SendImageFromUrl(mainChat, en.CurrentLevel.Images)
					SendCoords(mainChat, en.CurrentLevel.Coords)
				}
				go CheckHelps(en.CurrentLevel, li)
				//go CheckMixedActions(en.CurrentLevel, li)
				go CheckBonuses(en.CurrentLevel, li)
				go CheckSectors(en.CurrentLevel, li)
				go CheckLevelTimeLeft(fsm, li)
				en.CurrentLevel = li
			}
		}
	}()

	jar, _ := cookiejar.New(nil)
	log.Println(envConfig.User, envConfig.Password)
	en = EnAPI{
		Username:      envConfig.User,
		Password:      envConfig.Password,
		Client:        &http.Client{Jar: jar},
		CurrentGameID: envConfig.GameId,
		CurrentLevel:  nil,
		Domain:        envConfig.EngineDomain,
		Levels:        list.New()}
	en.Login()

	log.Printf("Authorized on account %s", bot.Identity.Username)
	updates = make(chan telebot.Message, 50)
	bot.Listen(updates, 30*time.Second)

	setChat(initChat(bot, envConfig.MainChat))
	en.CurrentLevel, _ = en.GetLevelInfo()
	if en.CurrentLevel != nil {
		fsm.ResetState(en.CurrentLevel.TimeoutSecondsRemain * time.Second)
		log.Printf("MAIN fsm: %.0f minute(s)", fsm.CurrentState().(TimeChecker).compareTime.Minutes())
	}
	startWatching(&en)

	go startServer(&en)

	for {
		select {
		case update = <-updates:
			//log.Printf("Read updates from Telegram: %s", update.Text)
			if update.Text != "" {
				if IsBotCommand(&update) {
					go ProcessBotCommand(&update, &en)
				}

				log.Printf("[%s@%s(%d)] %s", update.Sender.Username, update.Chat.Title, update.Chat.ID,
					update.Text)
			}
			//bot.SendMessage(update.Chat, fmt.Sprintf("Dear %s, I can't understand you", update.Sender.Username),
			//	&telebot.SendOptions{ReplyTo: update, ParseMode: telebot.ModeMarkdown})
		}
	}

}
