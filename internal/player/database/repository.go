package database

import (
	"github.com/xBlaz3kx/event-processing-challenge/internal/casino"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Configuration struct {
	// DSN is the data source name.
	DSN string `json:"dsn"`
}

type Repository interface {
	// GetPlayerDetails returns player details from the database.
	GetPlayerDetails(ctx context.Context, playerID int) (*casino.Player, error)
}

type PostgresRepository struct {
	gormDriver *gorm.DB
	logger     *zap.Logger
}

func NewPostgresPlayerRepository(logger *zap.Logger, cfg Configuration) *PostgresRepository {
	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		logger.Panic("failed to connect to the database", zap.Error(err))
	}

	return &PostgresRepository{
		gormDriver: db,
		logger:     logger.Named("postgres-player-repository"),
	}
}

// GetPlayerDetails returns player details from the database.
func (r *PostgresRepository) GetPlayerDetails(ctx context.Context, playerID int) (*casino.Player, error) {
	var player casino.Player
	result := r.gormDriver.WithContext(ctx).First(&player, "id = ?", playerID)
	if result.Error != nil {
		return nil, result.Error
	}

	return &player, nil
}

// Close closes the database connection.
func (r *PostgresRepository) Close() error {
	db, err := r.gormDriver.DB()
	if err != nil {
		return err
	}

	return db.Close()
}

func (r *PostgresRepository) Pass() bool {
	db, err := r.gormDriver.DB()
	if err != nil {
		return false
	}

	return db.Ping() == nil
}

func (r *PostgresRepository) Name() string {
	return "postgres-player-repository"
}
