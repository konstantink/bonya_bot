package main

type BotCommand int8

const (
	InfoCommand BotCommand = 1 + iota
	SetChatIdCommand
	WatchCommand
	StopWatchingCommand
	CodeCommand
	CompositeCodeCommand
	SectorsLeftCommand
	TimeLeftCommand
	ListHelpsCommand
	HelpTimeCommand
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
*Штраф за автопереход:* %s
*Секторов закрыть:* %d
*Ограничение:* %s
*Задание:*
%s`
	LevelBlockInfoString = `
=============================
*Тип ограничения:* %s
*Количество попыток:* %d
*Время ограничения:* %s
=============================`

	HelpInfoString = `
*Подсказка:* %d
*Текст:* %s`
	//MixedActionInfoString = `
	//*%s* вбил код *%q*.`
	//CorrectAnswerString   = `*+* %q *%s*`
	CorrectAnswerString = "*+* %s\n"
	//IncorrectAnswerString = `*-* %q *%s*`
	IncorrectAnswerString = "*-* %s\n"
	NotSentAnswersString  = "*блок:* %s"
	SectorClosedString    = "Сектор *%q* закрыт. Осталось %d из %d"
	SectorInfoString      = `
Осталось *%d* из *%d*
Оставшиеся сектора:
%s`
	BonusInfoString = `
Бонус *%q* открыт
Закрыт кодом: %s
Текст: %s`
	TimeLeftString = `
Осталось %s
*Pink Panther, вперед!*
`
	HelpTimeLeft = `
*Подсказка %d* будет через %s`
)

var (
	BotCommandDict map[string]BotCommand = map[string]BotCommand{
		"info":         InfoCommand,
		"setchat":      SetChatIdCommand,
		"watch":        WatchCommand,
		"stopwatching": StopWatchingCommand,
		"c":            CodeCommand,
		"с":            CodeCommand,
		"cc":           CompositeCodeCommand,
		"сс":           CompositeCodeCommand,
		"sl":           SectorsLeftCommand,
		"ос":           SectorsLeftCommand,
		"tl":           TimeLeftCommand,
		"ов":           TimeLeftCommand,
		"lh":           ListHelpsCommand,
		"ht":           HelpTimeCommand,
		"helpchange":   TestHelpChange}
)
