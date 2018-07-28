package en

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/tucnak/telebot"
)

type ToChat interface {
	ToText() string
	ReplyTo() telebot.Message
}

// Extra - extra information about coordinates or images that are available in level information,
// hint, bonus, etc.
type Extra struct {
	Coords        Coordinates
	Images        Images
	ProcessedText string
}

// HelpState type to define the state of the hint
type HelpState int8

const (
	// Closed means that help is unavailable
	Closed HelpState = iota
	// Opened means that help is available
	Opened HelpState = 1 << iota
)

type HelpInfo struct {
	Extra `json:"-"`

	HelpID           int
	Number           int8
	HelpText         string
	IsPenalty        bool
	Penalty          int16
	PenaltyComment   string
	RequestConfirm   bool
	PenaltyHelpState HelpState
	RemainSeconds    time.Duration
	PenaltyMessage   string
}

func (help *HelpInfo) ProcessText() {
	//log.Printf("Before %s", help.HelpText)
	help.HelpText, help.Coords = ExtractCoordinates(help.HelpText)
	help.HelpText, help.Images = ExtractImages(help.HelpText, "Картинка")
	help.HelpText = ReplaceCommonTags(help.HelpText)
	//log.Printf("After %s", help.HelpText)
}

func (help *HelpInfo) ToText() (result string) {
	//result, images := ReplaceImages(help.HelpText, "Картинка")
	result = fmt.Sprintf(HelpInfoString, help.Number, ReplaceCommonTags(help.HelpText))
	return
}

func (help *HelpInfo) ReplyTo() (message telebot.Message) {
	return
}

type LevelHelps []HelpInfo
type LevelPenaltyHelps []HelpInfo

//
// Task related types
//
type TaskInfo struct {
	ReplaceNlToBr     bool
	TaskText          string
	TaskTextFormatted string
}

type LevelTasks []TaskInfo

//
// Mixed Action related types
//
type MixedActionKind int8

const (
	LevelAnswer MixedActionKind = iota
	BonusAnswer
)

type MixedActionInfo struct {
	ActionID      int
	LevelID       int
	LevelNumber   int8
	UserID        int
	Kind          MixedActionKind
	Login         string
	Answer        string
	AnswForm      string
	EnterDateTime map[string]float64 `json:"-"`
	LocDateTime   string
	IsCorrect     bool
	Award         int8 `json:"-"`
	LocAward      int8 `json:"-"`
	Penalty       int8 `json:"-"`
}

type LevelMixedActions []MixedActionInfo

func (m LevelMixedActions) Len() int {
	return len(m)
}

func (m LevelMixedActions) Less(i, j int) bool {
	return m[i].ActionID > m[j].ActionID
}

func (m LevelMixedActions) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m MixedActionInfo) ToText() (result string) {
	log.Println("New MixedAction is added")
	if m.IsCorrect {
		result = fmt.Sprintf(CorrectAnswerString, m.Answer)
	} else {
		result = fmt.Sprintf(IncorrectAnswerString, m.Answer)
	}
	return
}

func (m MixedActionInfo) ReplyTo() (message telebot.Message) {
	return
}

//
// Sector related types
//
type SectorInfo struct {
	SectorId   int32
	Order      int16
	Name       string
	Answer     map[string]interface{}
	IsAnswered bool
}

type sectorStatistics struct {
	sectorsPassed int16
	sectorsLeft   int16
	totalSectors  int16
}

func newSectorStatistics(levelInfo *Level) sectorStatistics {
	return sectorStatistics{
		sectorsPassed: levelInfo.PassedSectorsCount,
		sectorsLeft:   levelInfo.SectorsLeftToClose,
		totalSectors:  int16(len(levelInfo.Sectors)),
	}
}

type ExtendedSectorInfo struct {
	sectorStatistics
	sectorInfo *SectorInfo
}

func (esi *ExtendedSectorInfo) ToText() (result string) {
	result = fmt.Sprintf(SectorClosedString, esi.sectorInfo.Name, esi.sectorsLeft, esi.totalSectors)
	return
}

func (esi *ExtendedSectorInfo) ReplyTo() (message telebot.Message) {
	return
}

type LevelSectors []SectorInfo

type ExtendedLevelSectors struct {
	sectorStatistics
	levelSectors LevelSectors
}

func NewExtendedLevelSectors(levelInfo *Level) *ExtendedLevelSectors {
	sectorStatistic := newSectorStatistics(levelInfo)
	return &ExtendedLevelSectors{
		sectorStatistics: sectorStatistic,
		levelSectors:     levelInfo.Sectors,
	}
}

func (ls *ExtendedLevelSectors) ToText() (result string) {
	var openSectorNames []string
	for _, sector := range ls.levelSectors {
		if !sector.IsAnswered {
			openSectorNames = append(openSectorNames, sector.Name)
		}
	}
	result = fmt.Sprintf(SectorInfoString, ls.sectorsLeft, ls.totalSectors, strings.Join(openSectorNames, "\n"))
	return
}

func (ls *ExtendedLevelSectors) ReplyTo() (message telebot.Message) {
	return
}

//
// Bonus related types
//
type BonusInfo struct {
	Extra `json:"-"`

	BonusId        int32
	Name           string
	Number         int16
	Task           string
	Help           string
	IsAnswered     bool
	Expired        bool
	SecondsToStart time.Duration
	SecondsLeft    time.Duration
	AwardTime      time.Duration
	Answer         map[string]interface{}
}

type LevelBonuses []BonusInfo

func (bi *BonusInfo) ProcessText() {
	bi.Help, bi.Coords = ExtractCoordinates(bi.Help)
	bi.Help, bi.Images = ExtractImages(bi.Help, "Бонус")
	bi.Help = ReplaceCommonTags(bi.Help)
}

func (bi *BonusInfo) ToText() (result string) {
	result = fmt.Sprintf(BonusInfoString, bi.Name, bi.Help)
	return
}

func (li *BonusInfo) ReplyTo() (message telebot.Message) {
	return
}

//
// Level info related types
//
type Codes struct {
	Message                     telebot.Message
	Correct, Incorrect, NotSent []string
}

func (codes *Codes) ToText() (result string) {
	if len(codes.Correct) > 0 {
		result = fmt.Sprintf(CorrectAnswerString, strings.Join(codes.Correct, ", "))
	}
	if len(codes.Incorrect) > 0 {
		result += fmt.Sprintf(IncorrectAnswerString, strings.Join(codes.Incorrect, ", "))
	}
	if len(codes.NotSent) > 0 {
		result += fmt.Sprintf(NotSentAnswersString, strings.Join(codes.NotSent, ", "))
	}
	return
}

func (codes *Codes) ReplyTo() telebot.Message {
	return codes.Message
}

//
// Code related types
//
type codeRequest struct {
	LevelID     int32 `json:"LevelId"`
	LevelNumber int8  `json:"LevelNumber"`
}

type SendCodeRequest struct {
	codeRequest
	LevelAction string `json:"LevelAction.Answer"`
}

type SendBonusCodeRequest struct {
	codeRequest
	LevelAction string `json:"BonusAction.Answer"`
}

// // GameLevel structure that represents level of the game, it contains level information and
// // extracted additional information such as coordinates, images, etc.
// type GameLevel struct {
// 	Level  *LevelInfo
// 	Coords Coordinates
// 	Images Images
// }

// func (gl *GameLevel) GetLevelTask() string {
// 	return gl.Level.GetLevelTask()
// }
