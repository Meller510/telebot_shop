package postgres

import "errors"

var (
	errOpenPostgres       = errors.New("error open data base")
	errPingDB             = errors.New("error ping data base")
	errUnmarshalPriceProd = errors.New("unmarshal price product ")
)
