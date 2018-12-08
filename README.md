# smock [![Build Status](https://travis-ci.org/vvatanabe/smock.svg?branch=master)](https://travis-ci.org/vvatanabe/smock)
simple mock generator

## Description
Automatically generate simple mock code of Golang interface.

## Installation
This package can be installed with the go get command:
```
$ go get github.com/vvatanabe/smock/cmd/smock
```

Built binaries are available on Github releases: https://github.com/vvatanabe/smock/releases

## Usage
```
Usage of smock:
        smock [flags] -type T [directory] # Default: process whole package in current directory
        smock [flags] -type T files... # Must be a single package
For more information, see:
        https://godoc.org/github.com/vvatanabe/smock
Flags:
  -output string
        output directory; default process whole package in current directory
  -type string
        comma-separated list of type names; must be set
  -v    show version
```

## Example
### go generate
1. This comment anywhere in the file:
```
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
```
2. exec `go generate ./yourpkg/...` .

### generate code
```
// Code generated by smock; DO NOT EDIT.
package mock

import (
	"github.com/vvatanabe/smock/example/model"
)

type UserRepositoryMock struct {
	FindByIDFunc   func(id int) *model.User
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
```

## Bugs and Feedback
For bugs, questions and discussions please use the Github Issues.

## License
[MIT License](http://www.opensource.org/licenses/mit-license.php)

## Author
[vvatanabe](https://github.com/vvatanabe)