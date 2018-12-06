//go:generate /home/vvatanabe/go/src/github.com/vvatanabe/smock/dist/current/smock --type=UserRepository --output=../mock
package repository

import (
	"github.com/vvatanabe/smock/example/model"
)

type UserRepository interface {
	FindByID(id int) *model.User
	FindByIDs(ids []int) []*model.User
	RemoveByID(id int)
	Create(user *model.User)
	Update(user *model.User)
}
