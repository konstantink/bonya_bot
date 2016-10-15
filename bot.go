package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"net/http"
	"fmt"
	"net/url"
	"io/ioutil"
	"log"
	"encoding/json"
	"net/http/cookiejar"
	"bytes"
	"strings"
	"errors"
	"io"
	"os"
	"container/list"
	"time"
	"sort"
	"regexp"
	"strconv"
	"path"
)

type EnResponse interface {
	ReadData(data []byte)
}

//type EnCookie struct {
//	guid   string `json:"GUID"`
//	atoken string `json:"atoken"`
//	stoken string `json:"stoken"`
//}

type EnAPI struct {
	login         string       `json:"Login"`
	password      string       `json:"Password"`
	Client        *http.Client `json:"-"`
	CurrentGameId int32        `json:"-"`
	CurrentLevel  *LevelInfo   `json:"-"`
	Levels        *list.List   `json:"-"`
}

type EnAPIAuthResponse struct {
	Ok          bool
	Cookies     []*http.Cookie
	Result      json.RawMessage
	StatusCode  int
	Description string
}

func (apiResp *EnAPIAuthResponse) CreateFromResponse(resp *http.Response) error {
	var (
		bytes []byte
		err error
		respBody map[string]interface{}
	)
	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = json.Unmarshal(bytes, &respBody)
	if err != nil {
		return err
	}

	apiResp.Ok = respBody["Error"].(float64) == 0
	if !apiResp.Ok {
		apiResp.Description = respBody["Description"].(string)
	} else {
		apiResp.Description = ""
	}
	apiResp.Result = bytes
	apiResp.Cookies = resp.Cookies()
	apiResp.StatusCode = resp.StatusCode

	return nil
}

type BotCommand int8

const (
	InfoCommand BotCommand = 1 << iota
	SetChatIdCommand
	WatchCommand
	StopWatchingCommand
	TestHelpChange
)

var (
	BotCommandDict map[string]BotCommand = map[string]BotCommand{
	"info": InfoCommand,
	"setchat": SetChatIdCommand,
	"watch": WatchCommand,
	"stopwatching": StopWatchingCommand,
	"helpchange": TestHelpChange}
)

//type LevelChangeStruct struct {}
//type HelpChangeStruct struct {}

type ExtendedSectorInfo struct {
	sectorInfo    *SectorInfo
	sectorsPassed int8
	sectorsLeft   int8
	totalSectors  int8
}

type SectorChan chan ExtendedSectorInfo
type LevelChan chan LevelInfo
type MessageChan chan tgbotapi.MessageConfig
type PhotoChan chan tgbotapi.PhotoConfig

var (
	quit chan struct{}
	messageChan MessageChan
	photoChan PhotoChan
	levelChan chan LevelInfo
	levelChangeChan LevelChan
	helpChangeChan chan HelpInfo
	sectorChangeChan SectorChan
	mixedActionChangeChan chan MixedActionInfo
)


