package repository

import "bot/pkg/shop"

type Repository interface {
	HeaderMenu(id int) string
	HeaderProduct(id int) string
	CurMenu(parentID int) *shop.CategoryMenu
	FindMenu(parentID int) []shop.CategoryMenu
	FindProduct(parentID int) []shop.Product
	Price(productID int) []shop.PriceItem
	Product(productID int) *shop.Product
	SelectUser(userID int64, date int64) (*shop.User, error)
	AddOrder(user *shop.User)
	AddUser(user *shop.User)
	CleanBasket(user *shop.User)
	DelItemBasket(user *shop.User, idItem int)
	AddProductBasket(user *shop.User, p *shop.BasketItem)
	AddPath(user *shop.User, point *shop.BackPoint)
	CleanPath(user *shop.User)
	TakeItemPath(user *shop.User) shop.BackPoint
	ItemPath(user *shop.User) shop.BackPoint
}
