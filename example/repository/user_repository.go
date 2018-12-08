//go:generate smock --type=UserRepository,TeamRepository --output=../mock
package repository

import (
	"github.com/vvatanabe/smock/example/model"
)

type UserRepository interface {
	FindByID(id int) *model.User
	RemoveByID(id int)
	Create(user *model.User)
	Update(user *model.User)
}

type TeamRepository interface {
	FindByID(id int) *model.Team
	RemoveByID(id int)
	Create(user *model.Team)
	Update(user *model.Team)
}
