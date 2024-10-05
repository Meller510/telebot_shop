package shop

import (
	"bot/pkg/telegram/label"
	"database/sql"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"

	t "gopkg.in/telebot.v3"
)

type User struct {
	ID               int64            `db:"id"`
	FirstName        sql.NullString   `db:"first_name"`
	UserName         sql.NullString   `db:"user_name"`
	PhoneNumbers     []sql.NullString `db:"phone_numbers"`
	DateLastVisit    sql.NullTime     `db:"date_last_visit"`
	DateRegistration sql.NullTime     `db:"date_registration"`
	Basket           []BasketItem     `db:"basket"`
	WayBack          []BackPoint      `db:"way_back"`
	Order            Order            `db:"order"`
	Orders           []Order
	OrderStatus      sql.NullString `db:"cur_order_status"`
	StoredMessages   StoredMessages
}

func NewUser(id int64, fName string, uNane string, date int64) *User {
	return &User{
		ID:               id,
		FirstName:        sql.NullString{String: fName},
		UserName:         sql.NullString{String: uNane},
		DateRegistration: sql.NullTime{Time: time.Unix(date, 0)},
	}
}

func (u *User) TotalCost() (total float64) {
	for i := range u.Basket {
		total += u.Basket[i].Price * float64(u.Basket[i].Count)
	}

	return total
}

func (u *User) TotalGoods() (total int) {
	for i := range u.Basket {
		total += u.Basket[i].Count
	}

	return total
}

func (u *User) Title() (title string) {
	printer := message.NewPrinter(language.English)

	if len(u.Basket) > 0 {
		title = printer.Sprintf("<b>%s\n🔷 Кол-во товара : %d ед\n🔷 Общая стоимость : %.2f ₽</b>\n\n"+
			"⚠ Нажмите на продукцию для удаления", label.Basket, u.TotalGoods(), u.TotalCost())
	} else {
		title = printer.Sprintf("<b>%s\n⛔ Пусто</b>", label.Basket)
	}

	return title
}

type StoredMessages struct {
	MainMenuMsg t.StoredMessage
	PointerMsg  t.StoredMessage
}

type BackPoint struct {
	ID   string
	Uniq string
}

func (bp *BackPoint) Join() string {
	return bp.Uniq + "|" + bp.ID
}
