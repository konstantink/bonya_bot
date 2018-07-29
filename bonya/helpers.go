package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/bonya_bot/en"
	"github.com/tucnak/telebot"
)

type EnvConfig struct {
	BotToken     string `envconfig:"bot_token"`
	GameID       int32  `envconfig:"game_id"`
	EngineDomain string `envconfig:"engine_domain"`
	MainChat     int64  `envconfig:"main_chat"`
	User         string
	Password     string
}

type BotMessage struct {
	msg string
}

func NewBotMessage(msg string) *BotMessage {
	return &BotMessage{msg}
}

func (bm *BotMessage) ToText() string {
	return bm.msg
}

func (bm *BotMessage) ReplyTo() (message telebot.Message) {
	return
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func ReplaceImages(text string, caption string) (string, en.Images) {
	log.Print("Replace images in task text")
	var (
		re     = regexp.MustCompile("<img.+?src=\"\\s*(https?://.+?)\\s*\".*?>")
		reA    = regexp.MustCompile("<a.+?href=\\\\?\"(https?://.+?\\.(jpg|png|bmp))\\\\?\".*?>(.*?)</a>")
		mr     = re.FindAllStringSubmatch(text, -1)
		mrA    = reA.FindAllStringSubmatch(text, -1)
		result = text
		images = make(en.Images, 0)
	)
	//log.Printf("Before image replacing: %s", text)
	if len(mr) > 0 {
		//copy(result, []byte(text))
		for i, item := range mr {
			images = append(images, en.Image{URL: item[1], Caption: fmt.Sprintf("%s #%d", caption, i+1)})
			result = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(result, fmt.Sprintf("%s #%d", caption, i+1))
			//ReplaceAllLiteralString(result, fmt.Sprintf("[%s #%d](%s)", caption, i+1, item[1]))
		}
		//log.Printf("After image replacing: %s", text)
		return result, images
	}
	if len(mrA) > 0 {
		for i, item := range mrA {
			images = append(images, en.Image{URL: item[1], Caption: fmt.Sprintf("%s #%d", caption, i+1)})
			result = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(result, fmt.Sprintf("%s #%d", caption, i+1))
			//ReplaceAllLiteralString(result, fmt.Sprintf("[%s #%d](%s)", caption, i+1, item[1]))
		}
		//log.Printf("After image replacing: %s", text)
		return result, images
	}
	return result, images
}

func ExtractImages(text string, caption string) (images en.Images) {
	var (
		re  = regexp.MustCompile("<img.+?src=\\\\?\"(https?://.+?)\\\\?\".*?>")
		reA = regexp.MustCompile("<a.+?href=\\\\?\"(https?://.+?\\.(jpg|png|bmp))\\\\?\".*?>")
		mr  = re.FindAllStringSubmatch(text, -1)
		mrA = reA.FindAllStringSubmatch(text, -1)
	)
	images = make(en.Images, 0)
	if len(mr) > 0 {
		for i, item := range mr {
			images = append(images, en.Image{URL: item[1], Caption: fmt.Sprintf("%s #%d", caption, i+1)})
		}
	} else if len(mrA) > 0 {
		for i, item := range mrA {
			images = append(images, en.Image{URL: item[1], Caption: fmt.Sprintf("%s #%d", caption, i+1)})
		}
	}
	return
}

//func GetBotCommandEntity(m *tgbotapi.Message) *tgbotapi.MessageEntity {
//	for _, entity := range *m.Entities {
//		if entity.Type == "bot_command" {
//			return &entity
//		}
//	}
//	return nil
//}

func isNewLevel(oldLevel *en.Level, newLevel *en.Level) bool {
	if oldLevel == nil {
		return true
	}
	return oldLevel.LevelID != newLevel.LevelID
}

// SplitText splits the input text into chunks of size `size`
func SplitText(text string, size int) (result []string) {
	if len(text) > size {
		var lIndex = 0
		for {
			spaceIndex := strings.LastIndex(text[lIndex:lIndex+size+1], " ")
			result = append(result, text[lIndex:lIndex+spaceIndex])
			lIndex = lIndex + spaceIndex + 1
			if len(text[lIndex:]) < 4096 {
				result = append(result, text[lIndex:])
				break
			}
		}
	} else {
		result = append(result, text)
	}
	return
}
