package aggregator

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/tilinna/clock"
)

func TestRepository(t *testing.T) {
	db := newTestDB(t)

	loc, err := time.LoadLocation("Etc/UTC")
	assert.Nil(t, err)

	clockMock := clock.NewMock(time.Date(2020, 1, 1, 12, 0, 0, 0, loc))
	repo := NewArticleRepository(db, clockMock)

	expected := Article{
		Title:    "test-title",
		Body:     "test-body",
		Category: "testcategory",
	}

	err = repo.store(expected)
	assert.Nil(t, err)

	var actual Article
	row := db.QueryRow("SELECT title, body, category, published_at FROM articles")
	err = row.Scan(&actual.Title, &actual.Body, &actual.Category, &actual.PublishedAt)
	assert.Nil(t, err)

	expected.PublishedAt = clockMock.Now()
	assert.Equal(t, expected, actual)
}

func newTestDB(t *testing.T) *sql.DB {
	dbConfig := DbConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "y",
		Password: "y",
		Name:     "y",
		SSLMode:  "disable",
	}
	db, err := sql.Open("postgres", dbConfig.String())
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("DELETE FROM articles")
	t.Cleanup(func() {
		defer db.Close()

		_, err = db.Exec("DELETE FROM articles")
		assert.Nil(t, err)
	})

	return db
}
