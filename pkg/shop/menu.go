package shop

import (
	"bot/pkg/key"
	"bot/pkg/telegram/label"
	"strconv"

	"gopkg.in/guregu/null.v4"
	t "gopkg.in/telebot.v3"
)

type CategoryMenu struct {
	ID         int         `db:"id" mapstructure:"id"`
	PositionID int         `db:"position_id" mapstructure:"position_id"`
	ParentID   null.Int    `db:"parent_id" mapstructure:",squash"`
	Name       string      `db:"name" mapstructure:"name"`
	ImageFlag  bool        `db:"image_flag" mapstructure:"image_flag"`
	Manual     null.String `db:"manual" mapstructure:",squash"`
}

func (c *CategoryMenu) Title() string {
	if c.Manual.Valid {
		return c.Manual.String
	}
	return c.Name
}

func (c *CategoryMenu) Image() bool {
	return c.ImageFlag
}

func (c *CategoryMenu) ImageName() string {
	return c.Name
}

func (c *CategoryMenu) Navigation(support string) t.Row {
	menu := &t.ReplyMarkup{}
	row := t.Row{}

	if c.ID != key.MainMenu {
		row = append(row, menu.Data(label.Back, "", key.Menu,
			strconv.FormatInt(c.ParentID.Int64, 10)))
	} else {
		return append(row, menu.URL(label.Support, support))
	}

	if c.ParentID.Valid && c.ParentID.Int64 != int64(key.MainMenu) {
		row = append(row, menu.Data(label.IMainMenu, "", key.Menu, strconv.Itoa(key.MainMenu)))
	}

	return row
}
