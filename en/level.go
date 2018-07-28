package en

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/tucnak/telebot"
)

// Sequence represents the type of the game
type Sequence int8

const (
	// Linear corresponds to linear type of the game when players get level task after they
	// finish previous one
	Linear Sequence = iota
	// Said some type of game
	Said
	// Random corresponds to the type of game when players receive all tasks at a time and
	// then they decide alone in what sequence will they pass all levels
	Random
	// Assault corresponds to the type of game when players need to do some strange things
	Assault
	// DynamicRandom something similar to Random
	DynamicRandom
)

// Level represents the whole level with all settings
type Level struct {
	Extra `json:"-"`

	Parent *GameResponse `json:"-"`

	LevelID              int32 `json:"LevelId"`
	Name                 string
	Number               int8
	Timeout              time.Duration
	TimeoutSecondsRemain time.Duration
	TimeoutAward         time.Duration
	IsPassed             bool
	Dismissed            bool
	StartTime            map[string]float64 `json:"-"`
	HasAnswerBlockRule   bool
	BlockDuration        time.Duration
	BlockTargetID        int8 `json:"BlockTargetId"`
	AttemtsNumber        int8
	AttemtsPeriod        time.Duration
	RequiredSectorsCount int16
	PassedSectorsCount   int16
	SectorsLeftToClose   int16
	Tasks                LevelTasks
	MixedActions         LevelMixedActions
	Helps                LevelHelps
	PenaltyHelps         LevelPenaltyHelps
	Bonuses              LevelBonuses
	Sectors              LevelSectors
	//Messages             []string
}

// NewLevel constructor for the Level structure, builds Level object out from the response
// to the request of getting level information from the engine
func NewLevel(response *http.Response) *Level {
	// var lvlResponse = &LevelResponse{}
	var gameResponse = &GameResponse{}

	if response == nil {
		return &Level{}
	}

	body, _ := ioutil.ReadAll(response.Body)

	err := json.Unmarshal(body, gameResponse)
	if err != nil {
		log.Println("ERROR: failed to parse level json:", err)
		return &Level{}
	}
	gameResponse.Level.Parent = gameResponse

	return gameResponse.Level
}

// ProcessText process the initial task text that is received from the server:
// - extracts some useful information like coordinates where to go, or images
// - removes all html tags and leaves just the text
func (li *Level) ProcessText() {
	li.Tasks[0].TaskText, li.Coords = ExtractCoordinates(li.GetLevelTask())
	li.Tasks[0].TaskText, li.Images = ExtractImages(li.GetLevelTask(), "Картинка")
	// li.Tasks[0].TaskText = ReplaceCommonTags(li.Tasks[0].TaskText)
}

func (li *Level) getTask() string {
	return li.Tasks[0].TaskText
}

// GetLevelDetails returns basic information regarding level, such as time till the end,
// number of sectors, are there any blocks, etc.
func (li *Level) GetLevelDetails() (result string) {
	const emoji = "\xF0\x9F\x86\x99"
	var block string
	if li.HasAnswerBlockRule {
		blockLine := strings.Repeat("\xE2\x9D\x97", 10)
		block = fmt.Sprintf("Есть"+LevelBlockInfoString, blockLine, BlockTypeToString(li.BlockTargetID),
			li.AttemtsNumber, PrettyTimePrint(li.AttemtsPeriod, true), blockLine)
	} else {
		block = "Нет"
	}
	result = fmt.Sprintf(LevelInfoString,
		emoji,
		li.Number,
		len(*li.Parent.Levels),
		li.Name,
		PrettyTimePrint(li.Timeout, true),
		PrettyTimePrint(li.TimeoutSecondsRemain, false),
		PrettyTimePrint(time.Duration(math.Abs(float64(li.TimeoutAward))), true),
		li.RequiredSectorsCount,
		block)
	return
}

// GetLevelTask returns the formatted version of current task
func (li *Level) GetLevelTask() string {
	if li.ProcessedText != "" {
		if DEBUG {
			log.Printf("[DEBUG] Get level task from cache")
		}
		return li.ProcessedText
	}
	result := li.getTask()
	result, li.Coords = ExtractCoordinates(result)
	result, li.Images = ExtractImages(result, "Картинка")
	result = ReplaceHTMLTags(result)
	// log.Printf("[INFO] Parsed text %s", result)
	li.ProcessedText = result
	// TODO: uncomment
	// li.DownloadImages()
	return result
}

// DownloadImages downloads all images that were found in text in order to send them to chat
func (li *Level) DownloadImages() {
	log.Printf("[INFO] Downloading images %d", len(li.Images))
	downloadPath := path.Join("/tmp", "quest", strconv.Itoa(li.Parent.GameID), strconv.Itoa(int(li.Number)))
	if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
		log.Printf("[INFO] Creating directory for level %d", li.Number)
		os.MkdirAll(downloadPath, 0755)
	}
	for idx, image := range li.Images {
		fileName := path.Join(downloadPath, path.Base(image.URL))
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			log.Printf("[INFO] Download file %s to temporary directory %s", fileName, downloadPath)
			response, err := http.Get(image.URL)
			if err != nil {
				log.Println("[ERROR] Can't download image:", err)
				continue
			}
			file, err := os.Create(fileName)
			if err != nil {
				log.Fatal("[ERROR] Cannot create file:", err)
				continue
			}
			// Use io.Copy to just dump the response body to the file. This supports huge files
			defer response.Body.Close()
			defer file.Close()
			_, err = io.Copy(file, response.Body)
			if err != nil {
				log.Printf("[ERORR] %s", err)
			}
		}
		li.Images[idx].Filepath = fileName
	}
}

// ToText - deprecated
func (li *Level) ToText() (result string) {
	var (
		block string
		//coords Coordinates
	)
	//task, _ = ReplaceCoordinates(li.Tasks[0].TaskText)
	//log.Printf("After coordinates: %s", task)
	//task = ReplaceImages(task, "Картинка")
	//log.Printf("After images: %s", task)
	//task = ReplaceCommonTags(task)
	//log.Printf("After tags: %s", task)

	if li.HasAnswerBlockRule {
		block = fmt.Sprintf("Есть"+LevelBlockInfoString, "=", BlockTypeToString(li.BlockTargetID),
			li.AttemtsNumber, PrettyTimePrint(li.AttemtsPeriod, true), "=")
	} else {
		block = "Нет"
	}

	result = fmt.Sprintf(LevelInfoString,
		"=",
		li.Number, 10,
		li.Name,
		PrettyTimePrint(li.Timeout, true),
		PrettyTimePrint(li.TimeoutSecondsRemain, false),
		PrettyTimePrint(time.Duration(math.Abs(float64(li.TimeoutAward))), true),
		li.RequiredSectorsCount,
		block,
	) //li.Tasks[0].TaskText)
	return
}

// ReplyTo - deprecated
func (li *Level) ReplyTo() (message telebot.Message) {
	return
}

// ShortLevelInfo - maybe will be deprecated, as it is not used
type ShortLevelInfo struct {
	LevelID     int32
	LevelNumber int8
	LevelName   string
	Dismissed   bool
	IsPassed    bool
	Task        string
	LevelAction string
}

// LevelsList list of short level objects
type LevelsList []ShortLevelInfo

// Len returns the number of levels in current game
func (l *LevelsList) Len() int {
	return len(*l)
}

// LevelResponse represents the structure of regular response from the server
type LevelResponse struct {
	Level  *Level
	Levels *LevelsList
}
