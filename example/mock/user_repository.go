package mock

import (
	"github.com/vvatanabe/smock/example/model"
)

type UserRepositoryMock struct {
	FindByIDFunc   func(id int) *model.User
	FindByIDsFunc  func(ids []int) []*model.User
	RemoveByIDFunc func(id int)
	CreateFunc     func(user *model.User)
	UpdateFunc     func(user *model.User)
}

func (m *UserRepositoryMock) FindByID(id int) *model.User {
	if m.FindByIDFunc == nil {
		panic("This method is not defined.")
	}
	return m.FindByIDFunc(id)
}

func (m *UserRepositoryMock) FindByIDs(ids []int) []*model.User {
	if m.FindByIDsFunc == nil {
		panic("This method is not defined.")
	}
	return m.FindByIDsFunc(ids)
}

func (m *UserRepositoryMock) RemoveByID(id int) {
	if m.RemoveByIDFunc == nil {
		panic("This method is not defined.")
	}
	m.RemoveByIDFunc(id)
}

func (m *UserRepositoryMock) Create(user *model.User) {
	if m.CreateFunc == nil {
		panic("This method is not defined.")
	}
	m.CreateFunc(user)
}

func (m *UserRepositoryMock) Update(user *model.User) {
	if m.UpdateFunc == nil {
		panic("This method is not defined.")
	}
	m.UpdateFunc(user)
}
