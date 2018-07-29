#!/bin/sh

go build bot_telebot.go constants.go en_api.go en_types.go helpers.go time_checking_machine.go web_server.go
BONYA_MAIN_CHAT=-1001063770731 BONYA_USER=Germiona BONYA_PASSWORD=tonkpils85 BONYA_ENGINE_DOMAIN=demo.en.cx BONYA_GAME_ID=25733 BONYA_BOT_TOKEN=152794909:AAE5xuhVDQqJNopSOV6AQ333Em8PQ1bxQzM ./bot_telebot
