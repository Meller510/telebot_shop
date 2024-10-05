package postgres

import (
	"bot/pkg/config"
	"bot/pkg/shop"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"github.com/mitchellh/mapstructure"
)

type Repository struct {
	db       *sqlx.DB
	listener *pq.Listener
	menu     map[int]shop.CategoryMenu
	product  map[int]shop.Product
	user     map[int64]*shop.User
}

func ConnectDB(cfg config.DB) *sqlx.DB {
	conninfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBname, cfg.Password, cfg.SSLMode)

	db, err := sqlx.Open("postgres", conninfo)

	if err != nil {
		fmt.Errorf("%s :%w", errOpenPostgres, err)
	}

	if err = db.Ping(); err != nil {
		fmt.Errorf("%s :%w", errPingDB, err)
	}

	return db
}

func Listener(cfg config.DB) *pq.Listener {
	conninfo := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBname, cfg.Password, cfg.SSLMode)

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	minReconn := 10 * time.Second
	maxReconn := 30 * time.Second

	return pq.NewListener(conninfo, minReconn, maxReconn, reportProblem)
}

func NewRepository(cfg config.DB) *Repository {

	rep := &Repository{
		db:       ConnectDB(cfg),
		listener: Listener(cfg),
		menu:     make(map[int]shop.CategoryMenu, 50),
		product:  make(map[int]shop.Product, 100),
		user:     make(map[int64]*shop.User, 100),
	}

	rep.initMenu()
	rep.initProduct()
	rep.initUser()

	go rep.listenEvents()

	return rep
}

func (r *Repository) initMenu() {
	rows, err := r.db.Queryx(`SELECT * FROM category_menu`)

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		m := new(shop.CategoryMenu)

		if err := rows.StructScan(m); err != nil {
			log.Fatal(err)
		}

		r.menu[m.ID] = *m
	}
}

func (r *Repository) initProduct() {
	rows, err := r.db.Queryx(`SELECT * FROM product`)

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		p := new(shop.Product)

		var price types.JSONText

		if err := rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.Manual, &price, &p.ImageFlag); err != nil {
			log.Fatal(err)
		}

		if len(price) != 0 {
			if err = price.Unmarshal(&p.PriceList); err != nil {
				log.Fatal(fmt.Errorf("%w,%s", errUnmarshalPriceProd, p.Name))
			}
		}

		r.product[p.ID] = *p
	}
}

func (r *Repository) initUser() {
	rows, err := r.db.Queryx(`SELECT * FROM "user"`)

	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var (
			user   = new(shop.User)
			basket types.JSONText
			way    types.JSONText
		)

		if err := rows.Scan(&user.ID, &user.FirstName, &user.UserName,
			pq.Array(&user.PhoneNumbers), &user.DateLastVisit, &user.DateRegistration, &basket, &way); err != nil {
			log.Fatal(err)
		}

		if len(basket) != 0 {
			if err = basket.Unmarshal(&user.Basket); err != nil {
				log.Fatal(err)
			}
		}

		if len(way) != 0 {
			if err = way.Unmarshal(&user.WayBack); err != nil {
				log.Fatal(err)
			}
		}

		r.user[user.ID] = user
	}
}

func (r *Repository) FindMenu(parentID int) []shop.CategoryMenu {
	val, ok := r.menu[parentID]
	if !ok {
		panic("FindMenu not found parent_ID")
	}

	list := make([]shop.CategoryMenu, 0, 20)

	for k := range r.menu {
		if r.menu[k].ParentID.Int64 == int64(val.ID) {
			list = append(list, r.menu[k])
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].PositionID < list[j].PositionID
	})

	return list
}

