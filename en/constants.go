package en

const (
	//CoordinateLink = `<a href="https://maps.google.com/maps?daddr=%v&saddr=My+Location">%v</a>`
	//CoordinateLink = `<a href="comgooglemapurl://maps.google.com/maps?daddr=%v&saddr=My+Location">%v</a>`
	//CoordinateLink = `comgooglemapsurl://maps.google.com/maps?daddr=%v&saddr=My+Location`
	CoordinateLink = `[%v](http://maps.google.com/maps?daddr=%v&saddr=My+Location)`
	//CoordinateLink = `[%v](comgooglemapsurl://maps.google.com/maps?daddr=%v&saddr=My+Location)`
)

const (
	// LevelInfoString general information about level
	LevelInfoString = `
%s #Ап
*Номер уровня:* %d из %d
*Название уровня:* %q
*Времени на уровень:* %s
*Автопереход через:* %s
*Штраф за автопереход:* %s
*Секторов закрыть:* %d
*Ограничение:* %s`

	// LevelTaskString level task
	LevelTaskString = `
*Задание:*
%s`

	// LevelBlockInfoString information about blocking for incorrect code
	LevelBlockInfoString = `
%s
*Тип ограничения:* %s
*Количество попыток:* %d
*Время ограничения:* %s
%s`

	// HelpInfoString hint information
	HelpInfoString = `
*Подсказка:* %d
*Текст:* %s`
	//MixedActionInfoString = `
	//*%s* вбил код *%q*.`

	//CorrectAnswerString   = `*+* %q *%s*`
	CorrectAnswerString = "*+* %s\n"
	//IncorrectAnswerString = `*-* %q *%s*`
	IncorrectAnswerString = "*-* %s\n"

	// NotSentAnswersString codes that were not sent because of block
	NotSentAnswersString = "*блок:* %s"

	// SectorClosedString message for closed sector
	SectorClosedString = "Сектор *%q* закрыт. Осталось %d из %d"

	// SectorInfoString information about how many sectors left to close
	SectorInfoString = `
Осталось *%d* из *%d*
Оставшиеся сектора:
%s`

	// BonusInfoString information for bonus
	BonusInfoString = `
Бонус *%q* открыт
Текст: %s`

	// TimeLeftString string with time left for the level
	TimeLeftString = `
Осталось %s
*ГО, КиПеш, ГО!!!*
`
	// HelpTimeLeft string with time left before hint
	HelpTimeLeft = `
*Подсказка %d* будет через %s`
)
