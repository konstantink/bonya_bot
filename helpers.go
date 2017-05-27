package main

import (
	"bytes"
	"fmt"
	//"github.com/tucnak/telebot"
	//"gopkg.in/telegram-bot-api.v4"
	//"io"
	"log"
	//"net/http"
	//"os"
	//"path"
	"regexp"
	"strconv"
	"time"
	//"io/ioutil"
	"github.com/tucnak/telebot"
	"strings"
)

type EnvConfig struct {
	BotToken     string `envconfig:"bot_token"`
	GameId       int32  `envconfig:"game_id"`
	EngineDomain string `envconfig:"engine_domain"`
	MainChat     int64  `envconfig:"main_chat"`
	User         string
	Password     string
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

type Images []Image

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

func ReplaceCoordinates(text string) (string, Coordinates) {
	var (
		// <a href="geo:49.976136, 36.267256">49.976136, 36.267256</a>
		geoHrefRe *regexp.Regexp = regexp.MustCompile("<a.+?href=\"geo:(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,})\">(.+?)</a>")
		// <a href="https://www.google.com.ua/maps/@50.0363257,36.2120039,19z" target="blank">50.036435 36.211914</a>
		hrefRe *regexp.Regexp = regexp.MustCompile("<a.+?href=\"https?://.+?(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,}).*?\">(.+?)</a>")
		// 49.976136, 36.267256
		numbersRe *regexp.Regexp = regexp.MustCompile("(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,})")

		mr     [][]string
		res    string      = text
		coords Coordinates = make(Coordinates, 0, 0)
	)

	log.Print("Replace coordinates in task")
	mr = geoHrefRe.FindAllStringSubmatch(res, -1)
	if len(mr) > 0 {
		for _, item := range mr {
			lon, _ := strconv.ParseFloat(item[1], 32)
			lat, _ := strconv.ParseFloat(item[2], 32)
			coords = append(coords, Coordinate{lon: lon, lat: lat, originalString: item[3]})
			res = regexp.MustCompile(item[0]).ReplaceAllLiteralString(res, "#coords#")
		}
	}

	mr = hrefRe.FindAllStringSubmatch(res, -1)
	if len(mr) > 0 {
		for _, item := range mr {
			lon, _ := strconv.ParseFloat(item[1], 32)
			lat, _ := strconv.ParseFloat(item[2], 32)
			coords = append(coords, Coordinate{lon: lon, lat: lat, originalString: item[3]})
			res = regexp.MustCompile(item[0]).ReplaceAllLiteralString(res, "#coords#")
		}
	}

	mr = numbersRe.FindAllStringSubmatch(res, -1)
	if len(mr) > 0 {
		for _, item := range mr {
			lon, _ := strconv.ParseFloat(item[1], 32)
			lat, _ := strconv.ParseFloat(item[2], 32)
			coords = append(coords, Coordinate{lon: lon, lat: lat, originalString: item[0]})
			res = regexp.MustCompile(item[0]).ReplaceAllLiteralString(res, "#coords#")
		}
	}

	for _, coord := range coords {
		res = strings.Replace(res, "#coords#", coord.originalString, 1)
	}

	return res, coords
}

//func ReplaceCoordinates(text string) string {
//	//fmt.Printf("%v", Coordinate{lon:1.23, lat:0.234})
//	log.Print("Replace coordinates in task")
//	var (
//		numbersRe  *regexp.Regexp = regexp.MustCompile("[^@](\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,})")
//		hrefRe     *regexp.Regexp = regexp.MustCompile("<a.+?href=\"geo:(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,})\">(.+?)</a>")
//		numbersMr  [][]string     = numbersRe.FindAllStringSubmatch(text, -1)
//		hrefMr     [][]string     = hrefRe.FindAllStringSubmatch(text, -1)
//		res        string         = text
//		mr         [][]string
//	)
//
//	if len(hrefMr) > 0 {
//		mr = hrefMr
//	} else if len(numbersMr) > 0 {
//		mr = numbersMr
//	} else {
//		return res
//	}
//
//	coords := make(Coordinates, len(mr), len(mr))
//	for i, item := range mr {
//		lon, _ := strconv.ParseFloat(item[1], 64)
//		lat, _ := strconv.ParseFloat(item[2], 64)
//		coords[i] = Coordinate{lon: lon, lat: lat, originalString: item[0]}
//		res = regexp.MustCompile(coords[i].originalString).
//			ReplaceAllLiteralString(res, fmt.Sprintf(CoordinateLink, coords[i], coords[i]))
//	}
//
//	return res
//}

//func ReplaceImages(text string) (string, []Image) {
//	log.Print("Replace images in task")
//	var (
//		re     *regexp.Regexp = regexp.MustCompile("<img.+?src=\"(https?://.+?)\".*>")
//		mr     [][]string     = re.FindAllStringSubmatch(text, -1)
//		res    []byte         = make([]byte, len(text))
//		images []Image        = make([]Image, 0)
//	)
//	copy(res, []byte(text))
//	if len(mr) > 0 {
//		for i, item := range re.FindAllStringSubmatch(text, -1) {
//			img := Image{url: item[1], caption: fmt.Sprintf("Картинка #%d", i+1)}
//			res = regexp.MustCompile(item[0]).
//				ReplaceAllLiteral(res, []byte(fmt.Sprintf("[%s](%s)", img.caption, img.url)))
//			images = append(images, img)
//		}
//		return string(res), images
//	}
//	return text, images
//}

func BlockTypeToString(typeId int8) string {
	if typeId == 0 || typeId == 1 {
		return "Игрок"
	}
	return "Команда"
}

func ReplaceImages(text string, caption string) (string, Images) {
	log.Print("Replace images in task text")
	var (
		re     *regexp.Regexp = regexp.MustCompile("<img.+?src=\"\\s*(https?://.+?)\\s*\".*?>")
		reA    *regexp.Regexp = regexp.MustCompile("<a.+?href=\\\\?\"(https?://.+?\\.(jpg|png|bmp))\\\\?\".*?>(.*?)</a>")
		mr     [][]string     = re.FindAllStringSubmatch(text, -1)
		mrA    [][]string     = reA.FindAllStringSubmatch(text, -1)
		result string         = text
		images Images         = make(Images, 0)
	)
	//log.Printf("Before image replacing: %s", text)
	if len(mr) > 0 {
		//copy(result, []byte(text))
		for i, item := range mr {
			images = append(images, Image{url: item[1], caption: fmt.Sprintf("%s #%d", caption, i+1)})
			result = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(result, fmt.Sprintf("%s #%d", caption, i+1))
				//ReplaceAllLiteralString(result, fmt.Sprintf("[%s #%d](%s)", caption, i+1, item[1]))
		}
		//log.Printf("After image replacing: %s", text)
		return result, images
	}
	if len(mrA) > 0 {
		for i, item := range mrA {
			images = append(images, Image{url: item[1], caption: fmt.Sprintf("%s #%d", caption, i+1)})
			result = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(result, fmt.Sprintf("%s #%d", caption, i+1))
				//ReplaceAllLiteralString(result, fmt.Sprintf("[%s #%d](%s)", caption, i+1, item[1]))
		}
		//log.Printf("After image replacing: %s", text)
		return result, images
	}
	return result, images
}

func ExtractImages(text string, caption string) (images []Image) {
	var (
		re  *regexp.Regexp = regexp.MustCompile("<img.+?src=\\\\?\"(https?://.+?)\\\\?\".*?>")
		reA *regexp.Regexp = regexp.MustCompile("<a.+?href=\\\\?\"(https?://.+?\\.(jpg|png|bmp))\\\\?\".*?>")
		mr  [][]string     = re.FindAllStringSubmatch(text, -1)
		mrA [][]string     = reA.FindAllStringSubmatch(text, -1)
	)
	images = make([]Image, 0)
	if len(mr) > 0 {
		for i, item := range mr {
			images = append(images, Image{url: item[1], caption: fmt.Sprintf("%s #%d", caption, i+1)})
		}
	} else if len(mrA) > 0 {
		for i, item := range mrA {
			images = append(images, Image{url: item[1], caption: fmt.Sprintf("%s #%d", caption, i+1)})
		}
	}
	return
}

func ReplaceCommonTags(text string) string {
	log.Print("Replace html tags")
	var (
		reBr     *regexp.Regexp = regexp.MustCompile("<br\\s*/?>")
		reHr     *regexp.Regexp = regexp.MustCompile("<hr.*?/?>")
		reP      *regexp.Regexp = regexp.MustCompile("<p>([^ ]+?)</p>")
		reBold   *regexp.Regexp = regexp.MustCompile("<b.*?/?>((?s:.*?))</b>")
		reStrong *regexp.Regexp = regexp.MustCompile("<strong.*?>(.*?)</strong>")
		reItalic *regexp.Regexp = regexp.MustCompile("<i>((?s:.+?))</i>")
		reSpan   *regexp.Regexp = regexp.MustCompile("<span.*?>(.*?)</span>")
		reCenter *regexp.Regexp = regexp.MustCompile("<center>(.*?)</center>")
		reFont   *regexp.Regexp = regexp.MustCompile("<font.+?color\\s*=\\\\?[\"«]?#?(\\w+)\\\\?[\"»]?.*?>((?s:.*?))</font>")
		reA      *regexp.Regexp = regexp.MustCompile("<a.+?href=\\\\?\"(.+?)\\\\?\".*?>(.+?)</a>")
		res      string         = text
	)

	res = strings.Replace(text, "_", "\\_", -1)
	if mrBr := reBr.FindAllStringSubmatch(text, -1); len(mrBr) > 0 {
		for _, item := range mrBr {
			res = regexp.MustCompile(item[0]).ReplaceAllLiteralString(res, "\n")
		}
	}
	if mrHr := reHr.FindAllStringSubmatch(res, -1); len(mrHr) > 0 {
		for _, item := range mrHr {
			res = regexp.MustCompile(item[0]).ReplaceAllLiteralString(res, "\n")
		}
	}
	if mrP := reP.FindAllStringSubmatch(res, -1); len(mrP) > 0 {
		for _, item := range mrP {
			res = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(res, fmt.Sprintf("\n%s", item[1]))
		}
	}
	if mrFont := reFont.FindAllStringSubmatch(res, -1); len(mrFont) > 0 {
		for _, item := range mrFont {
			res = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(res, fmt.Sprintf("%s", item[2]))
			//ReplaceAllLiteral(res, []byte(fmt.Sprintf("#%s#%s#", item[1], item[2])))
		}
	}
	if mrBold := reBold.FindAllStringSubmatch(res, -1); len(mrBold) > 0 {
		for _, item := range mrBold {
			res = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(res, fmt.Sprintf("*%s*", item[1]))
		}
	}
	if mrStrong := reStrong.FindAllStringSubmatch(res, -1); len(mrStrong) > 0 {
		for _, item := range mrStrong {
			res = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(res, fmt.Sprintf("*%s*", item[1]))
		}
	}
	if mrItalic := reItalic.FindAllStringSubmatch(res, -1); len(mrItalic) > 0 {
		for _, item := range mrItalic {
			res = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(res, fmt.Sprintf("_%s_", item[1]))
		}
	}
	if mrSpan := reSpan.FindAllStringSubmatch(res, -1); len(mrSpan) > 0 {
		for _, item := range mrSpan {
			res = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(res, item[1])
		}
	}
	if mrCenter := reCenter.FindAllStringSubmatch(res, -1); len(mrCenter) > 0 {
		for _, item := range mrCenter {
			res = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(res, item[1])
		}
	}
	if mrA := reA.FindAllStringSubmatch(res, -1); len(mrA) > 0 {
		for _, item := range mrA {
			res = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(res, fmt.Sprintf("[%s](%s)", item[2], item[1]))
		}
	}
	res = strings.Replace(res, "&nbsp;", " ", -1)
	res = regexp.MustCompile("</?p>").ReplaceAllLiteralString(res, "")
	return string(res)
}

//func GetBotCommandEntity(m *tgbotapi.Message) *tgbotapi.MessageEntity {
//	for _, entity := range *m.Entities {
//		if entity.Type == "bot_command" {
//			return &entity
//		}
//	}
//	return nil
//}

func isNewLevel(oldLevel *LevelInfo, newLevel *LevelInfo) bool {
	if oldLevel == nil {
		return true
	}
	return oldLevel.LevelId != newLevel.LevelId
}

func PrettyTimePrint(d time.Duration, nominative bool) (res *bytes.Buffer) {
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
			if nominative {
				res.WriteString(fmt.Sprintf("%d минута ", (d/60)%60))
			} else {
				res.WriteString(fmt.Sprintf("%d минуту ", (d/60)%60))
			}
		case 2, 3, 4, 22, 23, 24, 32, 33, 34, 42, 43, 44, 52, 53, 54:
			res.WriteString(fmt.Sprintf("%d минуты ", (d/60)%60))
		default:
			res.WriteString(fmt.Sprintf("%d минут ", (d/60)%60))
		}

	}
	if d%60 > 0 {
		switch d % 60 {
		case 1, 21, 31, 41, 51:
			if nominative {
				res.WriteString(fmt.Sprintf("%d секунда", d%60))
			} else {
				res.WriteString(fmt.Sprintf("%d секунду", d%60))
			}
		case 2, 3, 4, 22, 23, 24, 32, 33, 34, 42, 43, 44, 52, 53, 54:
			res.WriteString(fmt.Sprintf("%d секунды", d%60))
		default:
			res.WriteString(fmt.Sprintf("%d секунд", d%60))
		}
	}
	return
}
