package publisher

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBroker(t *testing.T) {
	h := newTestHarness(t)
	broker := NewBroker(h.dbConfig, 1)
	broker.Run()

	ch := broker.AddSubscriber("testUserID")

	expected := Article{
		Title:       "title",
		Body:        "body",
		Category:    "category",
		PublishedAt: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	_, err := h.db.Exec(`
		INSERT INTO articles (
				title,
				body,
				category,
				published_at
		) VALUES ($1, $2, $3, $4)`,
		expected.Title,
		expected.Body,
		expected.Category,
		expected.PublishedAt,
	)
	assert.Nil(t, err)

	actual := <-ch

	assert.Equal(t, expected, actual)

	// After reading the previous article,
	// user balance is now 0, so the next article should be paywalled
	fullArticle := Article{
		Title:       "Full Title",
		Body:        "Full Body",
		Category:    "Full Category",
		PublishedAt: time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	_, err = h.db.Exec(`
		INSERT INTO articles (
				title,
				body,
				category,
				published_at
		) VALUES ($1, $2, $3, $4)`,
		fullArticle.Title,
		fullArticle.Body,
		fullArticle.Category,
		fullArticle.PublishedAt,
	)
	assert.Nil(t, err)

	actual = <-ch

	noFundsArticle := Article{
		Title:       fullArticle.Title,
		Body:        "Top up your account to read the full content",
		Category:    fullArticle.Category,
		PublishedAt: fullArticle.PublishedAt,
	}

	assert.Equal(t, noFundsArticle, actual)
}

type testHarness struct {
	db       *sql.DB
	dbConfig DbConfig
}

func newTestHarness(t *testing.T) *testHarness {
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

	t.Cleanup(func() {
		defer db.Close()

		_, err = db.Exec("DELETE FROM articles")
		assert.Nil(t, err)
	})

	return &testHarness{
		db:       db,
		dbConfig: dbConfig,
	}
}
