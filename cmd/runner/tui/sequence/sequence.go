package sequence

import (
	"context"
	"fmt"
	"github.com/jt05610/petri/db"
	"github.com/jt05610/petri/db/db"
	"github.com/jt05610/petri/sequence"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	*sequence.ListItem
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type Model struct {
	db        *prisma.RunClient
	sequences []db.RunModel
	list      list.Model
}

func New(client *db.PrismaClient) Model {
	m := Model{
		db: &prisma.RunClient{PrismaClient: client},
	}
	sequences, err := m.db.List(context.Background())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	l := makeList(sequences)
	m.list = list.New(l, list.NewDefaultDelegate(), 80, 10)
	m.list.Title = "Sequences"
	return m
}

func makeItem(r *sequence.ListItem) item {
	return item{
		ListItem: r,
		title:    r.Name,
		desc:     r.Description,
	}
}

func makeList(rs []*sequence.ListItem) []list.Item {
	items := make([]list.Item, len(rs))
	for i, r := range rs {
		items[i] = makeItem(r)
	}
	return items
}

func (m Model) Init() tea.Cmd {
	return nil
}

type Msg struct {
	*sequence.Sequence
	Error error
}

func (m Model) LoadSequence() tea.Msg {
	i := m.list.SelectedItem().(item)
	r, err := m.db.Load(context.Background(), i.ID)
	return Msg{r, err}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyEnter {
			return m, m.LoadSequence
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return "\n" + m.list.View()
}
