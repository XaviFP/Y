package publisher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPublisher(t *testing.T) {
	broker := &mockBroker{}
	srv := NewServer(broker)

	s := httptest.NewServer(http.HandlerFunc(srv.subscribe))

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")

	t.Run("success", func(t *testing.T) {
		ch := make(chan Article)
		broker.On("AddSubscriber", "testUserID").Return(ch)

		c, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Y-User-ID": []string{"testUserID"}})
		assert.Nil(t, err)

		defer c.Close()

		expected := Article{Title: "title", Body: "body", Category: "category"}

		// Simluate broker sending an article
		ch <- expected

		var actual Article
		err = c.ReadJSON(&actual)
		assert.Nil(t, err)
		assert.Equal(t, expected, actual)

		broker.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		// No Y-User-ID header to trigger error
		c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.Nil(t, err)

		defer c.Close()

		var actual errorPayload
		err = c.ReadJSON(&actual)
		assert.Nil(t, err)
		assert.Equal(t, "User id not found", actual.Err)
	})
}

type mockBroker struct {
	mock.Mock
}

func (m *mockBroker) AddSubscriber(userID string) chan Article {
	return m.Called(userID).Get(0).(chan Article)
}

func (m *mockBroker) RemoveSubscriber(userID string) {
	m.Called(userID)
}

func (m *mockBroker) Run() {
	m.Called()
}

func (m *mockBroker) Stop() {
	m.Called()
}
