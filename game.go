	package main

	import "fmt"

	type GameSettings struct {
		ChatId int64
		Domain string
		GameId int32
		UserName string
		Password string
	}

	func (gs GameSettings) String() string {
		return fmt.Sprintf("<GameSettings: %d %s %s %s>", gs.ChatId, gs.Domain, gs.GameId, gs.UserName)
	}

	//func (gs *GameSettings)