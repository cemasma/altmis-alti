package room

import (
	"altmis-alti/game"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

const (
	writeWait              = 10 * time.Second
	pongWait               = 60 * time.Second
	pingPeriod             = (pongWait * 9) / 10
	maxMessageSize         = 512
	connectionLimitPerRoom = 2
)

var H = Hub{
	Broadcast:  make(chan Message),
	Register:   make(chan Subscription),
	Unregister: make(chan Subscription),
	Rooms:      make(map[string]map[*Connection]bool),
	Games:      make(map[string]*game.Game),
}

type Connection struct {
	Id   string
	Ws   *websocket.Conn
	Send chan []byte
}

type Message struct {
	playerId string
	data     []byte
	room     string
}

type Subscription struct {
	Conn *Connection
	Room string
}

type Hub struct {
	Rooms      map[string]map[*Connection]bool
	Broadcast  chan Message
	Register   chan Subscription
	Unregister chan Subscription
	Games      map[string]*game.Game
}

type startMessage struct {
	Command      string      `json:"command"`
	PlayerId     string      `json:"playerId"`
	Cards        []game.Card `json:"cards"`
	CardOnGround game.Card   `json:"cardOnGround"`
	IsPlayerTurn bool        `json:"isPlayerTurn"`
}

type moveMessage struct {
	Command  string `json:"command"`
	Kind     string `json:"kind"`
	CardName string `json:"cardName"`
}

type moveResult struct {
	Command     string      `json:"command"`
	MoveMessage moveMessage `json:"moveMessage"`
	IsItMine    bool        `json:"isItMine"`
	TurnScore   uint8       `json:"turnScore"`
	NewCardKind string      `json:"newCardKind"`
	NewCardName string      `json:"newCardName"`
	Ground      []game.Card `json:"ground"`
}

type turnResult struct {
	Command         string `json:"command"`
	TotalScore      uint8  `json:"totalScore"`
	EnemyTotalScore uint8  `json:"enemyTotalScore"`
}

func (h *Hub) Run() {
	for {
		select {
		case s := <-h.Register:
			connections := h.Rooms[s.Room]
			if connections == nil {
				connections = make(map[*Connection]bool)
				h.Rooms[s.Room] = connections
			}

			if len(h.Rooms[s.Room]) < connectionLimitPerRoom {
				h.Rooms[s.Room][s.Conn] = true

				if len(h.Rooms[s.Room]) == connectionLimitPerRoom {
					var player1Id string
					var player2Id string

					for key := range h.Rooms[s.Room] {
						if player1Id == "" {
							player1Id = key.Id
						} else {
							player2Id = key.Id
						}
					}

					g := game.NewGame(s.Room, player1Id, player2Id)
					h.Games[s.Room] = &g
					h.sendNewTurnMessage(s.Room)
				}
			}
		case s := <-h.Unregister:
			connections := h.Rooms[s.Room]
			if connections != nil {
				if _, ok := connections[s.Conn]; ok {
					delete(connections, s.Conn)
					close(s.Conn.Send)
					if len(connections) == 0 {
						delete(h.Rooms, s.Room)
						delete(h.Games, s.Room)
					}
				}
			}
		case m := <-h.Broadcast:
			// connections := h.Rooms[m.room]
			g := h.Games[m.room]
			player := g.PlayerMap[m.playerId]
			var otherPlayer game.Player
			for key, val := range g.PlayerMap {
				if key != player.Id {
					otherPlayer = val
				}
			}

			// var output string
			var movement moveMessage
			err := json.Unmarshal(m.data, &movement)

			if err != nil {
				fmt.Errorf("%v", err)
				panic(err)
			}

			if movement.Command == "play" {
				var groundKind string

				if len(g.Ground) > 0 {
					groundKind = g.Ground[len(g.Ground)-1].Kind
				} else {
					groundKind = ""
				}

				playable := player.Move(movement.Kind, movement.CardName, groundKind, g.IsGameOff, len(g.Ground) == 0, len(g.Cards) == 0)

				if playable {
					_, index, _ := player.IsPlayerHaveCard(movement.Kind, movement.CardName)
					player.MoveCard(index, g)

					time.AfterFunc(time.Second, func() {
						if len(g.Ground) == 2 {
							g.CalculateTurnPoints(player, otherPlayer)
							player = g.PlayerMap[player.Id]
							otherPlayer = g.PlayerMap[otherPlayer.Id]

							var moveResultP1 moveResult
							var moveResultP2 moveResult
							moveResultP1.MoveMessage = movement
							moveResultP1.TurnScore = player.TurnScore
							moveResultP1.IsItMine = true
							moveResultP2.MoveMessage = movement
							moveResultP2.TurnScore = otherPlayer.TurnScore
							moveResultP2.IsItMine = false
							moveResultP1.Command = "play2"
							moveResultP2.Command = "play2"

							if len(player.Cards) == 6 {
								fmt.Printf("%v\n%v\n\n", player.Cards, otherPlayer.Cards)
								moveResultP1.NewCardKind = g.PlayerMap[player.Id].Cards[len(player.Cards)-1].Kind
								moveResultP1.NewCardName = g.PlayerMap[player.Id].Cards[len(player.Cards)-1].CardName

								moveResultP2.NewCardKind = otherPlayer.Cards[len(otherPlayer.Cards)-1].Kind
								moveResultP2.NewCardName = otherPlayer.Cards[len(otherPlayer.Cards)-1].CardName
							}

							message1, err := json.Marshal(moveResultP1)

							if err != nil {
								panic(err)
							}

							message2, err := json.Marshal(moveResultP2)

							if err != nil {
								panic(err)
							}

							h.sendDataToConnection(m.room, player.Id, message1)
							h.sendDataToConnection(m.room, otherPlayer.Id, message2)

							if len(g.Cards) == 0 {
								var removeGroundMoveResult moveResult
								removeGroundMoveResult.Command = "removeGround"

								msg, err := json.Marshal(removeGroundMoveResult)

								if err != nil {
									panic(err)
								}

								h.sendDataToConnection(m.room, player.Id, msg)
								h.sendDataToConnection(m.room, otherPlayer.Id, msg)
							}

							if g.IsTurnEnded(player) {
								g.CalculatePoints(player.Id, otherPlayer.Id)
								g.NewTurn()

								time.AfterFunc(time.Second*2, func() {
									h.sendEndTurnMessage(m.room, player.Id, otherPlayer.Id)

									time.AfterFunc(time.Second*3, func() {
										h.sendNewTurnMessage(m.room)
									})
								})
							}
						} else {
							g.SwitchTurns(player.Id, otherPlayer.Id)

							var moveResult moveResult
							moveResult.Command = "play1"
							moveResult.MoveMessage = movement
							moveResult.IsItMine = true
							moveResult.Ground = g.Ground

							message1, err := json.Marshal(moveResult)

							if err != nil {
								panic(err)
							}

							moveResult.IsItMine = false
							message2, err := json.Marshal(moveResult)

							if err != nil {
								panic(err)
							}

							h.sendDataToConnection(m.room, player.Id, message1)
							h.sendDataToConnection(m.room, otherPlayer.Id, message2)
						}
					})
				}
			} else if movement.Command == "switch" && player.Turn && g.Turn > 0 && len(g.Cards) > 2 {
				_, index, exist := player.IsPlayerHaveCard(g.Ace, "9")

				if exist {
					playerCard := player.Cards[index]
					player.Cards[index] = g.Cards[0]
					g.Cards[0] = playerCard

					g.PlayerMap[player.Id] = player

					var moveResult moveResult

					moveResult.Command = "switch"
					moveResult.NewCardName = player.Cards[index].CardName
					moveResult.NewCardKind = player.Cards[index].Kind
					moveResult.IsItMine = true

					message1, err := json.Marshal(moveResult)

					if err != nil {
						panic(err)
					}

					moveResult.NewCardName = g.Cards[0].CardName
					moveResult.NewCardKind = g.Cards[0].Kind
					moveResult.IsItMine = false

					message2, err := json.Marshal(moveResult)

					if err != nil {
						panic(err)
					}

					h.sendDataToConnection(m.room, player.Id, message1)
					h.sendDataToConnection(m.room, otherPlayer.Id, message2)
				}
			} else if movement.Command == "endTurn" && player.Turn && player.TurnScore >= 66 {
				g.CalculatePoints(player.Id, otherPlayer.Id)
				g.NewTurn()
				h.sendEndTurnMessage(m.room, player.Id, otherPlayer.Id)
				h.sendNewTurnMessage(m.room)
			} else if movement.Command == "gameOff" && player.Turn && len(g.Cards) > 2 && !g.IsGameOff {
				g.GameOff(player.Id)

				var moveResult moveResult
				moveResult.Command = "gameOff"

				msg, err := json.Marshal(moveResult)

				if err != nil {
					panic(err)
				}

				h.sendDataToConnection(m.room, player.Id, msg)
				h.sendDataToConnection(m.room, otherPlayer.Id, msg)
			}
			// h.sendDataToConnections(connections, m.room, output)
		}
	}
}

func (h *Hub) sendEndTurnMessage(roomId, player1Id, player2Id string) {
	g := h.Games[roomId]
	var turnResult turnResult

	turnResult.Command = "endTurn"
	turnResult.TotalScore = g.PlayerMap[player1Id].TotalScore
	turnResult.EnemyTotalScore = g.PlayerMap[player2Id].TotalScore

	turnResultPlayer1, err := json.Marshal(turnResult)

	if err != nil {
		panic(err)
	}

	turnResult.TotalScore = g.PlayerMap[player2Id].TotalScore
	turnResult.EnemyTotalScore = g.PlayerMap[player1Id].TotalScore

	turnResultPlayer2, err := json.Marshal(turnResult)

	if err != nil {
		panic(err)
	}

	h.sendDataToConnection(roomId, player1Id, turnResultPlayer1)
	h.sendDataToConnection(roomId, player2Id, turnResultPlayer2)
}

func (h *Hub) sendNewTurnMessage(roomId string) {
	g := h.Games[roomId]
	var player1Id string
	var player2Id string

	for _, p := range g.PlayerMap {
		if player1Id == "" {
			player1Id = p.Id
		} else {
			player2Id = p.Id
		}
	}

	startMessageP1 := startMessage{
		Command:      "start",
		PlayerId:     player1Id,
		Cards:        g.PlayerMap[player1Id].Cards,
		CardOnGround: g.Cards[0],
		IsPlayerTurn: g.PlayerMap[player1Id].Turn,
	}
	startMessageP2 := startMessage{
		Command:      "start",
		PlayerId:     player2Id,
		Cards:        g.PlayerMap[player2Id].Cards,
		CardOnGround: g.Cards[0],
		IsPlayerTurn: g.PlayerMap[player2Id].Turn,
	}

	byte1, _ := json.Marshal(startMessageP1)
	byte2, _ := json.Marshal(startMessageP2)
	h.sendDataToConnection(roomId, startMessageP1.PlayerId, byte1)
	h.sendDataToConnection(roomId, startMessageP2.PlayerId, byte2)
}

func (h *Hub) sendDataToConnection(roomId, playerId string, output []byte) {
	for c := range h.Rooms[roomId] {
		if c.Id == playerId {
			c.Send <- output
			break
		}
	}
}

func (h *Hub) sendDataToConnections(connections map[*Connection]bool, roomId string, output []byte) {
	for c := range connections {
		select {
		case c.Send <- output:
		default:
			close(c.Send)
			delete(connections, c)
			if len(connections) == 0 {
				delete(h.Rooms, roomId)
				delete(h.Games, roomId)
			}
		}
	}
}

func (s Subscription) ReadPump() {
	c := s.Conn
	defer func() {
		H.Unregister <- s
		c.Ws.Close()
	}()
	c.Ws.SetReadLimit(maxMessageSize)
	c.Ws.SetReadDeadline(time.Now().Add(pongWait))
	c.Ws.SetPongHandler(func(string) error { c.Ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.Ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		m := Message{c.Id, msg, s.Room}
		H.Broadcast <- m
	}
}

func (s *Subscription) WritePump() {
	c := s.Conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
func (c *Connection) write(mt int, payload []byte) error {
	c.Ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.Ws.WriteMessage(mt, payload)
}
