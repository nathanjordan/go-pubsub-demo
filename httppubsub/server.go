package httppubsub

import (
	"encoding/json"
	"github.com/nathanjordan/go-pubsub-demo/internal/httpjson"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/nathanjordan/go-pubsub-demo/internal/pubsub"
	"golang.org/x/net/websocket"
)

const defaultMaxRequestSize = 1024 * 1024 // 1MB

// Server is a broadcast-only pubsub server.
type Server struct {
	mux    *http.ServeMux
	pubsub *pubsub.PubSub
	log    Logger
	maxRequestSize int64
}

// NewServer returns a new Server.
func NewServer(opts ...ServerOption) (*Server, error) {
	mux := http.NewServeMux()
	s := &Server{
		mux:    mux,
		pubsub: pubsub.NewPubSub(),
		log:    nopLogger(),
		maxRequestSize: defaultMaxRequestSize,
	}
	for _, opt := range opts {
		opt(s)
	}
	mux.HandleFunc("/broadcast", s.broadcast)
	wsServer := websocket.Server{
		Handler: s.subscribe,
	}
	mux.Handle("/subscribe", wsServer)
	return s, nil
}

// ServeHTTP is the http.Handler interface for publish and subscribe requests.
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	s.mux.ServeHTTP(rw, req)
}

func (s *Server) subscribe(ws *websocket.Conn) {
	s.log("received new subscription")

	subscriber := newSubscriber(ws)
	subscriber.start()

	handle := s.pubsub.Subscribe(subscriber.sendMessage)
	defer handle.Close()

	subscriber.wait()
}

func (s *Server) broadcast(rw http.ResponseWriter, req *http.Request) {
	s.log("received publish from %s", req.RemoteAddr)

	if req.Method != "POST" {
		httpjson.JSONError(rw, http.StatusMethodNotAllowed, "use POST for publishing")
		return
	}

	// protect DoS from oversized requests
	r := &io.LimitedReader{
		R: req.Body,
		N: s.maxRequestSize+1, // read one more byte to be exact
	}
	body, err := ioutil.ReadAll(r)
	if err != nil {
		httpjson.JSONError(rw, http.StatusBadRequest, "failed to read response body: %v", err)
		return
	}
	if r.N == 0 {
		httpjson.JSONError(rw, http.StatusRequestEntityTooLarge, "request was too large")
		return
	}

	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		httpjson.JSONError(rw, http.StatusBadRequest, "failed to unmarshal message: %v", err)
		return
	}
	if err := s.pubsub.Broadcast(pubsub.NewMessageString(msg.Data)); err != nil {
		httpjson.JSONError(rw, http.StatusInternalServerError, "failed to publish message: %v", err)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

type subscriber struct {
	ws        *websocket.Conn
	messageCh chan *pubsub.Message

	wg        sync.WaitGroup
	closeCh   chan struct{}
	closeOnce sync.Once
}

func newSubscriber(ws *websocket.Conn) *subscriber {
	return &subscriber{
		ws:        ws,
		messageCh: make(chan *pubsub.Message, 1),
		closeCh:   make(chan struct{}),
	}
}

func (s *subscriber) sendMessage(m *pubsub.Message) {
	select {
	case s.messageCh <- m:
	case <-s.closeCh:
	}
}

func (s *subscriber) start() {
	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		s.readLoop()
	}()
	go func() {
		defer s.wg.Done()
		s.writeLoop()
	}()
}

func (s *subscriber) stop() {
	s.closeOnce.Do(func() {
		close(s.closeCh)
	})
	s.ws.Close()
}

func (s *subscriber) wait() {
	s.wg.Wait()
}

func (s *subscriber) readLoop() {
	defer s.stop()

	// we don't care about anything the client writes to us so we
	// will read until connection is terminated or errors out.
	io.Copy(ioutil.Discard, s.ws)
}

func (s *subscriber) writeLoop() {
	defer s.stop()

	// wait for new broadcasts or a close of the connection
	for {
		select {
		case <-s.closeCh:
			return
		case msg := <-s.messageCh:
			m := Message{Data: string(msg.Data())}
			if err := websocket.JSON.Send(s.ws, m); err != nil {
				return
			}
		}
	}
}
