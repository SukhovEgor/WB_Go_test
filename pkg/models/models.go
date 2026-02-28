package models

import (
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

type Order struct {
	OrderUID          string    `json:"order_uid" fake:"{uuid}"`
	TrackNumber       string    `json:"track_number"`
	Entry             string    `json:"entry"`
	Delivery          Delivery  `json:"delivery" fake:"skip"`
	Payment           Payment   `json:"payment" fake:"skip"`
	Items             []Item    `json:"items" fake:"skip"`
	Locale            string    `json:"locale" fake:"{languageabbreviation}"`
	InternalSignature string    `json:"internal_signature" fake:"skip"`
	CustomerID        string    `json:"customer_id" fake:"{uuid}""`
	DeliveryService   string    `json:"delivery_service" fake:"{company}"`
	Shardkey          string    `json:"shardkey"`
	SmID              int       `json:"sm_id" fake:"{number:1,100}"`
	DateCreated       time.Time `json:"date_created"`
	OofShard          string    `json:"oof_shard"`
}

type Delivery struct {
	OrderUID string `json:"-"`
	Name     string `json:"name" fake:"{name}"`
	Phone    string `json:"phone" fake:"{phone}"`
	Zip      string `json:"zip"`
	City     string `json:"city" fake:"{city}"`
	Address  string `json:"address" fake:"{street}"`
	Region   string `json:"region" fake:"{state}"`
	Email    string `json:"email" fake:"{email}"`
}

type Payment struct {
	OrderUID     string  `json:"-" db:"order_uid"`
	Transaction  string  `json:"transaction" fake:"{uuid}"`
	RequestID    string  `json:"request_id" fake:"{uuid}"`
	Currency     string  `json:"currency" fake:"{currencyshort}"`
	Provider     string  `json:"provider" fake:"{company}"`
	Amount       float64 `json:"amount"`
	PaymentDt    int     `json:"payment_dt" fake:"{number:100,1000}"`
	Bank         string  `json:"bank" fake:"{bankname}"`
	DeliveryCost float64 `json:"delivery_cost" fake:"{price:1,1000}"`
	GoodsTotal   float64 `json:"goods_total"`
	CustomFee    float64 `json:"custom_fee" fake:"{price:1,1000}"`
}

type Item struct {
	ID          int     `json:"-"`
	OrderUID    string  `json:"-"`
	ChrtID      int64   `json:"chrt_id"fake:"{number:1,10000}"`
	TrackNumber string  `json:"track_number" `
	Price       int     `json:"price" fake:"{number:1000,10000}"`
	Rid         string  `json:"rid" fake:"{uuid}"`
	Name        string  `json:"name" fake:"{productname}"`
	Sale        int     `json:"sale" fake:"{number:0,100}"`
	Size        string  `json:"size" fake:"{number:0,100}"`
	TotalPrice  float64 `json:"total_price"`
	NmID        int64   `json:"nm_id" fake:"{number:10000,99999}"`
	Brand       string  `json:"brand" fake:"{company}"`
	Status      int     `json:"status" fake:"{number:200,202}"`
}

func createRandomOrder(rng *rand.Rand) Order {
	var order Order
	var delivery Delivery
	var payment Payment

	if err := gofakeit.Struct(&order); err != nil {
		return order
	}
	if err := gofakeit.Struct(&delivery); err != nil {
		return order
	}
	if err := gofakeit.Struct(&payment); err != nil {
		return order
	}

	itemCount := rng.Intn(10) + 1
	items := make([]Item, 0, itemCount)
	var goodsTotal float64

	for i := 0; i < itemCount; i++ {
		var item Item
		if err := gofakeit.Struct(&item); err != nil {
			return order
		}

		quantity := rng.Intn(5) + 1

		item.Sale = rng.Intn(51)

		item.TotalPrice = float64(item.Price*quantity) * (1 - float64(item.Sale)/100.0)

		goodsTotal += item.TotalPrice
		items = append(items, item)
	}

	payment.GoodsTotal = goodsTotal
	payment.Amount = payment.DeliveryCost + goodsTotal + payment.CustomFee

	order.Delivery = delivery
	order.Payment = payment
	order.Items = items

	return order
}
