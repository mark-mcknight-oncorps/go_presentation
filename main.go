package main

import (
	"fmt"
	"time"

	"github.com/mark-mcknight-oncorps/go_presentation/pubsub"
)

type Process func(string) string

type Server struct {
	name     string
	channel  chan string
	inTopic  string
	outTopic string
	delay    time.Duration
	pubsub   *pubsub.Pubsub
	process  Process
}

func NewServer(
	name string,
	inTopic string,
	outTopic string,
	delay time.Duration,
	pubsub *pubsub.Pubsub,
	process Process,
) *Server {
	channel := make(chan string)
	server := Server{name, channel, inTopic, outTopic, delay, pubsub, process}
	pubsub.Subscribe(channel, inTopic)
	return &server
}

func StartServer(server *Server) {
	fmt.Printf("Starting server %v\n", server.name)
	for payload := range server.channel {
		fmt.Printf("%v receiving payload %v on topic %v\n", server.name, payload, server.inTopic)
		time.Sleep(server.delay * time.Millisecond)
		out := server.process(payload)
		fmt.Printf("%v sending payload %v on topic %v\n", server.name, out, server.outTopic)
		server.pubsub.Publish(server.outTopic, out)
	}
}

func StopServer(server *Server) {
	fmt.Printf("Stopping server %v\n", server.name)
	close(server.channel)
}

func main() {
	pubsub := pubsub.NewPubsub()
	firstProcess := func(in string) string { return in + "Second" }
	firstServer := NewServer("firstServer", "first", "second", 1000, pubsub, firstProcess)
	secondProcess := func(in string) string { return in + "Third" }
	secondServer := NewServer("secondServer", "second", "third", 1000, pubsub, secondProcess)
	thirdProcess := func(in string) string { return in + "Fourth" }
	thirdServer := NewServer("thirdServer", "third", "fourth", 1000, pubsub, thirdProcess)

	go StartServer(firstServer)
	go StartServer(secondServer)
	go StartServer(thirdServer)

	pubsub.Publish("first", "First")

	endChannel := make(chan string)
	pubsub.Subscribe(endChannel, "fourth")

	// This blocks until we receive a message on the endChannel
	<-endChannel

	StopServer(firstServer)
	StopServer(secondServer)
	StopServer(thirdServer)
}
