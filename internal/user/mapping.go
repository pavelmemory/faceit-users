package user

import (
	"github.com/pavelmemory/faceit-users/internal/storage"
)

func userEntity(u storage.User) Entity {
	return Entity{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Nickname:  u.Nickname,
		Email:     u.Email,
		Country:   u.Country,
	}
}

func entityUser(e Entity) storage.User {
	return storage.User{
		FirstName: e.FirstName,
		LastName:  e.LastName,
		Nickname:  e.Nickname,
		Email:     e.Email,
		Country:   e.Country,
	}
}
