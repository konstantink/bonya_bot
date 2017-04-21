package main

import (
	"gopkg.in/telegram-bot-api.v4"
	//"net/http"
	"fmt"
	"log"
	//"net/http/cookiejar"
	//"errors"
	//"container/list"
	//"time"
	//"sort"
	"github.com/kelseyhightower/envconfig"
	//"github.com/tucnak/telebot"
)

type EnResponse interface {
	ReadData(data []byte)
}

//type EnCookie struct {
//	guid   string `json:"GUID"`
//	atoken string `json:"atoken"`
//	stoken string `json:"stoken"`
//}

//type LevelChangeStruct struct {}
//type HelpChangeStruct struct {}

//type SectorChan chan ExtendedSectorInfo
//type LevelChan chan LevelInfo
//type MessageChan chan tgbotapi.MessageConfig
//type PhotoChan chan tgbotapi.PhotoConfig

var (
// quit chan struct{}
//messageChan MessageChan
//photoChan PhotoChan
//levelChan chan LevelInfo
//levelChangeChan LevelChan
//helpChangeChan chan HelpInfo
//sectorChangeChan SectorChan
//mixedActionChangeChan chan MixedActionInfo
)

var (
	admins       = [...]int{47700349}
	chatId int64 = 0
)

//type EnLoginResponse struct {
//	Error int `json:"Error"`
//	Message string `json:"Message"`
//	InUnblockUrl string `json:"InUnblockUrl"`
//	BruteForceUnblockUrl string `json:"BruteForceUnblockUrl"`
//	ConfirmEmailUrl string `json:"ConfirmEmailUrl"`
//	CaptchaUrl string `json:"CaptchaUrl"`
//	AdminWhoCanActivate string `json:"AdminWhoCanActivate"`
//	Cookies []http.Cookie `json:"-"`
//}

//func (en *EnLoginResponse) ReadData(data []byte) error {
//	err := json.Unmarshal(data, en)
//	if err != nil {
//		return err
//	}
//	return nil
//}

//const (
//	enAPI *EnAPI = EnAPI{"Harry_Potter", "toknkpils85", new(http.Client)}
//)

func isAdmin(id int) bool {
	for _, userId := range admins {
		if userId == id {
			return true
		}
	}
	return false
}

func sendMessage(bot *tgbotapi.BotAPI, chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "Markdown"
	//msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	bot.Send(msg)
}

