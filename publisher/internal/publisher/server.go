package publisher

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	broker Broker
}

func NewServer(broker Broker) *Server {
	return &Server{

		broker: broker,
	}
}

func (s *Server) Start() {
	s.broker.Run()
}

type errorPayload struct {
	Err string `json:"error"`
}

func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()

	userID := r.Header.Get("Y-User-ID")
	if userID == "" {
		log.Println("user id not found")
		c.WriteJSON(errorPayload{Err: "User id not found"})
		return
	}

	ch := s.broker.AddSubscriber(userID)

	if ch == nil {
		log.Println(fmt.Sprintf("subscriber %s already exists", userID))
		c.WriteJSON(errorPayload{Err: "Upgrade to premium to use Y network from multiple devices"})
		return
	}

	log.Println(fmt.Sprintf("subscriber %s connected", userID))
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, _, err := c.NextReader()
			if err != nil {
				log.Println("connection interrupted:", err)
				break
			}
		}
	}()

	for {
		select {
		case <-done:
			log.Println(fmt.Sprintf("subscriber %s disconnected", userID))
			s.broker.RemoveSubscriber(userID)

			return

		case article := <-ch:
			err := c.WriteJSON(article)
			if err != nil {
				log.Println("write error:", err)
				s.broker.RemoveSubscriber(userID)

				return
			}
		}
	}
}

func (s *Server) RegistersRoutes() {
	http.HandleFunc("/subscribe", s.subscribe)
}
