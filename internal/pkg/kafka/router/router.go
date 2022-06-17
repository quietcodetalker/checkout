package router

import (
	"context"
	"github.com/Shopify/sarama"
	"log"
	"sync"
)

// HandlerFunc receives a topic name with a message from kafka and handles is somehow.
type HandlerFunc func(ctx context.Context, topic string, msg []byte) error

// Middleware takes HandlerFunc in and returns HandlerFunc as well.
type Middleware func(fn HandlerFunc) HandlerFunc

// SaramaRouter routes incoming kafka messages and route them out between handlers based on topic names.
type SaramaRouter struct {
	handlers map[string][]HandlerFunc
	hm       sync.RWMutex

	middlewares []Middleware
	mm          sync.RWMutex
}

// NewSaramaRouter craetes an instance of SaramaRouter.
func NewSaramaRouter() *SaramaRouter {
	return &SaramaRouter{
		handlers: make(map[string][]HandlerFunc),
	}
}

// Handle binds a provided HandlerFunc to a given topic.
func (r *SaramaRouter) Handle(topic string, fn HandlerFunc) {
	r.hm.Lock()
	r.handlers[topic] = append(
		r.handlers[topic],
		r.wrap(fn),
	)
	r.hm.Unlock()
}

// Use adds Middleware to middlewares queue.
func (r *SaramaRouter) Use(m Middleware) {
	r.hm.RLock()
	defer r.hm.RUnlock()
	if len(r.handlers) > 0 {
		log.Panic("Use method must be used before any route is set")
	}

	r.mm.Lock()
	r.middlewares = append(r.middlewares, m)
	r.mm.Unlock()
}

func (r *SaramaRouter) wrap(fn HandlerFunc) HandlerFunc {
	r.mm.RLock()
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		fn = r.middlewares[i](fn)
	}
	r.mm.RUnlock()

	return fn
}

// Setup does nothing but makes SaramaRouter match sarama.ConsumerGroupHandler interface.
func (r *SaramaRouter) Setup(s sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup does nothing but makes SaramaRouter match sarama.ConsumerGroupHandler interface.
func (r *SaramaRouter) Cleanup(s sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim receives messages from a channel and calls appropriate handlers based on routes.
func (r *SaramaRouter) ConsumeClaim(s sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-s.Context().Done():
			return nil
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			r.hm.RLock()
			for topic, topicHandlers := range r.handlers {
				if msg.Topic == topic {
					for _, handle := range topicHandlers {
						topic := topic
						go func() {
							if err := handle(s.Context(), topic, msg.Value); err != nil {
								log.Printf("handle message err: %v", err)
							}
						}()
					}
				}
			}
			r.hm.RUnlock()
		}
	}

	return nil
}
