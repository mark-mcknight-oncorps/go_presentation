package pubsub

import "fmt"

type Pubsub struct {
	// actions map topics to an array of payloads
	actions map[string][]string
	// subscriptions map topics to an array of channels
	subscriptions map[string][]chan string
}

func (p *Pubsub) Subscribe(channel chan string, topic string) {
	fmt.Printf("Server subscribing to topic %v\n", topic)
	if _, ok := p.subscriptions[topic]; !ok {
		p.subscriptions[topic] = make([]chan string, 0)
	}
	p.subscriptions[topic] = append(p.subscriptions[topic], channel)

	// Now go through existing actions and send previous actions with requested topic
	if actions, ok := p.actions[topic]; ok {
		for _, action := range actions {
			channel <- action
		}
	}
}

func (p *Pubsub) Publish(topic string, payload string) {
	if _, ok := p.actions[topic]; !ok {
		p.actions[topic] = make([]string, 0)
	}

	p.actions[topic] = append(p.actions[topic], payload)

	// Now go through existing subscriptions and send action to subscribed channels
	if subs, ok := p.subscriptions[topic]; ok {
		for _, sub := range subs {
			sub <- payload
		}
	}
}

func NewPubsub() *Pubsub {
	return &Pubsub{make(map[string][]string), make(map[string][]chan string)}
}
