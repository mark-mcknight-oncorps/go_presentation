package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mark-mcknight-oncorps/go_presentation/pubsub"
)

const xDim = 20
const yDim = 20

var upgrader = websocket.Upgrader{}

type Server struct {
	name         string
	channel      chan string
	pubsub       *pubsub.Pubsub
	neighborhood map[string]int
	alive        int
	changed      bool
}

func NewServer(
	x int,
	y int,
	pubsub *pubsub.Pubsub,
) *Server {
	channel := make(chan string, 10)
	neighborhood := make(map[string]int)
	alive := rand.Intn(2)
	changed := false
	server := Server{fmt.Sprintf("%v,%v", x, y), channel, pubsub, neighborhood, alive, changed}
	pubsub.Subscribe(channel, "director")
	for cx := -1; cx < 2; cx++ {
		for cy := -1; cy < 2; cy++ {
			if cx != 0 || cy != 0 {
				neighbor := fmt.Sprintf("%v,%v", (x+cx+xDim)%xDim, (y+cy+yDim)%yDim)
				neighborhood[neighbor] = 0
				pubsub.Subscribe(channel, neighbor)
			}
		}
	}
	return &server
}

func StartServer(server *Server) {
	fmt.Printf("Starting server %v\n", server.name)
	// if you're alive, report your initial value
	if server.alive == 1 {
		server.pubsub.Publish(server.name, server.name)
	}
	for payload := range server.channel {
		if payload == "propagate" {
			neighborCount := 0
			for _, val := range server.neighborhood {
				neighborCount += val
			}
			if server.alive == 1 && (neighborCount < 2 || neighborCount > 3) {
				server.alive = 0
				server.changed = true
			} else if server.alive == 0 && neighborCount == 3 {
				server.alive = 1
				server.changed = true
			}
		} else if payload == "report" {
			if server.changed {
				server.pubsub.Publish(server.name, server.name)
				server.changed = false
			}
		} else {
			// The neighborhood has changed
			server.neighborhood[payload] = 1 - server.neighborhood[payload]
		}
	}
}

func StopServer(server *Server) {
	fmt.Printf("Stopping server %v\n", server.name)
	close(server.channel)
}

func StartDirector(pubsub *pubsub.Pubsub, servers []*Server, ws chan string) {
	ticker := time.NewTicker(500 * time.Millisecond)
	endTimer := time.NewTimer(30 * time.Second)
	propagate := true

	for {
		select {
		case <-endTimer.C:
			ticker.Stop()
			pubsub.Publish("endSimulation", "end")
			fmt.Println("Ticker stopped")
		case <-ticker.C:
			if propagate {
				fmt.Println("propagate")
				ws <- "propogate"
				pubsub.Publish("director", "propagate")
			} else {
				fmt.Println("report")
				ws <- "report"
				pubsub.Publish("director", "report")
			}
			propagate = !propagate
		}
	}
}

func SetupWebsocket(wsChannel chan string) {
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebsocket(wsChannel, w, r)
	})

	http.ListenAndServe(":8080", nil)
}

func handleWebsocket(wsChannel chan string, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	go func() {
		for payload := range wsChannel {
			c.WriteMessage(websocket.TextMessage, []byte(payload))
		}
	}()

	go func() {
		for {
			_, message, _ := c.ReadMessage()
			wsChannel <- string(message)
			break
		}
	}()
}

func main() {
	// set-up websocket connection
	wsChannel := make(chan string)
	go SetupWebsocket(wsChannel)

	pubsub := pubsub.NewPubsub()

	servers := make([]*Server, (xDim * yDim))

	for x := 0; x < xDim; x++ {
		for y := 0; y < yDim; y++ {
			servers[x+(y*xDim)] = NewServer(x, y, pubsub)
			pubsub.Subscribe(wsChannel, fmt.Sprintf("%v,%v", x, y))
		}
	}

	// Wait until start is pushed
	<-wsChannel

	for _, server := range servers {
		go StartServer(server)
	}

	endChannel := make(chan string)
	pubsub.Subscribe(endChannel, "endSimulation")

	go StartDirector(pubsub, servers, wsChannel)

	// This blocks until we receive a message on the endChannel
	<-endChannel

	fmt.Println("received on endChannel, now going to stop servers")

	for _, server := range servers {
		StopServer(server)
	}
}
