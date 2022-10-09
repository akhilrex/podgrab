package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type EnqueuePayload struct {
	ItemIds   []string `json:"itemIds"`
	PodcastId string   `json:"podcastId"`
	TagIds    []string `json:"tagIds"`
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	activePlayers  = make(map[*websocket.Conn]string)
	allConnections = make(map[*websocket.Conn]string)
)

var broadcast = make(chan Message) // broadcast channel

type Message struct {
	Identifier  string          `json:"identifier"`
	MessageType string          `json:"messageType"`
	Payload     string          `json:"payload"`
	Connection  *websocket.Conn `json:"-"`
}

func Wshandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to set websocket upgrade: %+v", err))
		return
	}
	defer conn.Close()
	for {
		var mess Message
		err := conn.ReadJSON(&mess)
		if err != nil {
			//	fmt.Println("Socket Error")
			// fmt.Println(err.Error())
			isPlayer := activePlayers[conn] != ""
			if isPlayer {
				delete(activePlayers, conn)
				broadcast <- Message{
					MessageType: "PlayerRemoved",
					Identifier:  mess.Identifier,
				}
			}
			delete(allConnections, conn)
			break
		}
		mess.Connection = conn
		allConnections[conn] = mess.Identifier
		broadcast <- mess
		//	conn.WriteJSON(mess)
	}
}

func HandleWebsocketMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// fmt.Println(msg)

		switch msg.MessageType {
		case "RegisterPlayer":
			activePlayers[msg.Connection] = msg.Identifier
			for connection := range allConnections {
				connection.WriteJSON(Message{
					Identifier:  msg.Identifier,
					MessageType: "PlayerExists",
				})
			}
			fmt.Println("Player Registered")
		case "PlayerRemoved":
			for connection := range allConnections {
				connection.WriteJSON(Message{
					Identifier:  msg.Identifier,
					MessageType: "NoPlayer",
				})
			}
			fmt.Println("Player Registered")
		case "Enqueue":
			var payload EnqueuePayload
			fmt.Println(msg.Payload)
			err := json.Unmarshal([]byte(msg.Payload), &payload)
			if err == nil {
				items := getItemsToPlay(payload.ItemIds, payload.PodcastId, payload.TagIds)
				var player *websocket.Conn
				for connection, id := range activePlayers {
					if msg.Identifier == id {
						player = connection
						break
					}
				}
				if player != nil {
					payloadStr, err := json.Marshal(items)
					if err == nil {
						player.WriteJSON(Message{
							Identifier:  msg.Identifier,
							MessageType: "Enqueue",
							Payload:     string(payloadStr),
						})
					}
				}
			} else {
				fmt.Println(err.Error())
			}
		case "Register":
			var player *websocket.Conn
			for connection, id := range activePlayers {
				if msg.Identifier == id {
					player = connection
					break
				}
			}

			if player == nil {
				fmt.Println("Player Not Exists")
				msg.Connection.WriteJSON(Message{
					Identifier:  msg.Identifier,
					MessageType: "NoPlayer",
				})
			} else {
				msg.Connection.WriteJSON(Message{
					Identifier:  msg.Identifier,
					MessageType: "PlayerExists",
				})
			}
		}
		// Send it out to every client that is currently connected
		// for client := range clients {
		// 	err := client.WriteJSON(msg)
		// 	if err != nil {
		// 		log.Printf("error: %v", err)
		// 		client.Close()
		// 		delete(clients, client)
		// 	}
		// }
	}
}
