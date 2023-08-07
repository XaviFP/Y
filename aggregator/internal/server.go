package aggregator

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct {
	articleRepo ArticleRepository
}

func NewServer(articleRepository ArticleRepository) *Server {
	return &Server{
		articleRepo: articleRepository,
	}
}

func (s *Server) publish(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	defer c.Close()

	var a Article
	for {
		err := c.ReadJSON(&a)
		if err != nil {
			log.Println("closing connection. read error:", err)
			break
		}

		err = s.articleRepo.store(a)
		if err != nil {
			log.Println("closing connecton. processing error:", err)
			break
		}
	}
}

func (s *Server) RegistersRoutes() {
	http.HandleFunc("/publish", s.publish)
}
