package app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"go-sample/internal/cache"
	"go-sample/internal/config"
	"go-sample/internal/handlers"
	"go-sample/internal/models"
	"go-sample/internal/repository"
	"go-sample/internal/router"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	config *config.Config
	server *http.Server
}

func NewApp() (*App, error) {
	// Load configuration
	cfg := config.NewConfig()
	log.Printf("Starting application...")

	// Initialize database with retry logic
	db, err := initDatabase(cfg)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed: %w", err)
	}

	// Initialize cache
	log.Printf("Initializing cache...")
	cacheService := cache.NewCache(cfg.RedisHost, cfg.RedisPort)

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, cacheService)
	teamRepo := repository.NewTeamRepository(db, cacheService)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo)
	teamHandler := handlers.NewTeamHandler(teamRepo)
	importHandler := handlers.NewImportHandler(userRepo, teamRepo)

	// Setup router
	r := router.SetupRouter(userHandler, teamHandler, importHandler)

	// Configure server
	server := &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &App{
		config: cfg,
		server: server,
	}, nil
}

func (a *App) Start() error {
	log.Printf("Server starting on http://0.0.0.0:8080")
	return a.server.ListenAndServe()
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < 5; i++ {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=5432 sslmode=disable",
			cfg.PostgresHost, cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)
		
		log.Printf("Attempting to connect to database... (attempt %d)", i+1)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database: %v", err)
		time.Sleep(time.Second * 5)
	}

	if err != nil {
		return nil, fmt.Errorf("could not connect to database after 5 attempts: %w", err)
	}
	log.Printf("Successfully connected to database")

	// Auto migrate the schema
	log.Printf("Running database migrations...")
	if err := db.AutoMigrate(&models.User{}, &models.Team{}, &models.TeamUser{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Printf("Database migrations completed")

	return db, nil
} 