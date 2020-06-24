package httppubsub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

// TestServerSuccess tests normal operation of the server with publishes/subscriptions and unsubscriptions.
func TestServerSuccess(t *testing.T) {
	runServerTest(t, func(args serverTestArgs) {
		msg := &Message{
			Data: "Hello!",
		}
		sub1 := args.newSubscriber(t)
		defer sub1.Close()

		sub2 := args.newSubscriber(t)
		defer sub2.Close()

		args.broadcast(t, msg)

		msg1, err := sub1.Wait()
		require.NoError(t, err)
		require.Equal(t, msg, msg1)

		msg2, err := sub2.Wait()
		require.NoError(t, err)
		require.Equal(t, msg, msg2)

		// stop the first publisher
		sub1.Close()

		args.broadcast(t, msg)

		_, err = sub1.Wait()
		require.Error(t, err)

		msg2 = nil
		msg2, err = sub2.Wait()
		require.NoError(t, err)
		require.Equal(t, msg, msg2)
	})
}

func TestProduceWrongMethod(t *testing.T) {
	runServerTest(t, func(args serverTestArgs) {
		broadcastURL := args.server.URL + "/broadcast"
		resp, err := args.client.Get(broadcastURL)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	})
}

func TestProduceBadRequest(t *testing.T) {
	runServerTest(t, func(args serverTestArgs) {
		broadcastURL := args.server.URL + "/broadcast"
		data := "<notJson}"
		resp, err := args.client.Post(broadcastURL, "application/json" /*contentType*/, bytes.NewBufferString(data))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func runServerTest(t *testing.T, testFn func(args serverTestArgs)) {
	ps, err := NewServer()
	require.NoError(t, err)

	server := httptest.NewServer(ps)
	defer server.Close()
	client := server.Client()

	wsURL := fmt.Sprintf("ws://%s/subscribe", server.Listener.Addr())

	testFn(serverTestArgs{
		server: server,
		client: client,
		wsURL:  wsURL,
	})
}

type serverTestArgs struct {
	server *httptest.Server
	client *http.Client
	wsURL  string
}

func (args serverTestArgs) broadcast(t *testing.T, msg *Message) {
	data, err := json.Marshal(msg)
	require.NoError(t, err)
	resp, err := args.client.Post(args.server.URL+"/broadcast", "application/json" /*contentType*/, bytes.NewBuffer(data))
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func (args serverTestArgs) newSubscriber(t *testing.T) *clientSubscriber {
	ws, err := websocket.Dial(args.wsURL, "" /*protocol*/, args.server.URL /*origin*/)
	require.NoError(t, err)

	return newClientSubscriber(ws)
}

type clientSubscriber struct {
	ws *websocket.Conn
}

func newClientSubscriber(ws *websocket.Conn) *clientSubscriber {
	return &clientSubscriber{
		ws: ws,
	}
}

func (s *clientSubscriber) Close() {
	s.ws.Close()
}

func (s *clientSubscriber) Wait() (*Message, error) {
	var received Message
	if err := websocket.JSON.Receive(s.ws, &received); err != nil {
		return nil, err
	}
	return &received, nil
}
