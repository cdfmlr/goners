// Package wsforwarder helps to forward messages (from chan, etc) to WebSocket clients.
//
// - from https://github.com/cdfmlr/live2ddriver.
// - included in https://github.com/cdfmlr/muvtuber under the MIT License.
//
// Update: log -> slog
package wsforwarder

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/exp/slog"
	"golang.org/x/net/websocket"
)

// chan buffer size
const BufferSize = 8

type Forwarder interface {
	ForwardMessageTo(ws *websocket.Conn)
	ForwardMessageFrom(msgCh <-chan []byte)
}

// messageForwarder forwards messages to connected clients, that are, Live2DViews.
type messageForwarder struct {
	msgChans []chan []byte
	mu       sync.RWMutex // to protect msgChans
}

func NewMessageForwarder() Forwarder {
	return &messageForwarder{
		msgChans: []chan []byte{},
	}
}

// ForwardMessageTo the WebSocket connection.
//
// Use SendMessage to send messages.
//
// Block until the websocket connection is closed.
func (f *messageForwarder) ForwardMessageTo(ws *websocket.Conn) {
	ch := make(chan []byte, BufferSize)

	// add

	f.mu.Lock()
	f.msgChans = append(f.msgChans, ch)
	f.mu.Unlock()

	logger.Info("Start ForwardMessageTo",
		"ws.RemoteAddr()", ws.RemoteAddr(),
		"chan", ch)

	// forward

	forwardMessage(ch, ws) // 阻塞

	// clean up

	close(ch)

	f.mu.Lock()
	for i, c := range f.msgChans {
		if c == ch {
			f.msgChans = append(f.msgChans[:i], f.msgChans[i+1:]...)
			break
		}
	}
	f.mu.Unlock()

	logger.Info("Stop ForwardMessageTo: %s by chan %v.", ws.RemoteAddr(), ch)
}

// SendMessage to WebSocket clients.
//
// Block until message is sent to all clients.
func (f *messageForwarder) SendMessage(msg []byte) {
	slog.Info("SendMessage", "msg", string(msg))

	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, ch := range f.msgChans {
		if ch != nil {
			ch <- msg
		}
	}
}

// ForwardMessageFrom the message channel.
//
// Block until the message channel is closed.
func (f *messageForwarder) ForwardMessageFrom(msgCh <-chan []byte) {
	for msg := range msgCh {
		f.SendMessage(msg)
	}
}

// forwardMessage forwards messages from the message channel to the websocket
// connection.
//
// The message channel is expected to receive JSON strings (bytes):
//
//	`{"motion": "shake"}`
//	`{"expression": "f03"}`
func forwardMessage(msgCh <-chan []byte, ws *websocket.Conn) {
	for msg := range msgCh {
		slog.Info(fmt.Sprintf("fwd msg: %s -> %s (chan %v).", string(msg), ws.RemoteAddr(), msgCh))
		_, err := ws.Write(msg)
		if err != nil {
			slog.Info(fmt.Sprintf("fwd msg to %s (chan %v) error: %s.", ws.RemoteAddr(), msgCh, err))
			break
		}
	}
	_ = ws.Close()
}

// region useful ForwardMessageFrom* methods

// ForwardMessageFromStdin read Live2DRequest from stdin and send it to MessageForwarder.
//
// Block until EOF (that is, never).
func (f *messageForwarder) ForwardMessageFromStdin() {
	slog.Info("(in) Forwarding messages from stdin to WebSocket clients...\n")
	time.Sleep(time.Millisecond * 200) // 太快了日志和输入提示交错不好看
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Printf("Enter a message to send: ")
	for {
		scanner.Scan()
		f.SendMessage(scanner.Bytes())
		time.Sleep(time.Millisecond * 200) // 太快了日志和输入提示交错不好看
		fmt.Printf("Enter a message to send: ")
	}
}

// endregion useful ForwardMessageFrom* methods
