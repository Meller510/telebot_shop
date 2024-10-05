package telegram

import (
	"bot/pkg/key"
	"bot/pkg/shop"
	"bot/pkg/telegram/label"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	t "gopkg.in/telebot.v3"
)

const (
	capPage       int    = 10
	pageStartView string = "page"

	commandStart = "/start"
	callBack     = "\f"
)

func (b *Bot) handleUpdates() {
	b.bot.Handle(commandStart, b.viewMenu)

	b.bot.Handle(callBack+key.MenuEndP, b.menuEndPoint)

	b.bot.Handle(callBack+key.AddToBasket, b.addToBasket)

	b.bot.Handle(callBack+key.CleanBasket, b.cleanBasket)

	b.bot.Handle(callBack+key.DelItemBasket, b.delItemBasket)

	b.bot.Handle(callBack+key.Order, b.viewOrdering)

	b.bot.Handle(callBack+key.OrdersUser, b.viewOrdersUser)

	b.bot.Handle(callBack+key.OrderBasket, b.viewOrderBasket)

	b.bot.Handle(t.OnCallback, b.activityСhange)

	b.bot.Handle(callBack+key.OrderExit, b.orderExit)

	b.bot.Handle(t.OnText, b.ChekInput)

	b.bot.Handle(t.OnContact, b.ChekInput)
}

func (b *Bot) activityСhange(ctx t.Context) error {
	switch strings.TrimLeft(ctx.Args()[0], callBack) {
	case key.Menu:
		return b.viewMenu(ctx)
	case key.PriceList:
		return b.viewPrice(ctx)
	case key.Basket:
		return b.viewBasket(ctx)
	case key.Next, key.Prev:
		return b.shiftViewPage(ctx)
	}
	return errors.New("!!!!")
}

func (b *Bot) viewMenu(c t.Context) error {
	var menuID int
	var err error

	if len(c.Args()) == 0 {
		menuID = key.MainMenu
	} else {
		if menuID, err = strconv.Atoi(c.Args()[1]); err != nil {
			log.Fatalln("argID errror")
		}
	}

	menu := &t.ReplyMarkup{}
	rows := make([]t.Row, 0, 20)

	for _, v := range b.repository.FindMenu(menuID) {
		btn := menu.Data(v.Name, key.Menu, strconv.Itoa(v.ID))
		rows = append(rows, t.Row{btn})
	}

	startView, ok := c.Get(pageStartView).(int)
	if !ok {
		startView = 0
	}

	m := b.repository.FindProduct(menuID)
	for i := startView; i < capPage+startView && i != len(m); i++ {
		btn := menu.Data(m[i].Name, key.PriceList, strconv.Itoa(m[i].ID))
		rows = append(rows, t.Row{btn})
	}

	user := b.identificationUser(c)
	curMenu := b.repository.CurMenu(menuID)

	rows = append(rows, b.viewPageItems(menuID, startView),
		curMenu.Navigation(b.config.Support), b.btnBasket(user, c), b.btnOrders(user, c))

	menu.Inline(rows...)

	return b.switchTypeMsg(curMenu, menu, c)
}

func (b *Bot) viewPrice(c t.Context) error {
	var pID int
	var err error

	if pID, err = strconv.Atoi(c.Args()[1]); err != nil {
		log.Fatalln("argID errror")
	}

	menu := &t.ReplyMarkup{}
	rows := make([]t.Row, 0, 20)
	printer := message.NewPrinter(language.English)

	for i, p := range b.repository.Price(pID) {
		str := printer.Sprintf("Кол-во: %d шт. Цена: %.f ₽", p.Quantity, p.Price)
		btn := menu.Data(str, key.AddToBasket, key.PriceList, strconv.Itoa(pID), strconv.Itoa(i))
		rows = append(rows, t.Row{btn})
	}

	user := b.identificationUser(c)
	curProd := b.repository.Product(pID)
	rows = append(rows, curProd.Navigation(), b.btnBasket(user, c))

	menu.Inline(rows...)

	return b.switchTypeMsg(curProd, menu, c)
}

