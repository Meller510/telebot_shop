package telegram

import (
	"bot/pkg/config"
	"bot/pkg/repository"
	"bot/pkg/shop"
	"database/sql"
	"fmt"
	"log"
	"time"

	"gopkg.in/gomail.v2"
	t "gopkg.in/telebot.v3"
)

type Bot struct {
	bot        t.Bot
	repository repository.Repository
	config     *config.Config
	// Logger     logging.Logger
}

func NewBot(pref t.Settings, rep repository.Repository, cfg *config.Config) Bot {
	b, err := t.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	return Bot{
		bot:        *b,
		repository: rep,
		config:     cfg,
	}
}

func (b *Bot) Start() {
	log.Printf("Start bot on account %s \n", b.bot.Me.Username)
	b.handleUpdates()
	b.bot.Start()
}

func (b *Bot) switchTypeMsg(menu shop.MenuItem, itemsMenu *t.ReplyMarkup, c t.Context) error {
	en := c.Message().Entities

	switch {
	case len(en) != 0 && en[0].Type == "bot_command":
		return c.Send(menu.Title(), itemsMenu, t.ModeHTML)

	case c.Callback().Message.Photo == nil && menu.Image():
		a := &t.Photo{File: t.FromDisk(fmt.Sprintf("%s/%s.jpg", b.config.PathImage, menu.ImageName())),
			Caption: menu.Title()}

		c.Delete()

		return c.Send(a, itemsMenu, t.ModeHTML)

	case !menu.Image() && c.Callback().Message.Photo != nil:
		c.Delete()
		return c.Send(menu.Title(), itemsMenu, t.ModeHTML)

	case c.Callback().Message.Photo != nil && menu.Image():
		a := &t.Photo{File: t.FromDisk(fmt.Sprintf("%s/%s.jpg", b.config.PathImage, menu.ImageName())),
			Caption: menu.Title()}
		return c.Edit(a, menu.Title(), itemsMenu, t.ModeHTML)

	default:
		return c.Edit(menu.Title(), itemsMenu, t.ModeHTML)
	}
}

func (b *Bot) identificationUser(c t.Context) (user *shop.User) {
	var sender *t.User
	var date int64

	if c.Callback() != nil {
		sender = c.Callback().Sender
		date = c.Callback().Message.Unixtime
	} else if c.Message() != nil {
		sender = c.Message().Sender
		date = c.Message().Unixtime
	}

	user, err := b.repository.SelectUser(sender.ID, date)

	if err != nil {
		user = shop.NewUser(sender.ID, sender.FirstName, sender.Username, date)
		b.repository.AddUser(user)
	} else {
		en := c.Message().Entities
		if len(en) != 0 && en[0].Type == "bot_command" {
			user.OrderStatus = sql.NullString{}
		}
	}

	return user
}

func (b *Bot) sendMail(user *shop.User, header, body string) {
	msg := gomail.NewMessage()

	msg.SetAddressHeader("From", b.config.Email.Address, b.config.Email.Initial)

	msg.SetHeader("To", b.config.Email.Address)

	msg.SetHeader("Subject", header)
	msg.SetBody("text/html", body)

	dialer := gomail.NewDialer(b.config.Email.Host, b.config.Email.Port,
		b.config.Email.Address, b.config.Email.Password)

	min, _ := time.ParseDuration("5m")
	t := time.Now()

	var err error
	for time.Since(t).Minutes() < min.Minutes() {
		if err = dialer.DialAndSend(msg); err == nil {
			break
		}
		time.Sleep(10 * time.Second)
	}
}
