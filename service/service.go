package service

import (
	"context"
	"github.com/umerm-work/arcTest/data"
	"github.com/umerm-work/arcTest/db"
)

type Services interface {
	Login(ctx context.Context, u data.User)
	SignUp(ctx context.Context, u data.User)
}

type service struct {
	DbRepo db.Repository
	//logger     log.Logger
}

func NewBasicService(DbRepo db.Repository) Services {
	return &service{
		DbRepo: DbRepo,
	}
}
func (b *service) Login(ctx context.Context, u data.User) {

	err := b.DbRepo.Login(ctx, &u)

	if err != nil {
		return
	}

}

func (b *service) SignUp(ctx context.Context, u data.User) {


	return
}
