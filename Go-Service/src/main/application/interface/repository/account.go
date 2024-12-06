package repository

import "Go-Service/src/main/domain/entity/account"

type AccountRepository interface {
	Create(account account.Account) error
	GetAll() ([]account.Account, error)
	GetByUsername(username string) (*account.Account, error)
	Update(account account.Account) error
	Delete(username string) error
}
