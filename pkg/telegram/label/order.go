package label

const (
	OrderFullName = `<b><u>ФИО</u> 👉 :</b>
Номер :
Доставка :
Адрес :
Оплата :
Комментарий :`
	OrderPhoneNum = `<b>ФИО : {{.FullName}}
<u>Номер</u> 👉 :</b>
Доставка :
Адрес :
Оплата :
Комментарий :`
	OrderDelivery = `<b>ФИО : {{.FullName}}
Номер : {{.PhoneNum}}
<u>Доставка</u> 👉 :</b>
Адрес :
Оплата :
Комментарий :`
	OrderAddress = `<b>ФИО : {{.FullName}}
Номер : {{.PhoneNum}}
Доставка : {{.Delivery}}
<u>Адрес</u> 👉 :</b>
Оплата :
Комментарий :`
	OrderPay = `<b>ФИО : {{.FullName}}
Номер : {{.PhoneNum}}
Доставка : {{.Delivery}}
Адрес : {{.Address}}
<u>Оплата :</u> 👉 :</b>
Комментарий :`

	OrderComment = `<b>ФИО : {{.FullName}}
Номер : {{.PhoneNum}}
Доставка : {{.Delivery}}
Адрес : {{.Address}}
Оплата : {{.Pay}}
<u>Комментарий :</u> 👉 :</b>`

	OrderDone = `<b>✅<u> Заказ отправлен ✅
Мы свяжемся с вами в ближайшее время.</u>

Номер заказа : {{.ID}}

ФИО : {{.FullName}}
Номер : {{.PhoneNum}}
Доставка  : {{.Delivery}}
Адрес : {{.Address}}
Оплата : {{.Pay}}
Комментарий : {{.Comment}}</b>`
)