//func processBotCommand(m *tgbotapi.Message, en *EnAPI, bot *tgbotapi.BotAPI) error {
//	//var commandReader bufio.Reader
//	var (
//		command []byte
//		commandCode BotCommand
//		ok bool
//		ent *tgbotapi.MessageEntity
//	)
//
//	ent = GetBotCommandEntity(m)
//	if ent == nil {
//		return errors.New("No bot command found")
//	}
//	command = make([]byte, ent.Length-1, ent.Length-1)
//	copy(command, m.Text[ent.Offset+1:ent.Length])
//	if commandCode, ok = BotCommandDict[string(command)]; !ok {
//		return errors.New("Unknown command")
//	}
//
//	switch commandCode{
//	case InfoCommand:
//		var res, text string
//		if en.CurrentLevel == nil {
//			lvl, _ := en.GetLevelInfo()
//			en.CurrentLevel = lvl.Level
//		}
//		log.Println("TIMEOUT:", uint64(en.CurrentLevel.Timeout))
//		//res = regexp.MustCompile("_").ReplaceAllLiteralString(res, ("\\__"))
//		res = ReplaceCoordinates(en.CurrentLevel.Tasks[0].TaskText)
//		res = ReplaceImages(res)
//		res = ReplaceCommonTags(res)
//		text = fmt.Sprintf(LevelInfoString, en.CurrentLevel.Number, en.CurrentLevel.Name,
//			PrettyTimePrint(en.CurrentLevel.Timeout),
//			PrettyTimePrint(en.CurrentLevel.TimeoutSecondsRemain),
//			res)
//		                   //(en.CurrentLevel.TimeoutSecondsRemain*time.Second).String())
//		if len(text) > 4096 {
//			fullMessageCount := len(text) / 4096
//			for i := 0; i < fullMessageCount; i++ {
//				sendMessage(bot, m.Chat.ID, text[i*4096:(i+1)*4096])
//			}
//			sendMessage(bot, m.Chat.ID, text[fullMessageCount*4096:])
//		} else {
//			sendMessage(bot, m.Chat.ID, text)
//		}
//		//lvl, _ := en.GetLevelInfo()
//	case SetChatIdCommand:
//		if isAdmin(m.From.ID){
//			log.Println("Set chat ID as working chat")
//			chatId = m.Chat.ID
//			text := fmt.Sprintf("Теперь буду работать только с чатом %q, остальные буду игнорировать",
//				m.Chat.Title)
//			messageChan <- tgbotapi.NewMessage(chatId, text)
//		} else {
//			text := fmt.Sprint("Выбирать чат могут только администраторы")
//			messageChan <- tgbotapi.NewMessage(m.Chat.ID, text)
//		}
//	case WatchCommand:
//		ticker := time.NewTicker(time.Second)
//		quit = make(chan struct{})
//		go func() {
//			for {
//				select{
//				case <- ticker.C: {
//					lvl, err := en.GetLevelInfo()
//					//log.Println("Error:", err)
//					//en.GetLevelInfo()
//					if err != nil {
//						log.Println("Error:", err)
//						continue
//					}
//					log.Println("Send level to channel")
//					levelChan <- *(lvl.Level)
//				}
//				case <- quit:
//					ticker.Stop()
//				}
//			}
//		}()
//	case StopWatchingCommand:
//		close(quit)
//	case TestHelpChange:
//		log.Println("Command help change")
//		helpChangeChan <- HelpInfo{}
//
//	//if en.CurrentLevel == nil {
//	//	en.CurrentLevel = lvl
//	//	return lvl, err
//	//}
//	//
//	//if en.CurrentLevel.Number < lvl.Level.Number {
//	//	en.Levels.InsertAfter(en.CurrentLevel, en.Levels.Front())
//	//	en.CurrentLevel = lvl.Level
//	//} else {
//	//	//en.
//	//}
//
//	}
//	return nil
//}

//func helpChangeHandler(hi *HelpInfo) (msg tgbotapi.MessageConfig) {
//	log.Println("Help is changed")
//	var text string = fmt.Sprintf(HelpInfoString, hi.Number, hi.HelpText)
//	msg = tgbotapi.NewMessage(chatId, text)
//	msg.ParseMode = "HTML"
//	return
//	//msg := tgbotapi.NewMessage(m.Chat.ID, text)
//	//msg.ParseMode = "Markdown"
//	//bot.Send(msg)
//}

//func mixedActionChangeHandler(mai *MixedActionInfo) (msg tgbotapi.MessageConfig) {
//	log.Println("New MixedAction is added")
//	var text string
//	if mai.IsCorrect {
//		text = fmt.Sprintf(CorrectAnswerString, mai.Answer, mai.Login)
//	} else {
//		text = fmt.Sprintf(IncorrectAnswerString, mai.Answer, mai.Login)
//	}
//	msg = tgbotapi.NewMessage(chatId, text)
//	msg.ParseMode = "Markdown"
//	return
//}

//func sectorChangeHandler(esi *ExtendedSectorInfo) (msg tgbotapi.MessageConfig){
//	log.Println("Some sector is changed")
//	var text string = fmt.Sprintf(SectorInfoString, esi.sectorInfo.Name, esi.sectorsLeft, esi.totalSectors)
//	msg = tgbotapi.NewMessage(chatId, text)
//	msg.ParseMode = "Markdown"
//	return
//}

//func levelChangeHandler(li *LevelInfo) (msg tgbotapi.MessageConfig) {
//	log.Println("Level is changed")
//	//var text string = fmt.Sprintf(LevelInfoString)
//	return
//}

//func processChanges(en *EnAPI) {
//	for {
//		select {
//		case level := <- levelChan: {
//			log.Println("Receive level from channel")
//			if isNewLevel(en.CurrentLevel, &level) {
//				log.Println("Level is new")
//				levelChangeChan <- level
//				continue
//			}
//			go CheckHelps(*en.CurrentLevel, level)
//			go CheckMixedActions(*en.CurrentLevel, level)
//			go CheckSectors(*en.CurrentLevel, level)
//			en.CurrentLevel = &level
//			//time.Sleep(5*time.Second)
//		}
//		//default:
//		//	log.Println("No changes")
//		}
//	}
//}

