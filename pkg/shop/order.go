package shop

import (
	"bot/pkg/key"
	"bot/pkg/telegram/label"

	t "gopkg.in/telebot.v3"
)

type Order struct {
	ID       int
	UserID   int
	FullName string
	PhoneNum string
	Pay      string
	Delivery string
	Address  string
	Comment  string
	Date     string
	Basket   []BasketItem
}

func OrderNavigation() (row t.Row) {
	menu := &t.ReplyMarkup{}

	row = append(row, menu.Data(label.Basket, key.OrderExit, key.Basket),
		menu.Data(label.IMainMenu, key.OrderExit, key.Menu))

	return row
}
