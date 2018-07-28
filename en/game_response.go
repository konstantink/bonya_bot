package en

// GameResponse object represents the whole response from the engine
type GameResponse struct {
	Level         *Level
	Levels        *LevelsList
	EngineActions interface{} `json:"-"`
	Event         int8        `json:"-"`

	GameID        int  `json:"GameId"`
	GameTypeID    int8 `json:"GameTypeId"`
	GameZoneID    int8 `json:"GameZoneId"`
	GameNumber    int
	GameTitle     string
	LevelSequence Sequence
	UserID        int `json:"UserId"`
	TeamID        int `json:"TeamId"`
}
