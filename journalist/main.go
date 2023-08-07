package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var auto = flag.Bool("auto", false, "send random articles every 2 seconds")

type Article struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	Category string `json:"category"`
}

func main() {
	flag.Parse()

	sender, err := NewArticleSender()
	if err != nil {
		log.Fatal(err)
	}
	defer sender.Close()

	if *auto {
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		t := time.NewTicker(2 * time.Second)

		for {
			select {
			case <-interrupt:
				err := sender.Close()
				if err != nil {
					log.Fatal(err)
				}
				return
			case <-t.C:
				article := randomArticle()
				fmt.Printf("Sending article:\n Title: %s\n Body: %s\n Category: %s\n----------------------------------------\n",
					article.Title,
					article.Body,
					article.Category,
				)
				if err := sender.Send(article); err != nil {
					log.Fatal(err)
				}
			}
		}

	}

	p := tea.NewProgram(initialModel(sender))

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type ArticleSender struct {
	conn *websocket.Conn
}

func (a *ArticleSender) Send(article Article) error {
	return a.conn.WriteJSON(article)
}

func NewArticleSender() (*ArticleSender, error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/publish"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	return &ArticleSender{conn: c}, nil
}

func (a *ArticleSender) Close() error {
	defer a.conn.Close()
	err := a.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}

	return nil
}

type (
	errMsg error
)

const (
	title = iota
	body
	category
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	focusedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type model struct {
	inputs  []textinput.Model
	focused int
	err     error
	sender  *ArticleSender
}

func initialModel(s *ArticleSender) model {
	var inputs []textinput.Model = make([]textinput.Model, 3)
	inputs[title] = textinput.New()
	inputs[title].Placeholder = "Title"
	inputs[title].Focus()
	inputs[title].CharLimit = 50
	inputs[title].Width = 50
	inputs[title].Prompt = ""

	inputs[body] = textinput.New()
	inputs[body].Placeholder = "Body"
	inputs[body].CharLimit = 250
	inputs[body].Width = 250
	inputs[body].Prompt = ""

	inputs[category] = textinput.New()
	inputs[category].Placeholder = "Category"
	inputs[category].CharLimit = 10
	inputs[category].Width = 10
	inputs[category].Prompt = ""

	return model{
		inputs:  inputs,
		focused: 0,
		sender:  s,
		err:     nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs) {
				if err := m.sender.Send(m.toArticle()); err != nil {
					m.err = err
					return m, nil
				}

				m.reset()
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			err := m.sender.Close()
			if err != nil {
				m.err = err
			}

			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}

		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		if m.focused < len(m.inputs) {
			m.inputs[m.focused].Focus()
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	button := &blurredButton
	if m.focused == len(m.inputs) {
		button = &focusedButton
	}

	return fmt.Sprintf(
		`
 %s
 %s

 %s
 %s

 %s
 %s

 %s
`,
		inputStyle.Width(30).Render("Title"),
		m.inputs[title].View(),
		inputStyle.Width(30).Render("Body"),
		m.inputs[body].View(),
		inputStyle.Width(30).Render("Category"),
		m.inputs[category].View(),
		*button,
	)

}

func (m *model) nextInput() {
	m.focused = (m.focused + 1) % (len(m.inputs) + 1)
}

func (m *model) prevInput() {
	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}

func (m *model) reset() {
	for i := range m.inputs {
		m.inputs[i].Reset()
	}
}

func (m *model) toArticle() Article {
	return Article{
		Title:    m.inputs[title].Value(),
		Body:     m.inputs[body].Value(),
		Category: m.inputs[category].Value(),
	}
}

func randomArticle() Article {
	articles := []Article{
		{
			Title:    "Japan's new emperor: What's his name?",
			Body:     "Japan's new emperor Naruhito will formally ascend the Chrysanthemum throne on Wednesday, a day after his father's historic abdication.",
			Category: "Japan",
		},
		{
			Title:    "Japan's emperor: The cost of keeping one",
			Body:     "Japan's Emperor Akihito has declared his abdication in a historic ceremony at the Imperial Palace in Tokyo.",
			Category: "Japan",
		},
		{
			Title:    "Best season to visit Japan",
			Body:     "Although Japan is a year-round destination, the best time to visit Japan is in spring (March & April) or autumn (October & November), when the weather is great and there are many cultural festivals.",
			Category: "Travel",
		},
		{
			Title:    "Toyota's battery 'breakthrough' could lead to more range, faster charges",
			Body:     "Toyota says it has developed a more efficient and safer way of producing smaller, lighter-weight lithium-ion batteries that could increase EV range and reduce charging times.",
			Category: "Technology",
		},
	}
	return articles[rand.Intn(len(articles))]
}
