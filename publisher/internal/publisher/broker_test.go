package publisher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBroker_AddRemove(t *testing.T) {
	b := &broker{
		subscribers: make(map[string]*subscriberSession),
	}

	ch := b.AddSubscriber("test")
	session, ok := b.subscribers["test"]
	assert.True(t, ok)
	assert.Equal(t, subscriberSession{channel: ch, balance: 0}, *session)

	b.RemoveSubscriber("test")
	_, ok = b.subscribers["test"]
	assert.False(t, ok)
}