func (r *Repository) FindProduct(parentID int) []shop.Product {
	list := make([]shop.Product, 0, 20)

	for k := range r.product {
		if r.product[k].CategoryID == parentID {
			list = append(list, r.product[k])
		}
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}

func (r *Repository) Price(productID int) []shop.PriceItem {
	return r.product[productID].PriceList
}

func (r *Repository) CurMenu(parentID int) *shop.CategoryMenu {
	menu, ok := r.menu[parentID]
	if !ok {
		panic("Parent not found")
	}

	return &menu
}

func (r *Repository) HeaderMenu(id int) string {
	return fmt.Sprintf("<b>%s</b>", r.menu[id].Name)
}

func (r *Repository) HeaderProduct(id int) string {
	return fmt.Sprintf("<b> %s </b>", r.product[id].Name)
}

func (r *Repository) Product(productID int) *shop.Product {
	p, ok := r.product[productID]
	if !ok {
		log.Fatal("продукт не найден ")
	}
	return &p
}

func (r *Repository) AddOrder(user *shop.User) {
	basket, err := json.Marshal(user.Basket)

	if err != nil {
		log.Fatal(err)
	}

	tx := r.db.MustBegin()

	if err := tx.QueryRow(`INSERT INTO "order" (user_id, full_name, phone, pay, delivery, address, comment, date,basket) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		user.ID, user.Order.FullName, user.Order.PhoneNum, user.Order.Pay, user.Order.Delivery,
		user.Order.Address, user.Order.Comment, user.Order.Date, basket).Scan(&user.Order.ID); err != nil {
		log.Fatal(err)
	}

	addUserNumber(tx, user)

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (r *Repository) AddUser(user *shop.User) {
	tx := r.db.MustBegin()

	tx.MustExec(`INSERT INTO "user" (id, first_name, user_name, date_registration) VALUES ($1, $2, $3, $4)`,
		user.ID, user.FirstName.String, user.UserName.String, user.DateRegistration.Time)

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	r.user[user.ID] = user
}

func (r *Repository) SelectUser(userID int64, date int64) (*shop.User, error) {
	user, ok := r.user[userID]
	if !ok {
		return nil, errors.New("user not found in cache")
	}

	if time.Unix(date, 0).Format("2000-01-30") != user.DateLastVisit.Time.Format("2000-01-30") {
		tx := r.db.MustBegin()
		tx.MustExec(`UPDATE "user" SET date_last_visit=$1 WHERE id = $2`,
			time.Unix(date, 0), user.ID)

		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}

		r.user[userID].DateLastVisit.Time = time.Unix(date, 0)
	}

	r.userOrders(user)
	return user, nil
}

func (r *Repository) CleanBasket(user *shop.User) {
	if len(user.Basket) != 0 {
		user.Basket = nil
		r.updateBasket(user)
	}
}

func (r *Repository) DelItemBasket(user *shop.User, idItem int) {
	if user.Basket[idItem].Count > 1 {
		user.Basket[idItem].Count--
	} else {
		user.Basket = append(user.Basket[:idItem], user.Basket[idItem+1:]...)
	}

	r.updateBasket(user)
}

func (r *Repository) AddProductBasket(user *shop.User, bItem *shop.BasketItem) {
	itemFound := true

	for i := range user.Basket {
		if user.Basket[i].Name == bItem.Name && user.Basket[i].Quantity == bItem.Quantity {
			user.Basket[i].Count++

			itemFound = false

			break
		}
	}

	if itemFound {
		user.Basket = append(user.Basket, *bItem)
	}

	r.updateBasket(user)
}

func (r *Repository) updateBasket(user *shop.User) {
	item, err := json.Marshal(user.Basket)

	if err != nil {
		log.Fatal(err)
	}

	tx := r.db.MustBegin()
	tx.MustExec(`UPDATE "user" SET basket=$1 WHERE id = $2`,
		item, user.ID)

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (r *Repository) AddPath(user *shop.User, point *shop.BackPoint) {
	user.WayBack = append(user.WayBack, *point)
	r.updatePath(user)
}

func (r *Repository) TakeItemPath(user *shop.User) shop.BackPoint {
	bp := r.ItemPath(user)
	user.WayBack = user.WayBack[:len(user.WayBack)-1]
	r.updatePath(user)

	return bp
}

func (r *Repository) ItemPath(user *shop.User) shop.BackPoint {
	return user.WayBack[len(user.WayBack)-1]
}

func (r *Repository) CleanPath(user *shop.User) {
	user.WayBack = nil
	r.updatePath(user)
}

func (r *Repository) updatePath(user *shop.User) {
	item, err := json.Marshal(user.WayBack)

	if err != nil {
		log.Fatal(err)
	}

	tx := r.db.MustBegin()
	tx.MustExec(`UPDATE "user" SET way_back=$1 WHERE id = $2`,
		item, user.ID)

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
}

func (r *Repository) userOrders(user *shop.User) {
	var count int
	rows := r.db.QueryRow(`SELECT COUNT(user_id) FROM "order" WHERE user_id=$1`, user.ID)
	if err := rows.Scan(&count); err != nil {
		panic(err)
	}

	if count > len(user.Orders) {
		user.Orders = nil
		rows, err := r.db.Queryx(`SELECT * FROM "order" WHERE user_id=$1`, user.ID)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			order := new(shop.Order)
			var basket types.JSONText

			if err := rows.Scan(&order.ID, &order.UserID, &order.FullName, &order.PhoneNum, &order.Pay, &order.Delivery, &order.Address, &order.Comment, &order.Date, &basket); err != nil {
				log.Fatal(err)
			}

			if len(basket) != 0 {
				if err = basket.Unmarshal(&order.Basket); err != nil {
					log.Fatal(err)
				}
			}

			user.Orders = append(user.Orders, *order)
		}
	}
}

func addUserNumber(tx *sqlx.Tx, user *shop.User) {
	addNumber := true
	for _, val := range user.PhoneNumbers {
		if user.Order.PhoneNum == val.String && val.Valid {
			addNumber = false
			break
		}
	}

	if addNumber {
		tx.MustExec(`UPDATE "user" SET phone_numbers=array_append(phone_numbers,$1) WHERE id = $2`,
			user.Order.PhoneNum, user.ID)
	}
}

func (r *Repository) listenEvents() {
	err := r.listener.Listen("events")
	if err != nil {
		panic(err)
	}
	for {
		r.waitForNotification()
	}
}

func (r *Repository) waitForNotification() {
	for {
		select {
		case n := <-r.listener.Notify:
			r.getUpdate(n)
			return
		case <-time.After(2 * time.Second):
			go func() {
				r.listener.Ping()
			}()
			return
		}
	}
}

func (r *Repository) getUpdate(n *pq.Notification) {
	var m interface{}
	if err := json.Unmarshal([]byte(n.Extra), &m); err != nil {
		panic(err)
	}

	e := Event{}
	if err := mapstructure.Decode(m, &e); err != nil {
		panic(err)
	}

	switch e.Table {
	case "product":
		r.updateTableProduct(&e)
	case "category_menu":
		r.updateTableCategoryMenu(&e)
	}
}

func (r *Repository) updateTableProduct(e *Event) {
	var p Product
	if err := mapstructure.Decode(e.Data, &p); err != nil {
		log.Panic(err)
	}

	p.Product.Manual.Scan(p.Manual)

	switch e.Action {
	case "DELETE":
		delete(r.product, p.ID)
	case "INSERT", "UPDATE":
		r.product[p.ID] = p.Product
	}
}

func (r *Repository) updateTableCategoryMenu(e *Event) {
	var menu CategoryMenu
	if err := mapstructure.Decode(e.Data, &menu); err != nil {
		log.Panic(err)
	}

	menu.CategoryMenu.Manual.Scan(menu.Manual)
	menu.CategoryMenu.ParentID.Scan(menu.ParentID)

	switch e.Action {
	case "DELETE":
		delete(r.menu, menu.ID)
	case "INSERT", "UPDATE":
		r.menu[menu.ID] = menu.CategoryMenu
	}
}
