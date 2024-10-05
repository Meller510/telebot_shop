package shop

import (
	"bot/pkg/key"
	"bot/pkg/telegram/label"
	"database/sql"
	"strconv"

	t "gopkg.in/telebot.v3"
)

type Product struct {
	ID         int            `mapstructure:"id"`
	CategoryID int            `db:"category_id" mapstructure:"category_id"`
	Name       string         `mapstructure:"name"`
	Manual     sql.NullString `mapstructure:",squash"`
	PriceList  []PriceItem    `db:"price_list" mapstructure:"price_list"`
	ImageFlag  bool           `mapstructure:"image_flag"`
}

type PriceItem struct {
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

func (p *Product) Title() string {
	if p.Manual.Valid {
		return p.Manual.String
	}
	return p.Name
}

func (p *Product) Image() bool {
	return p.ImageFlag
}

func (p *Product) ImageName() string {
	return p.Name
}

func (p *Product) Navigation() (r t.Row) {
	menu := &t.ReplyMarkup{}

	r = append(r, menu.Data(label.Back, key.Menu, strconv.Itoa(p.CategoryID)),
		menu.Data(label.IMainMenu, key.Menu, strconv.Itoa(key.MainMenu)))

	return r
}
