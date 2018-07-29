package main

import (
	"encoding/json"
	"fmt"
	"github.com/tucnak/telebot"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strings"
	"time"
)

type ToChat interface {
	ToText() string
	ReplyTo() telebot.Message
}

type ExtraInfo struct {
	Coords Coordinates
	Images Images
}

type HelpState int8

const (
	Closed HelpState = iota
	Opened HelpState = 1 << iota
)

type HelpInfo struct {
	ExtraInfo `json:"-"`

	HelpId           int32
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
	help.HelpText, help.Coords = ReplaceCoordinates(help.HelpText)
	help.HelpText, help.Images = ReplaceImages(help.HelpText, "Картинка")
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
	ActionId      int32
	LevelId       int32
	LevelNumber   int8
	UserId        int32
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
	return m[i].ActionId > m[j].ActionId
}

func (m LevelMixedActions) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m MixedActionInfo) ToText() (result string) {
	log.Println("New MixedAction is added")
	if m.IsCorrect {
		result = fmt.Sprintf(CorrectAnswerString, m.Answer, m.Login)
	} else {
		result = fmt.Sprintf(IncorrectAnswerString, m.Answer, m.Login)
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

func newSectorStatistics(levelInfo *LevelInfo) sectorStatistics {
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

func NewExtendedLevelSectors(levelInfo *LevelInfo) *ExtendedLevelSectors {
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
	ExtraInfo `json:"-"`

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
	bi.Help, bi.Coords = ReplaceCoordinates(bi.Help)
	bi.Help, bi.Images = ReplaceImages(bi.Help, "Бонус")
	bi.Help = ReplaceCommonTags(bi.Help)
}

func (bi *BonusInfo) ToText() (result string) {
	result = fmt.Sprintf(BonusInfoString, bi.Name, bi.Answer["Answer"], bi.Help)
	return
}

func (li *BonusInfo) ReplyTo() (message telebot.Message) {
	return
}

//
// Level info related types
//
type Sequence int8

const (
	Linear Sequence = iota
	Said
	Random
	Assault
	DynamicRandom
)

type LevelInfo struct {
	ExtraInfo `json:"-"`

	LevelId              int32
	GameId               int32
	GameTypeId           int8
	GameZoneId           int8
	GameNumber           int32
	GameTitle            string
	LevelSequence        Sequence
	UserId               int32
	TeamId               int32
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
	BlockTargetId        int8
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
	//Messages             []string
	Sectors LevelSectors
}

func NewLevelInfo(response *http.Response) *LevelInfo {
	var lvlResponse = &LevelResponse{}

	if response == nil {
		return &LevelInfo{}
	}

	body, _ := ioutil.ReadAll(response.Body)

	err := json.Unmarshal(body, lvlResponse)
	if err != nil {
		log.Println("ERROR: failed to parse level json:", err)
		return &LevelInfo{}
	}

	return lvlResponse.Level
}

func (li *LevelInfo) ProcessText() {
	li.Tasks[0].TaskText, li.Coords = ReplaceCoordinates(li.Tasks[0].TaskText)
	li.Tasks[0].TaskText, li.Images = ReplaceImages(li.Tasks[0].TaskText, "Картинка")
	li.Tasks[0].TaskText = ReplaceCommonTags(li.Tasks[0].TaskText)
}

func (li *LevelInfo) ToText() (result string) {
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
		block = fmt.Sprintf("Есть"+LevelBlockInfoString, BlockTypeToString(li.BlockTargetId),
			li.AttemtsNumber, PrettyTimePrint(li.AttemtsPeriod, true))
	} else {
		block = "Нет"
	}

	result = fmt.Sprintf(LevelInfoString,
		li.Number,
		li.Name,
		PrettyTimePrint(li.Timeout, true),
		PrettyTimePrint(li.TimeoutSecondsRemain, false),
		PrettyTimePrint(time.Duration(math.Abs(float64(li.TimeoutAward))), true),
		li.RequiredSectorsCount,
		block,
		li.Tasks[0].TaskText)
	return
}

func (li *LevelInfo) ReplyTo() (message telebot.Message) {
	return
}

type ShortLevelInfo struct {
	LevelId     int32
	LevelNumber int8
	LevelName   string
	Dismissed   bool
	IsPassed    bool
	Task        string
	LevelAction string
}

type LevelsList []ShortLevelInfo

func (l *LevelsList) Len() int {
	return len(*l)
}

type LevelResponse struct {
	Level  *LevelInfo
	Levels *LevelsList
}

type Codes struct {
	replyTo                     telebot.Message
	correct, incorrect, notSent []string
}

func (codes *Codes) ToText() (result string) {
	if len(codes.correct) > 0 {
		result = fmt.Sprintf(CorrectAnswerString, strings.Join(codes.correct, ", "))
	}
	if len(codes.incorrect) > 0 {
		result += fmt.Sprintf(IncorrectAnswerString, strings.Join(codes.incorrect, ", "))
	}
	if len(codes.notSent) > 0 {
		result += fmt.Sprintf(NotSentAnswersString, strings.Join(codes.notSent, ", "))
	}
	return
}

func (codes *Codes) ReplyTo() telebot.Message {
	return codes.replyTo
}

//
// Code related types
//
type codeRequest struct {
	LevelId     int32 `json:"LevelId"`
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