func (b *Bot) viewBasket(c t.Context) error {
	user := b.identificationUser(c)

	if c.Args()[0] == callBack+key.Basket {
		b.repository.AddPath(user, &shop.BackPoint{Uniq: c.Args()[1], ID: c.Args()[2]})
	} else {
		bp := b.repository.ItemPath(user)
		update := c.Update()
		update.Callback.Data = key.Basket + "|" + bp.Join()
		c = b.bot.NewContext(update)
	}

	menu := &t.ReplyMarkup{}
	rows := make([]t.Row, 0, 20)
	printer := message.NewPrinter(language.English)

	for i := range user.Basket {
		str := printer.Sprintf("%s : %d шт. %.f ₽ ✖ %d", user.Basket[i].Name,
			user.Basket[i].Quantity, user.Basket[i].Price, user.Basket[i].Count)
		btn := menu.Data(str, key.DelItemBasket, c.Args()[1], c.Args()[2], strconv.Itoa(i))
		rows = append(rows, t.Row{btn})
	}

	rows = append(rows, shop.BasketNavigation(b.identificationUser(c), c)...)
	menu.Inline(rows...)

	return b.switchTypeMsg(&shop.Basket{Name: user.Title()}, menu, c)
}

func (b *Bot) viewOrdering(c t.Context) error {
	text := fmt.Sprintf("<b>%s</b>\n\n%s", label.Order, label.OrderFullName)

	msg, err := b.bot.Send(c.Recipient(), "⚠ Введите ФИО 👇 ⚠")
	if err != nil {
		log.Fatal("беда беда")
	}

	user := b.identificationUser(c)
	user.StoredMessages.PointerMsg = t.StoredMessage{
		MessageID: strconv.Itoa(msg.ID),
		ChatID:    msg.Chat.ID,
	}

	msgMain := &user.StoredMessages.MainMenuMsg
	msgMain.MessageID, msgMain.ChatID = c.Message().MessageSig()

	menu := &t.ReplyMarkup{}
	menu.Inline(shop.OrderNavigation())

	return c.Edit(text, menu, t.ModeHTML)
}

func (b *Bot) orderExit(c t.Context) error {
	user := b.identificationUser(c)
	user.OrderStatus = sql.NullString{}

	b.bot.Delete(&user.StoredMessages.PointerMsg)
	b.bot.Send(c.Recipient(), t.RemoveKeyboard)

	update := c.Update()

	switch c.Args()[0] {
	case key.Menu:
		update.Callback.Data = key.Menu + "|" + strconv.Itoa(key.MainMenu)
		b.viewMenu(b.bot.NewContext(update))
	case key.Basket:
		b.viewBasket(b.bot.NewContext(update))
	}

	return nil
}

func (b *Bot) viewOrdersUser(c t.Context) error {
	user := b.identificationUser(c)

	menu := &t.ReplyMarkup{}
	rows := make([]t.Row, 0, 10)

	for _, ord := range user.Orders {
		text := fmt.Sprintf("№ %d Дата: %s", ord.ID, ord.Date)
		btn := menu.Data(text, key.OrderBasket, strconv.Itoa(ord.ID))
		rows = append(rows, t.Row{btn})
	}

	rows = append(rows, t.Row{menu.Data(label.IMainMenu, key.Menu, strconv.Itoa(key.MainMenu))})
	menu.Inline(rows...)

	return c.Edit(fmt.Sprintf("<b>%s</b>", label.OrdersUser), menu, t.ModeHTML)
}

