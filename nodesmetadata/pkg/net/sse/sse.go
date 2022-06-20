package sse

import (
	"io"

	"github.com/gin-gonic/gin"
)

type ClientChan chan any

type SSE struct {
	EventsToSend        ClientChan
	CloseAllClientsChan chan bool

	newClientsToHandle    chan ClientChan
	closedClientsToHandle chan ClientChan
	clientsChan           map[ClientChan]bool
}

func New() *SSE {
	sse := &SSE{
		EventsToSend:        make(ClientChan),
		CloseAllClientsChan: make(chan bool),

		newClientsToHandle:    make(chan ClientChan),
		closedClientsToHandle: make(chan ClientChan),
		clientsChan:           make(map[ClientChan]bool),
	}

	go sse.process()

	return sse
}

func (sse *SSE) Handler() func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		clientChan := make(ClientChan)

		sse.newClientsToHandle <- clientChan

		defer func() {
			sse.closedClientsToHandle <- clientChan
		}()

		// handle connection closed by client
		go func() {
			<-c.Done()
			sse.closedClientsToHandle <- clientChan
		}()

		c.Stream(func(w io.Writer) bool {
			if event, ok := <-clientChan; ok {
				c.SSEvent("event", event)
				return true
			}
			return false
		})
	}
}

func (sse *SSE) closeAllClients() {
	for clientChan := range sse.clientsChan {
		close(clientChan)
	}
	sse.EventsToSend = make(chan any)
	sse.CloseAllClientsChan = make(chan bool)
	sse.newClientsToHandle = make(chan ClientChan)
	sse.closedClientsToHandle = make(chan ClientChan)
	sse.clientsChan = make(map[ClientChan]bool)
}

func (sse *SSE) process() {
	for {
		select {
		case <-sse.CloseAllClientsChan:
			sse.closeAllClients()
		default:
			select {
			case <-sse.CloseAllClientsChan:
				sse.closeAllClients()
			case clientChan := <-sse.newClientsToHandle:
				sse.clientsChan[clientChan] = true
			case clientChan := <-sse.closedClientsToHandle:
				delete(sse.clientsChan, clientChan)
			case eventToSend := <-sse.EventsToSend:
				for clientChan := range sse.clientsChan {
					clientChan <- eventToSend
				}
			}
		}
	}
}
