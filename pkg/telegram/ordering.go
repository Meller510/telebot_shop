package telegram

import (
	"bot/pkg/key"
	"bot/pkg/shop"
	"bot/pkg/telegram/label"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/looplab/fsm"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	t "gopkg.in/telebot.v3"
)

const (
	name      = "name"
	phoneNum  = "phoneNum"
	payType   = "payType"
	delivery  = "delivery"
	address   = "address"
	comment   = "comment"
	sendOrder = "sendOrder"
)

func (b *Bot) ChekInput(c t.Context) error {
	user := b.identificationUser(c)
	ord := NewOrdering(user)

	if err := b.bot.Delete(&user.StoredMessages.PointerMsg); err != nil {
		panic(err)
	}

	if err := b.bot.Delete(c.Message()); err != nil {
		panic(err)
	}

	switch ord.FSM.Current() {
	case name:
		if err := ord.FSM.Event(phoneNum, c.Message().Text); err != nil {
			panic(err)
		}

		return viewInputPhoneNumber(user, c)
	case phoneNum:
		var number string
		if c.Message().Contact != nil {
			number = c.Message().Contact.PhoneNumber
		} else {
			number = c.Message().Text
		}

		if err := ord.FSM.Event(delivery, number); err != nil {
			panic(err)
		}
		return viewInputDelivery(user, c)
	case delivery:
		if c.Message().Text == label.IPickup {
			user.Basket = append(user.Basket,
				*shop.NewBasketItem(label.IPickup, shop.PriceItem{
					Quantity: 1,
					Price:    -200,
				}))
			if err := ord.FSM.Event(address, c.Message().Text); err != nil {
				panic(err)
			}
			if err := ord.FSM.Event(payType, label.Empty); err != nil {
				panic(err)
			}
			return viewInputPayType(user, c)
		} else {
			if err := ord.FSM.Event(address, c.Message().Text); err != nil {
				panic(err)
			}
			return viewInputAddress(user, c)
		}
	case address:
		if err := ord.FSM.Event(payType, c.Message().Text); err != nil {
			panic(err)
		}
		return viewInputPayType(user, c)
	case payType:
		if user.Order.Pay == label.ICOD {
			user.Basket = append(user.Basket,
				*shop.NewBasketItem(label.ICOD, shop.PriceItem{
					Quantity: 1,
					Price:    300,
				}))
		}
		if err := ord.FSM.Event(comment, c.Message().Text); err != nil {
			panic(err)
		}

		return viewInputComment(user, c)
	case comment:
		if err := ord.FSM.Event(sendOrder, c); err != nil {
			panic(err)
		}

		b.repository.AddOrder(user)

		tmpl, err := template.New("mail.html").Funcs(template.FuncMap{
			"PriceFormat": func(price float64) string {
				p := message.NewPrinter(language.English)
				return p.Sprintf("%.f", price)
			}}).ParseFiles("templates/mail.html")

		if err != nil {
			panic(err)
		}

		var body strings.Builder
		if err = tmpl.Execute(&body, user); err != nil {
			panic(err)
		}

		b.sendMail(user, fmt.Sprintf("Заказ № %d", user.Order.ID), body.String())
		b.repository.CleanBasket(user)
		return viewOrderDone(user, c)
	}

	return nil
}

type Ordering struct {
	user *shop.User
	FSM  *fsm.FSM
}

func NewOrdering(u *shop.User) *Ordering {
	o := &Ordering{user: u}

	if !o.user.OrderStatus.Valid || o.user.OrderStatus.String == sendOrder {
		o.user.OrderStatus = sql.NullString{String: name}
	}

	o.FSM = fsm.NewFSM(
		u.OrderStatus.String,
		fsm.Events{
			{Name: phoneNum, Src: []string{name}, Dst: phoneNum},
			{Name: delivery, Src: []string{phoneNum}, Dst: delivery},
			{Name: address, Src: []string{delivery}, Dst: address},
			{Name: payType, Src: []string{address}, Dst: payType},
			{Name: comment, Src: []string{payType}, Dst: comment},
			{Name: sendOrder, Src: []string{comment}, Dst: sendOrder},
		},

		fsm.Callbacks{
			"before_" + phoneNum:  func(e *fsm.Event) { o.name(e) },
			"before_" + delivery:  func(e *fsm.Event) { o.phoneNum(e) },
			"before_" + address:   func(e *fsm.Event) { o.delivery(e) },
			"before_" + payType:   func(e *fsm.Event) { o.address(e) },
			"before_" + comment:   func(e *fsm.Event) { o.payType(e) },
			"before_" + sendOrder: func(e *fsm.Event) { o.comment(e) },
		},
	)
	return o
}

