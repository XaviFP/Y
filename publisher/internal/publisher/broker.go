package publisher

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/lib/pq"
)

type Article struct {
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Category    string    `json:"category"`
	PublishedAt time.Time `json:"published_at"`
}

type Broker interface {
	AddSubscriber(userID string) chan Article
	RemoveSubscriber(userID string)
	Run()
	Stop()
}

type subscriberSession struct {
	channel chan Article
	balance int
}

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

type broker struct {
	subscribers    map[string]*subscriberSession
	mut            sync.Mutex
	done           chan struct{}
	pgListener     *pq.Listener
	initialBalance int
}

func NewBroker(dbConfig DbConfig, initialBalance int) Broker {
	l := newPostgresListener(dbConfig)
	return &broker{
		subscribers:    make(map[string]*subscriberSession),
		pgListener:     l,
		initialBalance: initialBalance,
	}
}

func (b *broker) AddSubscriber(userID string) chan Article {
	b.mut.Lock()
	defer b.mut.Unlock()

	if _, ok := b.subscribers[userID]; ok {
		return nil
	}

	ch := make(chan Article, 10)
	b.subscribers[userID] = &subscriberSession{
		channel: ch,
		balance: b.initialBalance,
	}

	return ch
}

func (b *broker) RemoveSubscriber(userID string) {
	b.mut.Lock()
	defer b.mut.Unlock()

	s, ok := b.subscribers[userID]
	if !ok {
		return
	}

	close(s.channel)

	delete(b.subscribers, userID)
	return
}

func (b *broker) Run() {
	go func() {
		for {
			select {
			case <-b.done:
				log.Println("broker shutdown")
				return
			case n := <-b.pgListener.Notify:
				var article Article
				err := json.Unmarshal([]byte(n.Extra), &article)
				if err != nil {
					log.Println("Error processing JSON: ", err)
					return
				}
				article.PublishedAt = article.PublishedAt.UTC()

				b.mut.Lock()
				for userID := range b.subscribers {
					session := b.subscribers[userID]
					if session.balance > 0 {
						session.channel <- article
						session.balance--
					} else {
						session.channel <- Article{
							Title:       article.Title,
							Body:        "Top up your account to read the full content",
							Category:    article.Category,
							PublishedAt: article.PublishedAt,
						}
					}
				}
				b.mut.Unlock()
			}
		}
	}()
}

func (b *broker) Stop() {
	b.done <- struct{}{}
}

func newPostgresListener(dbConfig DbConfig) *pq.Listener {

	_, err := sql.Open("postgres", dbConfig.String())
	if err != nil {
		panic(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(dbConfig.String(), 10*time.Second, time.Minute, reportProblem)
	err = listener.Listen("new_articles")
	if err != nil {
		panic(err)
	}

	return listener
}
