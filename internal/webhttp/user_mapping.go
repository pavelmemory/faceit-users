package webhttp

import (
	"github.com/pavelmemory/faceit-users/internal/user"
)

type CreateUserReq struct {
	UserBase
	Password string `json:"password,omitempty"`
}

type UpdateUserReq struct {
	UserBase
}

type GetUserResp struct {
	UserBase
}

type UserBase struct {
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Nickname  string `json:"nickname,omitempty"`
	Email     string `json:"email,omitempty"`
	Country   string `json:"country,omitempty"`
}

type Mapper struct{}

func (m Mapper) createUserReq2Entity(req CreateUserReq) user.Entity {
	entity := m.userBase2Entity(req.UserBase)
	entity.Password = req.Password
	return entity
}

func (m Mapper) updateUserReq2Entity(req UpdateUserReq) user.Entity {
	return m.userBase2Entity(req.UserBase)
}

func (Mapper) userBase2Entity(ub UserBase) user.Entity {
	return user.Entity{
		FirstName: ub.FirstName,
		LastName:  ub.LastName,
		Nickname:  ub.Nickname,
		Email:     ub.Email,
		Country:   ub.Country,
	}
}

func (Mapper) entity2GetUserResp(entity user.Entity) GetUserResp {
	return GetUserResp{
		UserBase{
			FirstName: entity.FirstName,
			LastName:  entity.LastName,
			Nickname:  entity.Nickname,
			Email:     entity.Email,
			Country:   entity.Country,
		},
	}
}
