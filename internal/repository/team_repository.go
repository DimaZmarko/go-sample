package repository

import (
	"fmt"
	"time"
	"go-sample/internal/cache"
	"go-sample/internal/models"

	"gorm.io/gorm"
)

type teamRepository struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewTeamRepository(db *gorm.DB, cache *cache.Cache) TeamRepository {
	return &teamRepository{
		db:    db,
		cache: cache,
	}
}

func (r *teamRepository) Create(team *models.Team) error {
	if err := r.db.Create(team).Error; err != nil {
		return err
	}
	// Invalidate cache
	r.cache.Delete("teams_list")
	return nil
}

func (r *teamRepository) Update(team *models.Team) error {
	if err := r.db.Save(team).Error; err != nil {
		return err
	}
	// Invalidate caches
	r.cache.Delete("teams_list")
	r.cache.Delete(fmt.Sprintf("team_%d", team.ID))
	return nil
}

func (r *teamRepository) Delete(id uint) error {
	if err := r.db.Delete(&models.Team{}, id).Error; err != nil {
		return err
	}
	// Invalidate caches
	r.cache.Delete("teams_list")
	r.cache.Delete(fmt.Sprintf("team_%d", id))
	return nil
}

func (r *teamRepository) GetByID(id uint) (*models.Team, error) {
	var team models.Team
	cacheKey := fmt.Sprintf("team_%d", id)

	// Try to get from cache
	err := r.cache.Get(cacheKey, &team)
	if err == nil {
		return &team, nil
	}

	// If not in cache, get from DB
	if err := r.db.Preload("Users").First(&team, id).Error; err != nil {
		return nil, err
	}

	// Store in cache
	r.cache.Set(cacheKey, team, 5*time.Minute)
	return &team, nil
}

func (r *teamRepository) List() ([]models.Team, error) {
	var teams []models.Team
	cacheKey := "teams_list"

	// Try to get from cache
	err := r.cache.Get(cacheKey, &teams)
	if err == nil {
		return teams, nil
	}

	// If not in cache, get from DB
	if err := r.db.Preload("Users").Find(&teams).Error; err != nil {
		return nil, err
	}

	// Store in cache
	r.cache.Set(cacheKey, teams, 5*time.Minute)
	return teams, nil
}

func (r *teamRepository) AddUser(teamID, userID uint) error {
	// First check if team exists
	var team models.Team
	if err := r.db.First(&team, teamID).Error; err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	// Check if user exists
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Add the association
	if err := r.db.Model(&team).Association("Users").Append(&user); err != nil {
		return fmt.Errorf("failed to add user to team: %w", err)
	}

	// Invalidate caches
	r.cache.Delete(fmt.Sprintf("team_%d", teamID))
	r.cache.Delete(fmt.Sprintf("user_%d", userID))
	r.cache.Delete("teams_list")
	r.cache.Delete("users_list")

	return nil
} 