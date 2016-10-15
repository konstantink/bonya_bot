package main

import (
	"gopkg.in/telegram-bot-api.v4"
	"time"
	"fmt"
	"bytes"
)

func GetBotCommandEntity(m *tgbotapi.Message) *tgbotapi.MessageEntity {
	for _, entity := range(*m.Entities) {
		if (entity.Type == "bot_command") {
			return &entity
		}
	}
	return nil
}

func IsMessageBotCommand(m *tgbotapi.Message) bool {
	for _, entity := range(*m.Entities)  {
		if (entity.Type == "bot_command") {
			return true
		}
	}
	return false
}

func PrettyTimePrint(d time.Duration) (res *bytes.Buffer) {
	//var correctTime = d*time.Second
	var s string
	res = bytes.NewBufferString(s)
	//defer res.Close()
	if (d/3600) > 0 {
		//res.WriteString(fmt.Sprintf("%d часов ", d/3600))
		switch (d/3600) {
		case 1, 21, 31, 41, 51:
			res.WriteString(fmt.Sprintf("%d час ", d/3600))
		case 2, 3, 4, 22, 23, 24, 32, 33, 34, 42, 43, 44, 52, 53, 54:
			res.WriteString(fmt.Sprintf("%d часа ", d/3600))
		default:
			res.WriteString(fmt.Sprintf("%d часов ", d/3600))
		}
	}
	if (d/60)%60 > 0 {
		switch (d/60)%60 {
		case 1, 21, 31, 41, 51:
			res.WriteString(fmt.Sprintf("%d минуту ", (d/60)%60))
		case 2, 3, 4, 22, 23, 24, 32, 33, 34, 42, 43, 44, 52, 53, 54:
			res.WriteString(fmt.Sprintf("%d минуты ", (d/60)%60))
		default:
			res.WriteString(fmt.Sprintf("%d минут ", (d/60)%60))
		}

	}
	if d%60 > 0 {
		switch d%60 {
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