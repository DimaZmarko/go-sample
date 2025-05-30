package repository

import (
	"go-sample/internal/models"
)

type UserRepository interface {
	Create(user *models.User) error
	Update(user *models.User) error
	UpdateWithTeams(user *models.User, teamIDs []uint) error
	Delete(id uint) error
	GetByID(id uint) (*models.User, error)
	List() ([]models.User, error)
	GetWithTeams(id uint) (*models.User, error)
}

type TeamRepository interface {
	Create(team *models.Team) error
	Update(team *models.Team) error
	Delete(id uint) error
	GetByID(id uint) (*models.Team, error)
	List() ([]models.Team, error)
	AddUser(teamID, userID uint) error
}
