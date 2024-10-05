package postgres

import "bot/pkg/shop"

type Event struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	Data   interface{}
}

type EventTableProduct struct {
	Table  string `json:"table"`
	Action string `json:"action"`
	Data   shop.Product
}

type CategoryMenu struct {
	shop.CategoryMenu `mapstructure:",squash"`
	ParentID          int64  `mapstructure:"parent_id"`
	Manual            string `mapstructure:"manual"`
}

type Product struct {
	shop.Product `mapstructure:",squash"`
	Manual       string `mapstructure:"manual"`
}
