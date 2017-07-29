#!/bin/sh

go build bot_telebot.go constants.go en_api.go en_types.go helpers.go time_checking_machine.go web_server.go
BONYA_MAIN_CHAT=-1001075414838 BONYA_USER=bonyaBot BONYA_PASSWORD=tonkpils85 BONYA_ENGINE_DOMAIN=quest.ua BONYA_GAME_ID=58438 BONYA_BOT_TOKEN=152794909:AAE5xuhVDQqJNopSOV6AQ333Em8PQ1bxQzM ./bot_telebot