func (b *Bot) viewOrderBasket(c t.Context) error {
	user := b.identificationUser(c)

	printer := message.NewPrinter(language.English)
	text := strings.Builder{}
	text.WriteString(fmt.Sprintf("<b>%s</b>\n\n", label.Basket))

	for _, ord := range user.Orders {
		if strconv.Itoa(ord.ID) == c.Args()[0] {
			var itemCount int
			var total float64 = 0

			for i, item := range ord.Basket {
				itemCount += item.Count
				total += item.Price * float64(item.Count)
				text.WriteString(printer.Sprintf("%d. <b>%s</b> : <u>%d шт. %.f ₽ ✖ %d</u>\n",
					i+1, item.Name, item.Quantity, item.Price, item.Count))
			}

			p := printer.Sprintf("\n<b>🔷 Кол-во товара : %d ед\n"+
				"🔷 Общая стоимость : %.2f ₽\n"+
				"🔷 Оплата : %s\n"+
				"🔷 Доставка : %s</b>", itemCount, total, ord.Pay, ord.Delivery)
			text.WriteString(p)
		}
	}

	menu := &t.ReplyMarkup{}
	rows := make([]t.Row, 0, 2)
	rows = append(rows, t.Row{menu.Data(label.OrdersUser, key.OrdersUser),
		menu.Data(label.IMainMenu, key.Menu, strconv.Itoa(key.MainMenu))})
	menu.Inline(rows...)

	return c.Edit(text.String(), menu, t.ModeHTML)
}

func (b *Bot) shiftViewPage(c t.Context) error {
	var pageStartPont int
	var err error

	if pageStartPont, err = strconv.Atoi(c.Args()[2]); err != nil {
		panic(err)
	}

	uniq := strings.TrimLeft(c.Args()[0], callBack)
	if uniq == key.Next {
		pageStartPont += capPage
	} else if uniq == key.Prev {
		pageStartPont -= capPage
	}

	update := c.Update()
	update.Callback.Data = key.Menu + "|" + c.Args()[1]
	context := b.bot.NewContext(update)

	context.Set(pageStartView, pageStartPont)

	return b.viewMenu(context)
}

func (b *Bot) btnBasket(user *shop.User, c t.Context) (r t.Row) {
	if len(user.Basket) > 0 {
		str := fmt.Sprintf("%s ( %d )", label.Basket, user.TotalGoods())

		menu := &t.ReplyMarkup{}

		if len(c.Args()) == 0 {
			r = append(r, menu.Data(str, key.Basket, key.Menu, strconv.Itoa(key.MainMenu)))
		} else {
			r = append(r, menu.Data(str, key.Basket, c.Args()...))
		}
	}
	return r
}

func (b *Bot) btnOrders(user *shop.User, c t.Context) (r t.Row) {
	if len(user.Orders) > 0 && (c.Callback() == nil || c.Args()[1] == strconv.Itoa(key.MainMenu)) {
		str := fmt.Sprintf("%s ( %d )", label.OrdersUser, len(user.Orders))
		menu := &t.ReplyMarkup{}
		r = append(r, menu.Data(str, key.OrdersUser, key.Menu, strconv.Itoa(key.MainMenu)))
	}
	return r
}

func (b *Bot) viewPageItems(id, startView int) (row t.Row) {
	menu := &t.ReplyMarkup{}

	if startView >= capPage {
		row = append(row, menu.Data("Пред. ⬅️", key.Prev,
			strconv.Itoa(id), strconv.Itoa(startView)))
	}

	if startView+capPage < len(b.repository.FindProduct(id)) {
		row = append(row, menu.Data("След. ➡️", key.Next,
			strconv.Itoa(id), strconv.Itoa(startView)))
	}
	return row
}

func (b *Bot) menuEndPoint(c t.Context) error {
	if c.Args()[0] == key.Menu {
		b.repository.CleanPath(b.identificationUser(c))
		return b.activityСhange(c)
	}

	bp := b.repository.TakeItemPath(b.identificationUser(c))
	update := c.Update()
	update.Callback.Data = bp.Uniq + "|" + bp.ID

	return b.activityСhange(b.bot.NewContext(update))
}
