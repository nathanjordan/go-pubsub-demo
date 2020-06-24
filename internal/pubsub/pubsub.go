package pubsub

import (
	"container/list"
	"sync"

	"golang.org/x/sync/errgroup"
)

// PubSub is a simple broadcast-only pubsub implementation.
type PubSub struct {
	mu          sync.Mutex
	subscribers *list.List
}

// NewPubSub creates a new PubSub.
func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: list.New(),
	}
}

// Broadcast broadcasts a message to all registered subscribers. It blocks
// until all subscriber callbacks have returned.
//
// TODO(nathan): use a context.Context here as it blocks.
func (p *PubSub) Broadcast(m *Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// done to avoid accessing loop variables within a goroutine
	// which creates a race
	var grp errgroup.Group
	fn := func(h *subscriberHandle) {
		grp.Go(func() error {
			h.cb(m)
			return nil
		})
	}

	for e := p.subscribers.Front(); e != nil; e = e.Next() {
		fn(e.Value.(*subscriberHandle))
	}
	grp.Wait()
	return nil
}

// Subscribe registers a callback that will receive messages that are
// sent by Broadcast. The returned handle can be used to unsubscribe from
// further broadcasts. The provided callback must not block.
func (p *PubSub) Subscribe(cb func(*Message)) SubscriberHandle {
	p.mu.Lock()
	defer p.mu.Unlock()

	c := &subscriberHandle{
		ps: p,
		cb: cb,
	}
	p.subscribers.PushBack(c)
	return c
}

func (p *PubSub) unsubscribe(c *subscriberHandle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for e := p.subscribers.Front(); e != nil; e = e.Next() {
		if e.Value == c {
			p.subscribers.Remove(e)
			return
		}
	}
	// if this happens the handle was closed twice
}

type subscriberHandle struct {
	ps *PubSub
	cb func(*Message)
}

func (c *subscriberHandle) Close() error {
	c.ps.unsubscribe(c)
	return nil
}

// SubscriberHandle is used to stop the subscriber from
// receiving further updates.
type SubscriberHandle interface {
	Close() error
}

// Message is a message that was broadcast.
type Message struct {
	data []byte
}

// Data returns the data contained in the Message.
func (m *Message) Data() []byte {
	return m.data
}

// NewMessage creates a new Message from a byte slice.
func NewMessage(data []byte) *Message {
	return &Message{
		data: data,
	}
}

// NewMessageString creates a new Message from a string.
func NewMessageString(data string) *Message {
	return &Message{
		data: []byte(data),
	}
}
