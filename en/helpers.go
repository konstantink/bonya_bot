package en

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang-collections/collections/stack"
	"golang.org/x/net/html"
)

// Tag string type that corresponds to the html tags
type Tag struct {
	Tag   string
	Attrs map[string]string
}

const (
	iTag      string = "i"
	bTag      string = "b"
	strongTag string = "strong"
	scriptTag string = "script"
	aTag      string = "a"
)

type Coordinate struct {
	Lat            float64 `json:"lattitude"`
	Lon            float64 `json:"longtitude"`
	OriginalString string  `json:"name"`
}
type Coordinates []Coordinate

func (c Coordinate) String() (text string) {
	text = fmt.Sprintf("%s (%f, %f)", c.OriginalString, c.Lat, c.Lon)
	return
}

// Image stores data for the images in the text, e.g. URL to download the image,
// Filepath - path where file was downloaded
type Image struct {
	URL      string
	Caption  string
	Filepath string
}

// Images - array of Image objects
type Images []Image

func extractCoordinates(text string, re *regexp.Regexp) (string, Coordinates) {
	var (
		result = text
		mr     = re.FindAllStringSubmatch(text, -1)
		coords = Coordinates{}
	)
	if len(mr) > 0 {
		for _, item := range mr {
			lon, _ := strconv.ParseFloat(item[1], 64)
			lat, _ := strconv.ParseFloat(item[2], 64)
			if len(item) > 3 {
				coords = append(coords, Coordinate{Lat: lon, Lon: lat, OriginalString: item[3]})
			} else {
				coords = append(coords, Coordinate{Lat: lon, Lon: lat, OriginalString: item[0]})
			}
			result = regexp.MustCompile(item[0]).ReplaceAllLiteralString(result, "#coords#")
		}
	}

	return result, coords
}

// ExtractCoordinates extracts coordinates from the given text and returns the updated string
// with replaced coordinates and the list of coordinates
func ExtractCoordinates(text string) (string, Coordinates) {
	var (
		// <a href="geo:49.976136, 36.267256">49.976136, 36.267256</a>
		geoHrefRe = regexp.MustCompile("<a.+?href=\"geo:(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,})\">(.+?)</a>")
		// <a href="https://www.google.com.ua/maps/@50.0363257,36.2120039,19z" target="blank">50.036435 36.211914</a>
		hrefRe = regexp.MustCompile("<a.+?href=\"https?://.+?(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,}).*?\">(.+?)</a>")
		// 49.976136, 36.267256
		numbersRe = regexp.MustCompile("(\\d{2}[.,]\\d{3,}),?\\s*(\\d{2}[.,]\\d{3,})")

		res       = text
		coords    = Coordinates{}
		tmpCoords Coordinates
	)

	log.Print("[INFO] Extract coordinates from task text")
	for _, re := range []*regexp.Regexp{geoHrefRe, hrefRe, numbersRe} {
		res, tmpCoords = extractCoordinates(res, re)
		coords = append(coords, tmpCoords...)
	}

	for _, coord := range coords {
		res = strings.Replace(res, "#coords#", coord.OriginalString, 1)
	}
	if DEBUG {
		log.Printf("[DEBUG] Found %d coordinates", len(coords))
	}

	return res, coords
}

func extractImages(text string, re *regexp.Regexp, caption string, start int) (string, Images) {
	var (
		result = text
		mr     = re.FindAllStringSubmatch(text, -1)
		images = Images{}
	)
	if len(mr) > 0 {
		for i, item := range mr {
			images = append(images, Image{URL: item[1], Caption: fmt.Sprintf("%s #%d", caption, start+i)})
			result = regexp.MustCompile(regexp.QuoteMeta(item[0])).
				ReplaceAllLiteralString(result, fmt.Sprintf("%s #%d", caption, start+i))
		}
	}
	return result, images
}

// ExtractImages extracts images from the given text and returns the updated
// version of the text and the list of images
func ExtractImages(text string, caption string) (string, Images) {
	var (
		reImg     = regexp.MustCompile("<img.+?src=\"\\s*(https?://.+?)\\s*\".*?>")
		reA       = regexp.MustCompile("<a.+?href=\\\\?\"(https?://.+?\\.(jpg|png|bmp))\\\\?\".*?>(.*?)</a>")
		result    = text
		images    = Images{}
		tmpImages Images
	)
	//log.Printf("Before image replacing: %s", text)
	log.Print("[INFO] Extract images from task text")
	for _, re := range []*regexp.Regexp{reImg, reA} {
		result, tmpImages = extractImages(result, re, caption, len(images)+1)
		images = append(images, tmpImages...)
	}
	if DEBUG {
		log.Printf("[DEBUG] Found %d images", len(images))
	}
	return result, images
}

