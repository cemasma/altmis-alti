package main

import (
	"altmis-alti/room"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	go room.H.Run()

	r := gin.Default()
	r.Static("/cards", "./templates/cards")
	r.Static("/scripts", "./templates/scripts")
	r.LoadHTMLGlob("templates/views/*")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{})
	})

	r.POST("/room", func(c *gin.Context) {
		roomId := room.NewRequestId().String()

		c.JSON(http.StatusOK, gin.H{
			"roomId": roomId,
		})
	})

	r.GET("/room/:roomId", func(c *gin.Context) {
		c.HTML(http.StatusOK, "room.html", gin.H{
			"roomId": c.Param("roomId"),
		})
	})

	r.GET("/ws/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")
		serveWs(c.Writer, c.Request, roomId)
	})
	r.Run()
}

func serveWs(w http.ResponseWriter, r *http.Request, roomId string) {
	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c := &room.Connection{Id: room.NewRequestId().String(), Send: make(chan []byte, 256), Ws: ws}
	s := room.Subscription{Conn: c, Room: roomId}
	room.H.Register <- s
	go s.WritePump()
	go s.ReadPump()
}
