package main

import (
	"bytes"
	"fmt"
	"github.com/tucnak/telebot"
	"gopkg.in/telegram-bot-api.v4"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"time"
)

type EnvConfig struct {
	BotToken string `envconfig:"bot_token"`
	User     string
	Password string
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

type Image struct {
	url     string
	caption string
}

func ReplaceCoordinates(text string) string {
	//fmt.Printf("%v", Coordinate{lon:1.23, lat:0.234})
	var (
		re  *regexp.Regexp = regexp.MustCompile("(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,})")
		mr  [][]string     = re.FindAllStringSubmatch(text, -1)
		res []byte         = make([]byte, len(text))
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

func ReplaceImages(text string) (string, []Image) {
	var (
		re     *regexp.Regexp = regexp.MustCompile("<img.+?src=\"(https?://.+?)\">")
		mr     [][]string     = re.FindAllStringSubmatch(text, -1)
		res    []byte         = make([]byte, len(text))
		images []Image        = make([]Image, 0)
	)
	copy(res, []byte(text))
	if len(mr) > 0 {
		for i, item := range re.FindAllStringSubmatch(text, -1) {
			img := Image{url: item[1], caption: fmt.Sprintf("Картинка #%d", i+1)}
			res = regexp.MustCompile(item[0]).
				ReplaceAllLiteral(res, []byte(fmt.Sprintf("[%s](%s)", img.caption, img.url)))
			images = append(images, img)
		}
		return string(res), images
	}
	return text, images
}

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
		//telebotFile, _ := telebot.NewFile(file.Name())
		log.Println("Sending message with photo to the channel")
		//thumbnail := telebot.Thumbnail{File: telebotFile, Width: 120, Height: 120}
		//photoInfoChan <- &PhotoInfo{Recepient: recepient, Photo: &telebot.Photo{File: telebotFile,
		//	Thumbnail: thumbnail, Caption: img.caption}, Options: nil}
	}
}

func ReplaceCommonTags(text string) string {
	var (
		reBold   *regexp.Regexp = regexp.MustCompile("<b>(.+?)</b>")
		mrBold   [][]string     = reBold.FindAllStringSubmatch(text, -1)
		reItalic *regexp.Regexp = regexp.MustCompile("<i>(.+?)</i>")
		mrItalic [][]string     = reItalic.FindAllStringSubmatch(text, -1)
		reFont   *regexp.Regexp = regexp.MustCompile("<font.+?color=\"(\\w+)\".*?>(.+?)</font>")
		mrFont   [][]string     = reFont.FindAllStringSubmatch(text, -1)
		reA      *regexp.Regexp = regexp.MustCompile("<a.+?href=\"(.+?)\".*?>(.+?)</a>")
		mrA      [][]string     = reA.FindAllStringSubmatch(text, -1)
		res      []byte         = make([]byte, len(text))
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

func GetBotCommandEntity(m *tgbotapi.Message) *tgbotapi.MessageEntity {
	for _, entity := range *m.Entities {
		if entity.Type == "bot_command" {
			return &entity
		}
	}
	return nil
}

func isNewLevel(oldLevel *LevelInfo, newLevel *LevelInfo) bool {
	if oldLevel == nil {
		return true
	}
	return oldLevel.LevelId != newLevel.LevelId
}

//func IsBotCommand(m *tgbotapi.Message) bool {
//	for _, entity := range(*m.Entities)  {
//		if (entity.Type == "bot_command") {
//			return true
//		}
//	}
//	return false
//}

func PrettyTimePrint(d time.Duration) (res *bytes.Buffer) {
	//var correctTime = d*time.Second
	var s string
	res = bytes.NewBufferString(s)
	//defer res.Close()
	if (d / 3600) > 0 {
		//res.WriteString(fmt.Sprintf("%d часов ", d/3600))
		switch d / 3600 {
		case 1, 21, 31, 41, 51:
			res.WriteString(fmt.Sprintf("%d час ", d/3600))
		case 2, 3, 4, 22, 23, 24, 32, 33, 34, 42, 43, 44, 52, 53, 54:
			res.WriteString(fmt.Sprintf("%d часа ", d/3600))
		default:
			res.WriteString(fmt.Sprintf("%d часов ", d/3600))
		}
	}
	if (d/60)%60 > 0 {
		switch (d / 60) % 60 {
		case 1, 21, 31, 41, 51:
			res.WriteString(fmt.Sprintf("%d минуту ", (d/60)%60))
		case 2, 3, 4, 22, 23, 24, 32, 33, 34, 42, 43, 44, 52, 53, 54:
			res.WriteString(fmt.Sprintf("%d минуты ", (d/60)%60))
		default:
			res.WriteString(fmt.Sprintf("%d минут ", (d/60)%60))
		}

	}
	if d%60 > 0 {
		switch d % 60 {
		case 1, 21, 31, 41, 51:
			res.WriteString(fmt.Sprintf("%d секунду", d%60))
		case 2, 3, 4, 22, 23, 24, 32, 33, 34, 42, 43, 44, 52, 53, 54:
			res.WriteString(fmt.Sprintf("%d секунды", d%60))
		default:
			res.WriteString(fmt.Sprintf("%d секунд", d%60))
		}
	}
	return
}