//func processLevelChanges() {
//	for {
//		select {
//		case hi := <-helpChangeChan:
//			msg := helpChangeHandler(&hi)
//			messageChan <- msg
//		case mai := <-mixedActionChangeChan:
//			msg := mixedActionChangeHandler(&mai)
//			messageChan <- msg
//		case si := <-sectorChangeChan:
//			msg := sectorChangeHandler(&si)
//			messageChan <- msg
//		case li := <-levelChangeChan:
//			msg := levelChangeHandler(&li)
//			messageChan <- msg
//		}
//	}
//}

func not_main() {
	var (
		envConfig EnvConfig
	)
	err := envconfig.Process("bonya", &envConfig)
	fmt.Println(envConfig)
	//bot, err := tgbotapi.NewBotAPI(envConfig.BotToken)
	//bot, err := telebot.NewBot(envConfig.BotToken)
	if err != nil {
		log.Panic(err)
	}

	//bot.Debug = true
	//bot.RemoveWebhook()

	//log.Printf("Authorized on account %s", bot.Self.UserName)
	//log.Printf("Authorized on account %s", bot.Identity.Username)

	//jar, _ := cookiejar.New(nil)
	//en := &EnAPI{
	//	login: envConfig.User,
	//	password: envConfig.Password,
	//	Client: &http.Client{Jar: jar},
	//	CurrentGameId: 25733,
	//	CurrentLevel: nil,
	//	Levels: list.New()}
	//en.Login()

	//u := tgbotapi.NewUpdate(0)
	//u.Timeout = 5
	//updates, _ := bot.GetUpdatesChan(u)
	//updates := make(chan telebot.Message)
	//bot.Listen(updates, 60*time.Second)

	//helpChangeChan = make(chan HelpInfo, 5)
	//messageChan = make(MessageChan)
	//photoChan = make(PhotoChan)
	//levelChan = make(chan LevelInfo)
	//levelChangeChan = make(LevelChan, 5)
	//sectorChangeChan = make(SectorChan, 5)
	//mixedActionChangeChan = make(chan MixedActionInfo, 5)

	//go processChanges(en)
	//go processLevelChanges()
	//go func(bot *tgbotapi.BotAPI) {
	//	log.Println("Read message from channel to send to chat", chatId)
	//	for {
	//		if chatId != 0 {
	//			select {
	//			case msg := <- messageChan:
	//				//msg := tgbotapi.NewMessage(chatId, text)
	//				//msg.ParseMode = "Markdown"
	//				bot.Send(msg)
	//			case ph := <- photoChan:
	//				log.Println("Sending photo to the chat")
	//				bot.Send(ph)
	//			}
	//		}
	//	}
	//}(bot)

	//_, err = bot.SetWebhook(tgbotapi.NewWebhookWithCert("https://178.165.58.187:8443/"+bot.Token, "bonya-cert.pem"))
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//updates := bot.ListenForWebhook("/" + bot.Token)
	//go http.ListenAndServeTLS("0.0.0.0:8443", "bonya-cert.pem", "bonya-key.pem", nil)

	//for update := range updates {
	//	log.Printf("%+v\n", update)
	//if IsMessageBotCommand(update.Message) {
	//	log.Println("It is bot command")
	//	processBotCommand(update.Message, en, bot)
	//}
	//}

	//for update := range updates {
	//	if update.Message == nil {
	//	    continue
	//	}
	//
	//	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	//	if IsBotCommand(update.Message) {
	//		log.Println("It is bot command")
	//		processBotCommand(update.Message, en, bot)
	//	}
	// }

	//en.MakeRequest(fmt.Sprintf(EnAddress, LoginEndpoint))
	//en.MakeRequest(fmt.Sprintf(EnAddress, LoginEndpoint))
	//fmt.Println(authResponse)
	//fmt.Println(en.Client.Jar.Cookies())
	//lvl, _ := en.GetLevelInfo(25733)
	//log.Printf("Levels number: %d", lvl.Levels.Len())
	//en.SendCode(25733, "235035")
	//fmt.Println(resp.Error)
	//fmt.Println(time.Duration(259200*time.Second).String())
	//fmt.Println(time.Unix(63611117108400, 0))
}