// ReplaceHTMLTags finds all html tags and removes them. Some tags like bold, italic are replaed with
// makrkups for telegram
func ReplaceHTMLTags(text string) string {
	var (
		parser    = html.NewTokenizer(strings.NewReader(text))
		tagStack  = stack.New()
		textToTag = map[int]string{}
	)

	for {
		node := parser.Next()
		switch node {
		case html.ErrorToken:
			result := strings.Replace(textToTag[0], "&nbsp;", " ", -1)
			return result
		case html.TextToken:
			t := string(parser.Text())
			textToTag[tagStack.Len()] = strings.Join([]string{textToTag[tagStack.Len()], t}, "")
		case html.StartTagToken:
			tagName, hasAttr := parser.TagName()
			if string(tagName) == scriptTag {
				// We can skip script tags, as they are invisible for the user, but we can indicate that there are
				// scripts in the task. To skip tag, it is necessary to call Next() two times:
				// 1) returns TextToken with the script body
				// 2) returns EndTagToken for the closed script tag
				// Usually script tag doesn't have any neste tags, so this aproach should work
				log.Printf("[INFO] Skipping script tag")
				parser.Next()
				parser.Next()
				continue
			}
			tag := Tag{Tag: string(tagName), Attrs: map[string]string{}}
			if hasAttr {
				for {
					attr, val, moreAttr := parser.TagAttr()
					if DEBUG {
						log.Printf("[DEBUG] Found attr %s", attr)
					}
					tag.Attrs[string(attr)] = string(val)
					if !moreAttr {
						break
					}
				}
			}
			if DEBUG {
				log.Printf("[DEBUG] Found tag %q", tag)
			}
			tagStack.Push(tag)
		case html.EndTagToken:
			var (
				addText      string
				tagNo        = tagStack.Len()
				tag          = tagStack.Pop()
				closedTag, _ = parser.TagName()
			)
			if tag.(Tag).Tag != string(closedTag) {
				log.Printf("[WARNING] Found closed tag %q but expected %q", closedTag, tag)
				continue
			}
			if DEBUG {
				log.Printf("[DEBUG] Found end of tag %q", closedTag)
			}
			switch tag.(Tag).Tag {
			case iTag:
				addText = fmt.Sprintf("_%s_", textToTag[tagNo])
			case bTag, strongTag:
				addText = fmt.Sprintf("*%s*", textToTag[tagNo])
			case aTag:
				// if strings.Compare(string(attr), "href") == 0 {
				addText = fmt.Sprintf("[%s](%s)", textToTag[tagNo], tag.(Tag).Attrs["href"])
				// }
			default:
				addText = textToTag[tagNo]
			}
			textToTag[tagStack.Len()] = strings.Join([]string{textToTag[tagStack.Len()], addText}, "")
			delete(textToTag, tagNo)
		}
	}
}

// ReplaceCommonTags deprecated - should be removed!!!
func ReplaceCommonTags(text string) string {
	log.Print("Replace html tags")
	var (
		reBr     = regexp.MustCompile("<br\\s*/?>")
		reHr     = regexp.MustCompile("<hr.*?/?>")
		reP      = regexp.MustCompile("<p>([^ ]+?)</p>")
		reBold   = regexp.MustCompile("<b.*?/?>((?s:.*?))</b>")
		reStrong = regexp.MustCompile("<strong.*?>(.*?)</strong>")
		reItalic = regexp.MustCompile("<i>((?s:.+?))</i>")
		reSpan   = regexp.MustCompile("<span.*?>(.*?)</span>")
		reCenter = regexp.MustCompile("<center>((?s:.*?))</center>")
		reFont   = regexp.MustCompile("<font.+?color\\s*=\\\\?[\"«]?#?(\\w+)\\\\?[\"»]?.*?>((?s:.*?))</font>")
		reA      = regexp.MustCompile("<a.+?href=\\\\?\"(.+?)\\\\?\".*?>(.+?)</a>")
		res      = text
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

func BlockTypeToString(typeId int8) string {
	if typeId == 0 || typeId == 1 {
		return "Игрок"
	}
	return "Команда"
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
