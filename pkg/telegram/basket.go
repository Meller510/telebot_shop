package telegram

import (
	"bot/pkg/shop"
	"fmt"
	"log"
	"strconv"
	"strings"

	t "gopkg.in/telebot.v3"
)

func (b *Bot) addToBasket(c t.Context) error {
	prodID, err := strconv.Atoi(c.Args()[1])
	if err != nil {
		log.Fatal("беда беда ")
	}

	priceID, err := strconv.Atoi(c.Args()[2])
	if err != nil {
		log.Fatal("беда беда")
	}

	prod := b.repository.Product(prodID)

	b.repository.AddProductBasket(b.identificationUser(c),
		shop.NewBasketItem(prod.Name, prod.PriceList[priceID]))

	go c.Respond(&t.CallbackResponse{
		Text: fmt.Sprintf("Добавлено %s кол-во: %d шт.",
			prod.Name, prod.PriceList[priceID].Quantity)})

	update := c.Update()
	args := c.Args()[:len(c.Args())-1]

	update.Callback.Data = strings.Join(args, "|")

	return b.viewPrice(b.bot.NewContext(update))
}

func (b *Bot) cleanBasket(c t.Context) error {
	b.repository.CleanBasket(b.identificationUser(c))

	update := c.Update()
	update.Callback.Data = c.Callback().Unique + "|" + c.Data()

	return b.viewBasket(b.bot.NewContext(update))
}

func (b *Bot) delItemBasket(c t.Context) error {
	idPrice, err := strconv.Atoi(c.Args()[2])

	if err != nil {
		log.Fatal("ай ай")
	}

	b.repository.DelItemBasket(b.identificationUser(c), idPrice)

	update := c.Update()
	update.Callback.Data = c.Callback().Unique + "|" + c.Data()

	return b.viewBasket(b.bot.NewContext(update))
}
