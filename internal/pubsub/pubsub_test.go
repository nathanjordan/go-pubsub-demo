package pubsub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPubSubSimple(t *testing.T) {
	ps := NewPubSub()
	msg := NewMessageString("test123")

	var actual1, actual2 *Message
	handle1 := ps.Subscribe(func(m *Message) {
		actual1 = m
	})
	handle2 := ps.Subscribe(func(m *Message) {
		actual2 = m
	})

	// test both subscribers receive update
	assert.Equal(t, 2, ps.subscribers.Len())
	ps.Broadcast(msg)
	assert.Equal(t, msg, actual1)
	assert.Equal(t, msg, actual2)

	// remove a subscriber and assert that it no longer
	// receives an update
	actual1, actual2 = nil, nil
	handle1.Close()
	assert.Equal(t, 1, ps.subscribers.Len())
	assert.Nil(t, actual1)
	assert.Nil(t, actual2)

	ps.Broadcast(msg)
	assert.Nil(t, actual1)
	assert.Equal(t, msg, actual2)

	// ensure all subscribers removed
	handle1.Close()
	handle2.Close()
	assert.Equal(t, 0, ps.subscribers.Len())
}
