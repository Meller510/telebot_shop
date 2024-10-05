package label

import (
	"github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("namebtn")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatal("not found file config 'namebtn.yml' Path:/assets/label")
	}

	Support = viper.GetString("support")
	Back = viper.GetString("back")
	IMainMenu = viper.GetString("mainMenu")
	SendOrder = viper.GetString("sendOrder")
	Clean = viper.GetString("cleanse")
	Basket = viper.GetString("basket")
	Order = viper.GetString("order")
	OrdersUser = viper.GetString("yourOrders")
	Empty = viper.GetString("empty")
	IPickup = viper.GetString("pickup")
	IMail = viper.GetString("mail")
	ICourier = viper.GetString("courier")
	IPhoneNumber = viper.GetString("phoneNumber")
	IComment = viper.GetString("comment")
	IPayCash = viper.GetString("payCash")
	IPayNonCash = viper.GetString("payNon-cash")
	ICOD = viper.GetString("c.o.d")
}

var (
	Support      string
	Back         string
	IMainMenu    string
	SendOrder    string
	Clean        string
	Basket       string
	Order        string
	OrdersUser   string
	Empty        string
	ICOD         string
	IPayNonCash  string
	IPayCash     string
	IComment     string
	IPickup      string
	IMail        string
	ICourier     string
	IPhoneNumber string
)
