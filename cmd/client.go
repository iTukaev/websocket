package main

import (
	"go.uber.org/zap"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// BytesSender is interface of data for clients
type BytesSender interface {
	Bytes() ([]byte, error)
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	// the websocket connection.
	conn *websocket.Conn
	// channel of outbound data.
	send chan []byte
	// data for spending to websocket channel.
	senders []BytesSender
}

// writePump pumps messages from the hub to the websocket connection.
// A goroutine running writePump is started for each connection.
func (c *Client) writePump(logger *zap.SugaredLogger) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			logger.Warnf("websocket %s connection closing error: %v", c.conn.RemoteAddr().String(), err)
		}
	}()

	// first massage to websocket client
	for i := range c.senders {
		b, err := c.senders[i].Bytes()
		if err != nil {
			logger.Warn(err)
		} else {
			if err = c.conn.WriteMessage(websocket.TextMessage, b); err != nil {
				logger.Warnf("websocket %s message writing error: %v", c.conn.RemoteAddr().String(), err)
			}
		}
	}

	// regular massages to websocket client
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				logger.Warnf("NextWriter creating error: %v", err)
				return
			}

			if _, err = w.Write(message); err != nil {
				logger.Warnf("websocket %s message writing error: %v", c.conn.RemoteAddr().String(), err)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
// senders include links on data sets for messaging
func serveWs(w http.ResponseWriter, r *http.Request, logger *zap.SugaredLogger, hub *Hub, senders ...BytesSender) {
	logger.Info(r.RemoteAddr)
	//todo: сделать проверку подлинности подключения (узнать как делают)
	upgrade.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		logger.Warnf("websocket connection creating error: %v", err)
		return
	}

	client := &Client{
		conn:    conn,
		send:    make(chan []byte),
		senders: senders,
	}

	hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump(logger)
}
