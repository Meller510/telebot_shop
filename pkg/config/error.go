package config

import "errors"

var (
	errReadConfig   = errors.New("not found file config 'config.yml'. Path:/configs")
	errImagePath    = errors.New("pathImage: empty or not found")
	errSupport      = errors.New("errSupport: empty or not found")
	errVarToken     = errors.New(" TOKEN ")
	errDBUserName   = errors.New("DB_UserName empty")
	errDBPassword   = errors.New("DB_Password empty")
	errMailAddress  = errors.New("MailAddress empty")
	errMailPassword = errors.New("MailPassword empty")
	errFoundEnvFile = errors.New(`не найден файл конфигурации .env
	создайте файл в корне приложения с переменными :
	TOKEN=ваш токен
	DB_USERNAME=имя пользоавтеля базы данных
	DB_PASSWORD=пароль
	MAIL_ADDRESS=адрес отправки заказа
	MAIL_PASSWORD=пароль`)
)

func checkENV(cfg *Config) error {
	var err error

	switch {
	case cfg.TelegramToken == "":
		err = errVarToken
	case cfg.DataBase.Username == "":
		err = errDBUserName
	case cfg.DataBase.Password == "":
		err = errDBPassword
	case cfg.Email.Address == "":
		err = errMailAddress
	case cfg.Email.Password == "":
		err = errMailPassword
	}

	return err
}
