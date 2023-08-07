package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

type Article struct {
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	Category    string    `json:"category"`
	PublishedAt time.Time `json:"published_at"`
}

type ArticleReader struct {
	conn   *websocket.Conn
	sender TeaProgramSender
	done   chan struct{}
}

func (a *ArticleReader) Read() {
	go func() {
		<-a.done
		a.Close()
	}()

	go func() {
		for {
			_, message, err := a.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}

			var payload struct {
				Article
				Error string `json:"error"`
			}

			err = json.Unmarshal(message, &payload)
			if err != nil {
				log.Println("unmarshal:", err)
				continue
			}

			a.sender.Send(resultMsg{
				err: payload.Error,
				article: Article{
					Title:       payload.Title,
					Body:        payload.Body,
					Category:    payload.Category,
					PublishedAt: payload.PublishedAt,
				},
			})
		}
	}()
}

func (a *ArticleReader) Close() {
	defer a.conn.Close()
	err := a.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close:", err)
		return
	}
}

type TeaProgramSender interface {
	Send(msg tea.Msg)
}

func NewArticleReader(s TeaProgramSender) (*ArticleReader, error) {
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/subscribe"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"y-user-id": []string{*userID}})
	if err != nil {
		log.Fatal("dial:", err)
	}

	return &ArticleReader{conn: c, sender: s}, nil
}

var (
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle      = helpStyle.Copy().UnsetMargins()
	durationStyle = dotStyle.Copy()
	appStyle      = lipgloss.NewStyle().Margin(1, 2, 0, 2)
)

type resultMsg struct {
	article Article
	err     string
}

func (r resultMsg) String() string {
	return fmt.Sprintf(`
%s| %s | %s
%s
-------------------------
	`,
		durationStyle.Render(r.article.PublishedAt.String()),
		r.article.Category,
		r.article.Title,
		r.article.Body,
	)
}

type model struct {
	spinner  spinner.Model
	results  []resultMsg
	quitting bool
	err      string
}

func newModel() model {
	s := spinner.New()
	s.Style = spinnerStyle
	return model{
		spinner: s,
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.quitting = true
		return m, tea.Quit
	case resultMsg:
		if msg.err != "" {
			m.quitting = true
			m.err = msg.err
			return m, tea.Quit
		}
		m.results = append(m.results, msg)
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() string {
	var s string

	s += "\n\n"

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	if m.quitting {
		if m.err != "" {
			s += m.err
			s += "\n"
		}
		s += "\n"
	} else {
		s += m.spinner.View() + " Connected, receiving articles..."
		s += helpStyle.Render("Press any key to exit")
	}

	return appStyle.Render(s)
}

var addr = flag.String("addr", "localhost:8080", "http service address")
var userID = flag.String("user-id", "", "user id")

func main() {
	flag.Parse()

	p := tea.NewProgram(newModel())

	reader, err := NewArticleReader(p)
	if err != nil {
		log.Fatalf("Error creating article reader: %v", err)
	}

	reader.Read()

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
