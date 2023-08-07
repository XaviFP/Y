package aggregator

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestServer(t *testing.T) {
	done := make(chan struct{})
	repo := &mockArticleRepository{done: done}

	expected := Article{Title: "title", Body: "body"}
	repo.On("store", expected).Return(nil)

	srv := NewServer(repo)
	s := httptest.NewServer(http.HandlerFunc(srv.publish))

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	c, _, err := websocket.DefaultDialer.Dial(wsURL, http.Header{"Y-User-ID": []string{"testUserID"}})
	assert.Nil(t, err)
	defer c.Close()

	err = c.WriteJSON(expected)
	assert.Nil(t, err)

	<-done

	repo.AssertExpectations(t)
}

type mockArticleRepository struct {
	mock.Mock
	done chan struct{}
}

func (m *mockArticleRepository) store(a Article) error {
	args := m.Called(a)
	m.done <- struct{}{}

	return args.Error(0)
}