func (o *Ordering) name(e *fsm.Event) {
	o.user.Order.FullName = e.Args[0].(string)
	o.user.OrderStatus = sql.NullString{String: phoneNum, Valid: true}
}

func (o *Ordering) phoneNum(e *fsm.Event) {
	o.user.Order.PhoneNum = e.Args[0].(string)
	o.user.OrderStatus = sql.NullString{String: delivery, Valid: true}
}

func (o *Ordering) payType(e *fsm.Event) {
	o.user.Order.Pay = e.Args[0].(string)
	o.user.OrderStatus = sql.NullString{String: comment, Valid: true}
}

func (o *Ordering) delivery(e *fsm.Event) {
	o.user.Order.Delivery = e.Args[0].(string)
	o.user.OrderStatus = sql.NullString{String: address, Valid: true}
}

func (o *Ordering) address(e *fsm.Event) {
	o.user.Order.Address = e.Args[0].(string)
	o.user.OrderStatus = sql.NullString{String: payType, Valid: true}
}

func (o *Ordering) comment(e *fsm.Event) {
	c := e.Args[0].(t.Context)
	o.user.Order.Comment = c.Message().Text
	o.user.Order.Date = time.Unix(c.Message().Unixtime, 0).Format("2006-01-02 15:04")
	o.user.OrderStatus = sql.NullString{String: sendOrder, Valid: true}
}

func viewInputPhoneNumber(user *shop.User, c t.Context) error {
	btn := &t.ReplyMarkup{ResizeKeyboard: true}
	btn.Reply(btn.Row(btn.Contact(label.IPhoneNumber)))

	msg, err := c.Bot().Send(c.Recipient(), "⚠ Введите номер 👇 ⚠", btn)
	if err != nil {
		panic(err)
	}

	user.StoredMessages.PointerMsg = t.StoredMessage{
		MessageID: strconv.Itoa(msg.ID),
		ChatID:    msg.Chat.ID,
	}

	parse, err := template.New("").Parse(label.OrderPhoneNum)
	if err != nil {
		panic(err)
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf("<b>%s</b>\n\n", label.Order))

	if err := parse.Execute(&body, user.Order); err != nil {
		panic(err)
	}

	menu := &t.ReplyMarkup{}
	menu.Inline(shop.OrderNavigation())

	_, err = c.Bot().Edit(&user.StoredMessages.MainMenuMsg, body.String(), menu, t.ModeHTML)

	return err
}

func viewInputPayType(user *shop.User, c t.Context) error {
	menu := &t.ReplyMarkup{ResizeKeyboard: true}
	var btn t.Btn
	var text string

	if user.Order.Delivery == label.IMail {
		btn = menu.Text(label.ICOD)
		text = `❗❗❗Напоминаем❗❗❗
		 В случае выбора "Наложенный платеж 🧾"
		 сумма покупки увеличивается на 300 ₽.  
		 ⚠ Выберете тип оплаты 👇 ⚠`
	} else {
		btn = menu.Text(label.IPayCash)
		text = "⚠ Выберете тип оплаты 👇 ⚠"
	}

	menu.Reply(t.Row{btn, menu.Text(label.IPayNonCash)})

	msg, err := c.Bot().Send(c.Recipient(), text, menu)
	if err != nil {
		panic(err)
	}

	user.StoredMessages.PointerMsg = t.StoredMessage{
		MessageID: strconv.Itoa(msg.ID),
		ChatID:    msg.Chat.ID,
	}

	parse, err := template.New("").Parse(label.OrderPay)
	if err != nil {
		panic(err)
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf("<b>%s</b>\n\n", label.Order))

	if err := parse.Execute(&body, user.Order); err != nil {
		panic(err)
	}

	menu = &t.ReplyMarkup{}
	menu.Inline(shop.OrderNavigation())

	_, err = c.Bot().Edit(&user.StoredMessages.MainMenuMsg, body.String(), menu, t.ModeHTML)

	return err
}

