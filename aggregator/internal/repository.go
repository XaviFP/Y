package aggregator

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tilinna/clock"
)

type DbConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (dbC DbConfig) String() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbC.Host,
		dbC.Port,
		dbC.User,
		dbC.Password,
		dbC.Name,
		dbC.SSLMode,
	)
}

type Article struct {
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Category    string    `json:"category"`
	PublishedAt time.Time `json:"published_at"`
}

type ArticleRepository interface {
	store(a Article) error
}

type articleRepository struct {
	db    *sql.DB
	clock clock.Clock
}

func NewArticleRepository(db *sql.DB, c clock.Clock) ArticleRepository {
	return &articleRepository{
		db:    db,
		clock: c,
	}
}

func (r *articleRepository) store(a Article) error {
	_, err := r.db.Exec(`
		INSERT INTO articles (
				title,
				body,
				category,
				published_at
		) VALUES ($1, $2, $3, $4)`,
		a.Title,
		a.Body,
		a.Category,
		r.clock.Now().UTC(),
	)
	if err != nil {
		return err
	}
	return nil
}
