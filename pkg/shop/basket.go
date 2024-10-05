package shop

import (
	"bot/pkg/key"
	"bot/pkg/telegram/label"
	"strconv"

	t "gopkg.in/telebot.v3"
)

type BasketItem struct {
	ID    int
	Name  string `json:"name"`
	Count int    `json:"count"`
	PriceItem
}

func NewBasketItem(name string, price PriceItem) *BasketItem {
	return &BasketItem{
		Name:      name,
		PriceItem: price,
		Count:     1,
	}
}

type Basket struct {
	Name string
}

func BasketNavigation(user *User, c t.Context) (rows []t.Row) {
	menu := &t.ReplyMarkup{}

	if len(user.Basket) > 0 {
		rows = append(rows, t.Row{menu.Data(label.Clean, key.CleanBasket, c.Args()[1], c.Args()[2]),
			menu.Data(label.SendOrder, key.Order, c.Callback().Unique, c.Args()[0], c.Args()[1])})
	}

	row := make([]t.Btn, 0, 2)

	if c.Args()[2] != strconv.Itoa(key.MainMenu) {
		row = append(row, menu.Data(label.Back, key.MenuEndP, c.Args()[1], c.Args()[2]))
	}

	row = append(row, menu.Data(label.IMainMenu, key.MenuEndP, key.Menu, strconv.Itoa(key.MainMenu)))
	rows = append(rows, row)
	return rows
}

func (b *Basket) Title() string {
	return b.Name
}

func (b *Basket) Image() bool {
	return false
}

func (b *Basket) ImageName() string {
	return ""
}