var (
	admins = [...]int{47700349}
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

const (
	EnAddress = "http://quest.ua/%s"
	//EnAddress = "http://demo.en.cx/%s"
	LoginEndpoint = "login/signin?json=1"
	LevelInfoEndpoint = "GameEngines/Encounter/Play/%d?json=1"
	SendCodeEndpoint
	SendBonusCodeEndpoint
)

const (
	//CoordinateLink = `<a href="https://maps.google.com/maps?daddr=%v&saddr=My+Location">%v</a>`
	//CoordinateLink = `<a href="comgooglemapurl://maps.google.com/maps?daddr=%v&saddr=My+Location">%v</a>`
	//CoordinateLink = `comgooglemapsurl://maps.google.com/maps?daddr=%v&saddr=My+Location`
	CoordinateLink = `[%v](http://maps.google.com/maps?daddr=%v&saddr=My+Location)`
	//CoordinateLink = `[%v](comgooglemapsurl://maps.google.com/maps?daddr=%v&saddr=My+Location)`
)

const (
	LevelInfoString = `
*Номер уровня:* %d
*Название уровня:* %q
*Времени на уровень:* %s
*Автопереход через:* %s
*Задание:*
%s
`
	HelpInfoString = `
*Подсказка:* %d
*Текст:* %s`
	//MixedActionInfoString = `
//*%s* вбил код *%q*.`
	CorrectAnswerString = `*+* %q *%s*`
	IncorrectAnswerString = `*-* %q *%s*`
	SectorInfoString = `
	Сектор *%q* закрыт. Осталось %d из %d`
)

//const (
//	enAPI *EnAPI = EnAPI{"Harry_Potter", "toknkpils85", new(http.Client)}
//)

func (en *EnAPI) MakeRequest(endpoint string, params url.Values) (EnAPIAuthResponse, error) {
	var enUrl string = fmt.Sprintf(EnAddress, LoginEndpoint)

	resp, err := en.Client.PostForm(enUrl, params)
	if err != nil {
		fmt.Print("Exit 1")
		return EnAPIAuthResponse{}, err
	}

	var apiResp EnAPIAuthResponse
	apiResp.CreateFromResponse(resp)

	return apiResp, nil
}

func parseLevelJson(body io.ReadCloser) (*LevelResponse, error) {
	var (
		lvl *LevelResponse
		err error
	)
	defer body.Close()
	respBody, _ := ioutil.ReadAll(body)
	lvl = new(LevelResponse)
	err = json.Unmarshal(respBody, &lvl)
	if err != nil {
		log.Println("Error:", err)
		return &LevelResponse{}, err
	}

	return lvl, nil
}

func (en *EnAPI) Login() (EnAPIAuthResponse, error) {
	var (
		authResponse EnAPIAuthResponse
		err error
		params url.Values
	)
	params = make(url.Values)
	params.Set("Login", en.login)
	params.Set("Password", en.password)

	authResponse, err = en.MakeRequest(fmt.Sprintf(EnAddress, LoginEndpoint), params)
	if err != nil {
		log.Print(err)
		return EnAPIAuthResponse{}, nil
	}
	return authResponse, err
}

func (en *EnAPI) GetLevelInfo() (*LevelResponse, error) {
	//gameUrl := "http://demo.en.cx/GameEngines/Encounter/Play/25733?json=1"
	var (
		gameUrl string = fmt.Sprintf(EnAddress, fmt.Sprintf(LevelInfoEndpoint, en.CurrentGameId))
		lvl *LevelResponse
		err error
	)

	resp, err := en.Client.Get(gameUrl)
	if err != nil {
		log.Println("Error on GET request:", err)
		return &LevelResponse{}, err
	}

	if strings.HasPrefix(resp.Header["Content-Type"][0], "text/html") {
		log.Println("Incorrect cookies, need to re-login")
		return &LevelResponse{}, errors.New("Incorrect cookies, need to re-login")
	}

	lvl, err = parseLevelJson(resp.Body)

	return lvl, err
}

type sendCodeResponse struct {

}

func (en *EnAPI) SendCode(gameId int32, code string) (*LevelResponse, error) {
	var (
		codeUrl string = fmt.Sprintf(EnAddress, fmt.Sprintf(SendCodeEndpoint, gameId))
		resp *http.Response
		body SendCodeRequest
		lvl *LevelResponse
		bodyJson []byte
		err error
	)
	body = SendCodeRequest{
		codeRequest: codeRequest{
			LevelId: 249435,
			LevelNumber: 3},
		LevelAction: code,
	}
	bodyJson, err = json.Marshal(body)
	if err != nil {
		log.Println("Error while serializing body:", err)
		return nil, err
	}

	resp, err = en.Client.Post(codeUrl, "application/json", bytes.NewBuffer(bodyJson))
	if err != nil {
		log.Println("Error while preforming request:", err)
		return nil, err
	}

	lvl, err = parseLevelJson(resp.Body)
	return lvl, err
}

func (en *EnAPI) SendBonusCode() {

}

func isAdmin(id int) bool {
	for _, userId := range(admins) {
		if userId == id {
			return true
		}
	}
	return false
}

type Coordinate struct {
	lon            float64
	lat            float64
	originalString string
}
type Coordinates []Coordinate
func (c Coordinate) String() (text string) {
	text = fmt.Sprintf("%f,%f", c.lon, c.lat)
	return
}

func replaceCoordinates(text string) string {
	//fmt.Printf("%v", Coordinate{lon:1.23, lat:0.234})
	var (
		re  *regexp.Regexp = regexp.MustCompile("(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,})")
		mr  [][]string = re.FindAllStringSubmatch(text, -1)
		res []byte = make([]byte, len(text))
	)
	copy(res, []byte(text))
	if len(mr) > 0 {
		coords := make(Coordinates, len(mr), len(mr))
		for i, item := range re.FindAllStringSubmatch(text, -1) {
			lon, _ := strconv.ParseFloat(item[1], 64)
			lat, _ := strconv.ParseFloat(item[2], 64)
			coords[i] = Coordinate{lon: lon, lat: lat, originalString: item[0]}
			//text = re.ReplaceAllString(text, fmt.Sprintf(CoordinateLink, coord, coord))
			res = regexp.MustCompile(coords[i].originalString).
				ReplaceAllLiteral(res, []byte(fmt.Sprintf(CoordinateLink, coords[i], coords[i])))
		}

		return string(res)
	}
	return text
}

type Image struct {
	url     string
	caption string
}

func replaceImages(text string) string {
	var (
		re  *regexp.Regexp = regexp.MustCompile("<img.+?src=\"(https?://.+?)\">")
		mr  [][]string = re.FindAllStringSubmatch(text, -1)
		res []byte = make([]byte, len(text))
	)
	copy(res, []byte(text))
	if len(mr) > 0 {
		for i, item := range re.FindAllStringSubmatch(text, -1){
			img := Image{url: item[1], caption: fmt.Sprintf("Картинка #%d", i+1)}
			res = regexp.MustCompile(item[0]).
				ReplaceAllLiteral(res, []byte(fmt.Sprintf("[%s](%s)", img.caption, img.url)))
			go sendImageFromUrl(&img)
		}
		return string(res)
	}
	return text
}

func sendImageFromUrl(img *Image) {
	var (
		file *os.File
	)
	resp, err := http.Get(img.url)
	if err != nil {
		log.Println("Can't download image:", err)
	}

	defer resp.Body.Close()
	filename := fmt.Sprintf("/tmp/%s", path.Base(img.url))
	fileInfo, err := os.Stat(filename)
	if os.IsExist(err) && fileInfo.Size() > 0 {
			file, err = os.Open(filename)
	} else {
		log.Println("Image is not downloaded yet:", err)
		file, err = os.Create(filename)
		if err != nil {
			log.Fatal("Cannot create file:", err)
			return
		}
	}
	defer file.Close()
	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	msg := tgbotapi.NewPhotoUpload(chatId, filename)
	msg.Caption = strings.ToUpper(img.caption)
	log.Println("Sending message with photo to the channel")
	//return msg
	photoChan <- msg
	//bot.Send(msg)
}

func replaceCommonTags(text string) string {
	var (
		reBold   *regexp.Regexp = regexp.MustCompile("<b>(.+?)</b>")
		mrBold   [][]string = reBold.FindAllStringSubmatch(text, -1)
		reItalic *regexp.Regexp = regexp.MustCompile("<i>(.+?)</i>")
		mrItalic [][]string = reItalic.FindAllStringSubmatch(text, -1)
		reFont   *regexp.Regexp = regexp.MustCompile("<font.+?color=\"(\\w+)\".*?>(.+?)</font>")
		mrFont   [][]string = reFont.FindAllStringSubmatch(text, -1)
		reA      *regexp.Regexp = regexp.MustCompile("<a.+?href=\"(.+?)\".*?>(.+?)</a>")
		mrA      [][]string = reA.FindAllStringSubmatch(text, -1)
		res      []byte = make([]byte, len(text))
	)

	copy(res, []byte(text))
	if len(mrBold) > 0 {
		for _, item := range mrBold {
			res = regexp.MustCompile(item[0]).ReplaceAllLiteral(res, []byte(fmt.Sprintf("*%s*", item[1])))
		}
	}
	if len(mrItalic) > 0 {
		for _, item := range mrItalic {
			res = regexp.MustCompile(item[0]).ReplaceAllLiteral(res, []byte(fmt.Sprintf("_%s_", item[1])))
		}
	}
	if len(mrFont) > 0 {
		for _, item := range mrFont {
			res = regexp.MustCompile(item[0]).
				ReplaceAllLiteral(res, []byte(fmt.Sprintf("#%s#%s#", item[1], item[2])))
		}
	}
	if len(mrA) > 0 {
		for _, item := range mrA {
			res = regexp.MustCompile(item[0]).
				ReplaceAllLiteral(res, []byte(fmt.Sprintf("[%s](%s)", item[2], item[1])))
		}
	}
	res = regexp.MustCompile("</?p>").ReplaceAllLiteral(res, []byte(""))
	return string(res)
}

func sendMessage(bot *tgbotapi.BotAPI, chatId int64, text string) {
	msg := tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "Markdown"
	//msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	bot.Send(msg)
}

func processBotCommand(m *tgbotapi.Message, en *EnAPI, bot *tgbotapi.BotAPI) error {
	//var commandReader bufio.Reader
	var (
		command []byte
		commandCode BotCommand
		ok bool
		ent *tgbotapi.MessageEntity
	)

	ent = GetBotCommandEntity(m)
	if ent == nil {
		return errors.New("No bot command found")
	}
	command = make([]byte, ent.Length-1, ent.Length-1)
	copy(command, m.Text[ent.Offset+1:ent.Length])
	if commandCode, ok = BotCommandDict[string(command)]; !ok {
		return errors.New("Unknown command")
	}

	switch commandCode{
	case InfoCommand:
		var res, text string
		if en.CurrentLevel == nil {
			lvl, _ := en.GetLevelInfo()
			en.CurrentLevel = lvl.Level
		}
		log.Println("TIMEOUT:", uint64(en.CurrentLevel.Timeout))
		//res = regexp.MustCompile("_").ReplaceAllLiteralString(res, ("\\__"))
		res = replaceCoordinates(en.CurrentLevel.Tasks[0].TaskText)
		res = replaceImages(res)
		res = replaceCommonTags(res)
		text = fmt.Sprintf(LevelInfoString, en.CurrentLevel.Number, en.CurrentLevel.Name,
			PrettyTimePrint(en.CurrentLevel.Timeout),
			PrettyTimePrint(en.CurrentLevel.TimeoutSecondsRemain),
			res)
		                   //(en.CurrentLevel.TimeoutSecondsRemain*time.Second).String())
		if len(text) > 4096 {
			fullMessageCount := len(text) / 4096
			for i := 0; i < fullMessageCount; i++ {
				sendMessage(bot, m.Chat.ID, text[i*4096:(i+1)*4096])
			}
			sendMessage(bot, m.Chat.ID, text[fullMessageCount*4096:])
		} else {
			sendMessage(bot, m.Chat.ID, text)
		}
		//lvl, _ := en.GetLevelInfo()
	case SetChatIdCommand:
		if isAdmin(m.From.ID){
			log.Println("Set chat ID as working chat")
			chatId = m.Chat.ID
			text := fmt.Sprintf("Теперь буду работать только с чатом %q, остальные буду игнорировать",
				m.Chat.Title)
			messageChan <- tgbotapi.NewMessage(chatId, text)
		} else {
			text := fmt.Sprint("Выбирать чат могут только администраторы")
			messageChan <- tgbotapi.NewMessage(m.Chat.ID, text)
		}
	case WatchCommand:
		ticker := time.NewTicker(2*time.Second)
		quit = make(chan struct{})
		go func() {
			for {
				select{
				case <- ticker.C:
					lvl, err := en.GetLevelInfo()
					//log.Println("Error:", err)
					//en.GetLevelInfo()
					if err != nil {
						log.Println("Error:", err)
						continue
					}
					levelChan <- *(lvl.Level)
				case <- quit:
					ticker.Stop()
				}
			}
		}()
	case StopWatchingCommand:
		close(quit)
	case TestHelpChange:
		log.Println("Command help change")
		helpChangeChan <- HelpInfo{}

	//if en.CurrentLevel == nil {
	//	en.CurrentLevel = lvl
	//	return lvl, err
	//}
	//
	//if en.CurrentLevel.Number < lvl.Level.Number {
	//	en.Levels.InsertAfter(en.CurrentLevel, en.Levels.Front())
	//	en.CurrentLevel = lvl.Level
	//} else {
	//	//en.
	//}

	}
	return nil
}

func helpChangeHandler(hi *HelpInfo) (msg tgbotapi.MessageConfig) {
	log.Println("Help is changed")
	var text string = fmt.Sprintf(HelpInfoString, hi.Number, hi.HelpText)
	msg = tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "HTML"
	return
	//msg := tgbotapi.NewMessage(m.Chat.ID, text)
	//msg.ParseMode = "Markdown"
	//bot.Send(msg)
}

func mixedActionChangeHandler(mai *MixedActionInfo) (msg tgbotapi.MessageConfig) {
	log.Println("New MixedAction is added")
	var text string
	if mai.IsCorrect {
		text = fmt.Sprintf(CorrectAnswerString, mai.Answer, mai.Login)
	} else {
		text = fmt.Sprintf(IncorrectAnswerString, mai.Answer, mai.Login)
	}
	msg = tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "Markdown"
	return
}

func sectorChangeHandler(esi *ExtendedSectorInfo) (msg tgbotapi.MessageConfig){
	log.Println("Some sector is changed")
	var text string = fmt.Sprintf(SectorInfoString, esi.sectorInfo.Name, esi.sectorsLeft, esi.totalSectors)
	msg = tgbotapi.NewMessage(chatId, text)
	msg.ParseMode = "Markdown"
	return
}

func levelChangeHandler(li *LevelInfo) (msg tgbotapi.MessageConfig) {
	log.Println("Level is changed")
	//var text string = fmt.Sprintf(LevelInfoString)
	return
}

func processChanges(en *EnAPI) {
	for {
		select {
		case hi := <- helpChangeChan:
			msg := helpChangeHandler(&hi)
			messageChan <- msg
		case mai := <- mixedActionChangeChan:
			msg := mixedActionChangeHandler(&mai)
			messageChan <- msg
		case si := <- sectorChangeChan:
			msg := sectorChangeHandler(&si)
			messageChan <- msg
		case li := <- levelChangeChan:
			msg := levelChangeHandler(&li)
			messageChan <- msg
		case level := <- levelChan: {
			log.Println("Level is changed")
			if isNewLevel(en.CurrentLevel, &level) {
				log.Println("Level is new")
				levelChangeChan <- level
				continue
			}
			log.Println("Level is not new")
			go checkHelps(*en.CurrentLevel, level)
			go checkMixedActions(*en.CurrentLevel, level)
			go checkSectors(*en.CurrentLevel, level)
			en.CurrentLevel = &level
			//time.Sleep(5*time.Second)
		}
		//default:
		//	log.Println("No changes")
		}
	}
}

func isNewLevel(oldLevel *LevelInfo, newLevel *LevelInfo) bool {
	return oldLevel.LevelId != newLevel.LevelId
}

func checkHelps(oldLevel LevelInfo, newLevel LevelInfo) {
	log.Println("Start checking changes in Helps section")
	for i, _ := range oldLevel.Helps {
		if oldLevel.Helps[i].Number == newLevel.Helps[i].Number {
			if oldLevel.Helps[i].HelpText != newLevel.Helps[i].HelpText {
				helpChangeChan <- newLevel.Helps[i]
			}
		}
	}
	log.Println("Finish checking changes in Helps section")
}

func checkSectors(oldLevel LevelInfo, newLevel LevelInfo) {
	log.Println("Start checking changes in Sectors section")
	for i, _ := range oldLevel.Sectors {
		if oldLevel.Sectors[i].Name == newLevel.Sectors[i].Name {
			if oldLevel.Sectors[i].IsAnswered != newLevel.Sectors[i].IsAnswered {
				sectorChangeChan <- ExtendedSectorInfo{
					sectorInfo: &newLevel.Sectors[i],
					sectorsLeft: newLevel.SectorsLeftToClose,
					sectorsPassed: newLevel.PassedSectorsCount,
					totalSectors: int8(len(newLevel.Sectors))}
			}
		}
	}
	log.Println("Finish checking changes in Sectors section")
}

func checkMixedActions(oldLevel LevelInfo, newLevel LevelInfo) {
	log.Println("Start checking changes in MixedActions section")
	sort.Sort(newLevel.MixedActions)
	if len(oldLevel.MixedActions) < len(newLevel.MixedActions) {
		for i := len(oldLevel.MixedActions); i < len(newLevel.MixedActions); i++ {
			mixedActionChangeChan <- newLevel.MixedActions[i]
		}
	}
	log.Println("Finish checking changes in MixedActions section")
}

func main() {
	bot, err := tgbotapi.NewBotAPI("<TOKEN>")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	jar, _ := cookiejar.New(nil)
	en := &EnAPI{
		login: "<NICK>",
		password: "<PASSWORD>",
		Client: &http.Client{Jar: jar},
		CurrentGameId: 56326,
		CurrentLevel: nil,
		Levels: list.New()}
	en.Login()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, _ := bot.GetUpdatesChan(u)

	helpChangeChan = make(chan HelpInfo)
	messageChan = make(MessageChan)
	photoChan = make(PhotoChan)
	levelChan = make(chan LevelInfo)
	levelChangeChan = make(LevelChan)
	sectorChangeChan = make(SectorChan)
	mixedActionChangeChan = make(chan MixedActionInfo)


	go processChanges(en)
	go func(bot *tgbotapi.BotAPI) {
		log.Println("Read message from channel to send to chat", chatId)
		for {
			if chatId != 0 {
				select {
				case msg := <- messageChan:
					//msg := tgbotapi.NewMessage(chatId, text)
					//msg.ParseMode = "Markdown"
					bot.Send(msg)
				case ph := <- photoChan:
					log.Println("Sending photo to the chat")
					bot.Send(ph)
				}
			}
		}
	}(bot)

	for update := range updates {
		if update.Message == nil {
		    continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if IsMessageBotCommand(update.Message) {
			log.Println("It is bot command")
			processBotCommand(update.Message, en, bot)
		}
		//switch update.Message.Text

		//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		//msg.ReplyToMessageID = update.Message.MessageID

		//bot.Send(msg)
	}

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
