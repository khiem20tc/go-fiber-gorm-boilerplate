package dto

import (
	"fiber-gateway/model"
	"time"

	"github.com/dranikpg/dto-mapper"
)

type userListDTO []struct {
	ID        string    `json:"id" validate:"required"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

func DTOUserList(users *[]model.User) userListDTO {

	var _userListDTO userListDTO
	dto.Map(&_userListDTO, users)

	return _userListDTO
}
