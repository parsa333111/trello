package websocket_utils

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/skye-tan/trello/backend/database"
	"github.com/skye-tan/trello/backend/middlewares/monitoring"
)

type WebsocketHub struct {
	clients    map[uint]*websocket.Conn
	Register   chan *WebsocketConn
	Unregister chan uint
	Broadcast  chan *WebsocketBroadcast
	GetStatus  chan *WebsocketGetStatus
}

var Hub = WebsocketHub{
	clients:    make(map[uint]*websocket.Conn),
	Register:   make(chan *WebsocketConn),
	Unregister: make(chan uint),
	Broadcast:  make(chan *WebsocketBroadcast),
	GetStatus:  make(chan *WebsocketGetStatus),
}

func try_broadcast_offline_message(user_id uint) {
	associated_users, err := database.GetAssociatedUsersWithUser(user_id)

	if err != nil {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Unsuccessful).Inc()
	} else {
		monitoring.Statistics.Queries.WithLabelValues(monitoring.Successful).Inc()

		Hub.Broadcast <- &WebsocketBroadcast{
			TargetUserIDs: associated_users,
			Body: &WebsocketBody{
				Group:   MemeberGroup,
				Type:    UpdateType,
				Message: "A user is offline now.",
			},
		}
	}
}

func (hub *WebsocketHub) unregister(user_id uint) {
	if conn, ok := hub.clients[user_id]; ok {
		conn.Close()
		delete(hub.clients, user_id)
	}
	go try_broadcast_offline_message(user_id)
}

func (hub *WebsocketHub) broadcast(request *WebsocketBroadcast) {
	for _, user_id := range request.TargetUserIDs {
		if conn, ok := hub.clients[user_id]; ok {
			if err := conn.WriteJSON(request.Body); err != nil {
				log.Println("Error:", err)
				hub.unregister(user_id)
			}
		}
	}
}

func (hub *WebsocketHub) getStatus(target *WebsocketGetStatus) {
	if _, ok := hub.clients[target.TargetUserID]; ok {
		target.Responder <- Online
	} else {
		target.Responder <- Offline
	}
}

func (hub *WebsocketHub) Run() {
	for {
		select {
		case websocket_conn := <-hub.Register:
			hub.clients[websocket_conn.UserID] = websocket_conn.Conn
		case user_id := <-hub.Unregister:
			hub.unregister(user_id)
		case request := <-hub.Broadcast:
			hub.broadcast(request)
		case request := <-hub.GetStatus:
			hub.getStatus(request)
		}
	}
}

func HandleClientWebsocket(conn *websocket.Conn, user_id uint) {
	defer func() {
		Hub.Unregister <- user_id
	}()
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
