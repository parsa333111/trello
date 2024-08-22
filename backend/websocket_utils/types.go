package websocket_utils

import (
	"github.com/gorilla/websocket"
)

const (
	TaskGroup      = "task"
	SubtaskGroup   = "subtask"
	CommentGroup   = "comment"
	MemeberGroup   = "memeber"
	WorkspaceGroup = "workspace"

	UpdateType = "update"
	WatchType  = "watch"
)

const (
	Online  = "online"
	Offline = "offline"
)

type WebsocketConn struct {
	UserID uint
	Conn   *websocket.Conn
}

type WebsocketBody struct {
	Group   string `json:"group"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type WebsocketBroadcast struct {
	TargetUserIDs []uint
	Body          *WebsocketBody
}

type WebsocketGetStatus struct {
	TargetUserID uint
	Responder    chan string
}
