package repository

import (
    "fmt"
    "time"
    "go-sample/internal/cache"
    "go-sample/internal/models"

    "gorm.io/gorm"
)

type userRepository struct {
    db    *gorm.DB
    cache *cache.Cache
}

func NewUserRepository(db *gorm.DB, cache *cache.Cache) UserRepository {
    return &userRepository{
        db:    db,
        cache: cache,
    }
}

func (r *userRepository) Create(user *models.User) error {
    if err := r.db.Create(user).Error; err != nil {
        return err
    }
    // Invalidate cache
    r.cache.Delete("users_list")
    return nil
}

func (r *userRepository) Update(user *models.User) error {
    if err := r.db.Save(user).Error; err != nil {
        return err
    }
    // Invalidate caches
    r.cache.Delete("users_list")
    r.cache.Delete(fmt.Sprintf("user_%d", user.ID))
    return nil
}

func (r *userRepository) Delete(id uint) error {
    if err := r.db.Delete(&models.User{}, id).Error; err != nil {
        return err
    }
    // Invalidate caches
    r.cache.Delete("users_list")
    r.cache.Delete(fmt.Sprintf("user_%d", id))
    return nil
}

func (r *userRepository) GetByID(id uint) (*models.User, error) {
    var user models.User
    cacheKey := fmt.Sprintf("user_%d", id)

    // Try to get from cache
    err := r.cache.Get(cacheKey, &user)
    if err == nil {
        return &user, nil
    }

    // If not in cache, get from DB
    if err := r.db.First(&user, id).Error; err != nil {
        return nil, err
    }

    // Store in cache
    r.cache.Set(cacheKey, user, 5*time.Minute)
    return &user, nil
}

func (r *userRepository) GetWithTeams(id uint) (*models.User, error) {
    var user models.User
    cacheKey := fmt.Sprintf("user_teams_%d", id)

    // Try to get from cache
    err := r.cache.Get(cacheKey, &user)
    if err == nil {
        return &user, nil
    }

    // If not in cache, get from DB with teams
    if err := r.db.Preload("Teams").First(&user, id).Error; err != nil {
        return nil, err
    }

    // Store in cache
    r.cache.Set(cacheKey, user, 5*time.Minute)
    return &user, nil
}

func (r *userRepository) List() ([]models.User, error) {
    var users []models.User
    cacheKey := "users_list"

    // Try to get from cache
    err := r.cache.Get(cacheKey, &users)
    if err == nil {
        return users, nil
    }

    // If not in cache, get from DB
    if err := r.db.Find(&users).Error; err != nil {
        return nil, err
    }

    // Store in cache
    r.cache.Set(cacheKey, users, 5*time.Minute)
    return users, nil
}

func (r *userRepository) UpdateWithTeams(user *models.User, teamIDs []uint) error {
    // Start transaction
    tx := r.db.Begin()
    if tx.Error != nil {
        return tx.Error
    }

    // Update user basic info
    if err := tx.Save(user).Error; err != nil {
        tx.Rollback()
        return err
    }

    // If teamIDs is provided, update team associations
    if teamIDs != nil {
        // Clear existing associations
        if err := tx.Model(user).Association("Teams").Clear(); err != nil {
            tx.Rollback()
            return err
        }

        // Add new team associations
        if len(teamIDs) > 0 {
            var teams []models.Team
            if err := tx.Find(&teams, teamIDs).Error; err != nil {
                tx.Rollback()
                return err
            }

            if err := tx.Model(user).Association("Teams").Replace(&teams); err != nil {
                tx.Rollback()
                return err
            }
        }
    }

    // Commit transaction
    if err := tx.Commit().Error; err != nil {
        return err
    }

    // Invalidate caches
    r.cache.Delete("users_list")
    r.cache.Delete(fmt.Sprintf("user_%d", user.ID))
    r.cache.Delete(fmt.Sprintf("user_teams_%d", user.ID))
    return nil
} 