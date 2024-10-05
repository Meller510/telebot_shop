package main

import (
	"bot/pkg/config"
	"bot/pkg/repository/postgres"
	"bot/pkg/telegram"
	"log"
	"time"

	tele "gopkg.in/telebot.v3"
)

func main() {

	cfg, err := config.InitConfig()
	if err != nil {
		log.Fatal(err)
	}

	rep := postgres.NewRepository(cfg.DataBase)

	pref := tele.Settings{
		Synchronous: true,
		Token:       cfg.TelegramToken,
		Poller:      &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b := telegram.NewBot(pref, rep, cfg)
	b.Start()

}
