package main

type BotCommand int8

const (
	InfoCommand BotCommand = 1 << iota
	SetChatIdCommand
	WatchCommand
	StopWatchingCommand
	TestHelpChange
)

const (
	//CoordinateLink = `<a href="https://maps.google.com/maps?daddr=%v&saddr=My+Location">%v</a>`
	//CoordinateLink = `<a href="comgooglemapurl://maps.google.com/maps?daddr=%v&saddr=My+Location">%v</a>`
	//CoordinateLink = `comgooglemapsurl://maps.google.com/maps?daddr=%v&saddr=My+Location`
	CoordinateLink = `[%v](http://maps.google.com/maps?daddr=%v&saddr=My+Location)`
	//CoordinateLink = `[%v](comgooglemapsurl://maps.google.com/maps?daddr=%v&saddr=My+Location)`
)

const (
	LevelInfoString = `
*Номер уровня:* %d
*Название уровня:* %q
*Времени на уровень:* %s
*Автопереход через:* %s
*Задание:*
%s
`
	HelpInfoString = `
*Подсказка:* %d
*Текст:* %s`
	//MixedActionInfoString = `
//*%s* вбил код *%q*.`
	CorrectAnswerString = `*+* %q *%s*`
	IncorrectAnswerString = `*-* %q *%s*`
	SectorInfoString = `
	Сектор *%q* закрыт. Осталось %d из %d`
)

var (
	BotCommandDict map[string]BotCommand = map[string]BotCommand{
	"info": InfoCommand,
	"setchat": SetChatIdCommand,
	"watch": WatchCommand,
	"stopwatching": StopWatchingCommand,
	"helpchange": TestHelpChange}
)

