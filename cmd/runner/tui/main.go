package main

import (
	"fmt"
	"github.com/jt05610/petri/cmd/runner/tui/device"
	"github.com/jt05610/petri/cmd/runner/tui/login"
	"github.com/jt05610/petri/cmd/runner/tui/sequence"
	"github.com/jt05610/petri/db/db"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

type Tab struct {
	tea.Model
	Title string
}

const (
	sessionLogin sessionState = iota
	sessionSequence
	sessionDevice
	sessionParameters
	sessionRun
)

type model struct {
	state sessionState
	db    *db.PrismaClient
	Tabs  []Tab
}

func (m *model) Init() tea.Cmd {

	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case sequence.Msg:
		if msg.Error != nil {
			return m, tea.Quit
		}
		m.state = sessionDevice
		m.Tabs[m.state].Model = device.New(msg.Sequence)

	case login.LoginResult:
		if msg.Error != nil {
			return m, tea.Quit
		}
		m.state = sessionSequence
	}
	// update content of active tab
	var cmd tea.Cmd
	m.Tabs[m.state].Model, cmd = m.Tabs[m.state].Model.Update(msg)

	return m, cmd
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	instateBorder  = tabBorderWithBottom("┴", "─", "┴")
	stateBorder    = tabBorderWithBottom("┘", " ", "└")
	docStyle       = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	instateStyle   = lipgloss.NewStyle().Border(instateBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	stateStyle     = instateStyle.Copy().Border(stateBorder, true)
	windowStyle    = lipgloss.NewStyle().BorderForeground(highlightColor).Padding(2, 0).Align(lipgloss.Center).Border(lipgloss.NormalBorder()).UnsetBorderTop()
)

func (m *model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == int(m.state)
		if isActive {
			style = stateStyle.Copy()
		} else {
			style = instateStyle.Copy()
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "│"
		} else if isLast && !isActive {
			border.BottomRight = "┤"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t.Title))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)
	doc.WriteString(row)
	return docStyle.Render(doc.String() + m.Tabs[m.state].View())
}

func New(client *db.PrismaClient) tea.Model {
	return &model{
		state: sessionLogin,
		Tabs: []Tab{
			{Title: "Login", Model: login.New(client)},
			{Title: "Sequence", Model: sequence.New(client)},
			{Title: "Device"},
			{Title: "Parameters"},
			{Title: "Run"},
		},
	}
}

func main() {
	dbClient := db.NewClient()
	if err := dbClient.Connect(); err != nil {
		panic(err)
	}

	defer func() {
		if err := dbClient.Disconnect(); err != nil {
			panic(err)
		}
	}()
	if _, err := tea.NewProgram(New(dbClient)).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
