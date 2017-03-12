package main

import "time"

type HelpState int8

const (
	Closed HelpState = iota
	Opened HelpState = 1 << iota
)

type HelpInfo struct {
	HelpId           int32
	Number           int8
	HelpText         string
	IsPenalty        bool
	Penalty          int16
	PenaltyComment   string
	RequestConfirm   bool
	PenaltyHelpState HelpState
	RemainSeconds    int32
	PenaltyMessage   string
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

//
// Sector related types
//
type SectorInfo struct {
	SectorId   int32
	Order      int8
	Name       string
	Answer     map[string]interface{}
	IsAnswered bool
}

type ExtendedSectorInfo struct {
	sectorInfo    *SectorInfo
	sectorsPassed int8
	sectorsLeft   int8
	totalSectors  int8
}

type LevelSectors []SectorInfo

//
// Bonus related types
//
type BonusInfo struct {
	BonusId        int32
	Name           string
	Number         int8
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
	AttemptsNumber       int8
	AttemptsPeriod       time.Duration
	RequiredSectorsCount int8
	PassedSectorsCount   int8
	SectorsLeftToClose   int8
	Tasks                LevelTasks
	MixedActions         LevelMixedActions
	Helps                LevelHelps
	PenaltyHelps         LevelPenaltyHelps
	Bonuses              LevelBonuses
	//Messages             []string
	Sectors LevelSectors
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