func viewInputDelivery(user *shop.User, c t.Context) error {
	btn := &t.ReplyMarkup{ResizeKeyboard: true}
	btn.Reply(btn.Row(btn.Text(label.IPickup), btn.Text(label.IMail), btn.Text(label.ICourier)))

	msg, err := c.Bot().Send(c.Recipient(), "⚠ Выберете тип доставки 👇 ⚠", btn)
	if err != nil {
		panic(err)
	}

	user.StoredMessages.PointerMsg = t.StoredMessage{
		MessageID: strconv.Itoa(msg.ID),
		ChatID:    msg.Chat.ID,
	}

	parse, err := template.New("").Parse(label.OrderDelivery)
	if err != nil {
		panic(err)
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf("<b>%s</b>\n\n", label.Order))

	if err := parse.Execute(&body, user.Order); err != nil {
		panic(err)
	}

	menu := &t.ReplyMarkup{}
	menu.Inline(shop.OrderNavigation())

	_, err = c.Bot().Edit(&user.StoredMessages.MainMenuMsg, body.String(), menu, t.ModeHTML)

	return err
}

func viewInputAddress(user *shop.User, c t.Context) error {
	text := `🔷Пример: г.Новосибирск ул.Ленина 5 кв.78
🔷Индекс : 630099
⚠ Введите адресс доставки 👇 ⚠`
	msg, err := c.Bot().Send(c.Recipient(), text)
	if err != nil {
		panic(err)
	}

	user.StoredMessages.PointerMsg = t.StoredMessage{
		MessageID: strconv.Itoa(msg.ID),
		ChatID:    msg.Chat.ID,
	}

	parse, err := template.New("").Parse(label.OrderAddress)
	if err != nil {
		panic(err)
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf("<b>%s</b>\n\n", label.Order))

	if err := parse.Execute(&body, user.Order); err != nil {
		panic(err)
	}

	menu := &t.ReplyMarkup{}
	menu.Inline(shop.OrderNavigation())

	_, err = c.Bot().Edit(&user.StoredMessages.MainMenuMsg, body.String(), menu, t.ModeHTML)

	return err
}

func viewInputComment(user *shop.User, c t.Context) error {
	btn := &t.ReplyMarkup{ResizeKeyboard: true}
	btn.Reply(btn.Row(btn.Text(label.IComment)))

	msg, err := c.Bot().Send(c.Recipient(), "⚠ Введите комментарий 👇 ⚠", btn)
	if err != nil {
		panic(err)
	}

	user.StoredMessages.PointerMsg = t.StoredMessage{
		MessageID: strconv.Itoa(msg.ID),
		ChatID:    msg.Chat.ID,
	}

	parse, err := template.New("").Parse(label.OrderComment)
	if err != nil {
		panic(err)
	}

	var body strings.Builder
	body.WriteString(fmt.Sprintf("<b>%s</b>\n\n", label.Order))

	if err := parse.Execute(&body, user.Order); err != nil {
		panic(err)
	}

	menu := &t.ReplyMarkup{}
	menu.Inline(shop.OrderNavigation())

	_, err = c.Bot().Edit(&user.StoredMessages.MainMenuMsg, body.String(), menu, t.ModeHTML)

	return err
}

func viewOrderDone(user *shop.User, c t.Context) error {
	parse, err := template.New("").Parse(label.OrderDone)
	if err != nil {
		panic(err)
	}

	var body strings.Builder
	if err := parse.Execute(&body, user.Order); err != nil {
		panic(err)
	}

	menu := &t.ReplyMarkup{}

	menu.Inline(t.Row{menu.Data(label.IMainMenu, key.Menu, strconv.Itoa(key.MainMenu))})

	_, err = c.Bot().Edit(&user.StoredMessages.MainMenuMsg, body.String(), menu, t.ModeHTML)
	return err
}
