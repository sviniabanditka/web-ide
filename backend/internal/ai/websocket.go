package ai

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type WSClient struct {
	hub      *WSHub
	conn     *websocket.Conn
	send     chan []byte
	projects map[string]bool
}

type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan []byte
	register   chan *WSClient
	unregister chan *WSClient
	sync.RWMutex
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.clients[client] = true
			h.Unlock()
		case client := <-h.unregister:
			h.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.Unlock()
		case message := <-h.broadcast:
			h.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.RUnlock()
		}
	}
}

func (h *WSHub) SendToProject(projectID string, message []byte) {
	h.RLock()
	defer h.RUnlock()
	count := 0
	for client := range h.clients {
		if client.projects[projectID] {
			count++
			select {
			case client.send <- message:
			default:
			}
		}
	}
	log.Printf("WebSocket: sent to %d clients for project %s", count, projectID)
}

var AIHub = NewWSHub()

func init() {
	go AIHub.Run()
}

func RegisterWebSocketRoutes(app *fiber.App) {
	app.Use("/ws/ai", func(c *fiber.Ctx) error {
		if !websocket.IsWebSocketUpgrade(c) {
			return fiber.ErrUpgradeRequired
		}
		return nil
	})

	app.Get("/ws/ai", websocket.New(func(c *websocket.Conn) {
		client := &WSClient{
			hub:      AIHub,
			conn:     c,
			send:     make(chan []byte, 256),
			projects: make(map[string]bool),
		}

		client.hub.register <- client

		defer func() {
			client.hub.unregister <- client
			c.Close()
		}()

		go client.writePump()
		go client.readPump()
	}))
}

func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		if msg.Type == "subscribe" {
			if projectID, ok := msg.Payload.(string); ok {
				c.projects[projectID] = true
			}
		}
	}
}

func (c *WSClient) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send
		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		c.conn.WriteMessage(websocket.TextMessage, message)
	}
}

type JobUpdatePayload struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Result any    `json:"result,omitempty"`
}

func BroadcastJobUpdate(projectID, jobID, status, errorText string, result any) {
	log.Printf("BroadcastJobUpdate: projectID=%s, jobID=%s, status=%s", projectID, jobID, status)
	payload := JobUpdatePayload{
		JobID:  jobID,
		Status: status,
		Error:  errorText,
		Result: result,
	}
	msg := WSMessage{
		Type:    "job_update",
		Payload: payload,
	}
	data, _ := json.Marshal(msg)
	AIHub.SendToProject(projectID, data)
}

type ChangeSetPayload struct {
	ChangeSetID string `json:"changeset_id"`
	Title       string `json:"title"`
	Status      string `json:"status"`
	Summary     string `json:"summary,omitempty"`
}

func BroadcastChangeSetCreated(projectID, csID, title, status, summary string) {
	payload := ChangeSetPayload{
		ChangeSetID: csID,
		Title:       title,
		Status:      status,
		Summary:     summary,
	}
	msg := WSMessage{
		Type:    "changeset_created",
		Payload: payload,
	}
	data, _ := json.Marshal(msg)
	AIHub.SendToProject(projectID, data)
}
