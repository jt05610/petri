package device

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jt05610/petri/sequence"
)

// sessionState is used to track which InstModel is focused
type sessionState uint

var (
	// Available spinners
	spinners = []spinner.Spinner{
		spinner.Line,
		spinner.Dot,
		spinner.MiniDot,
		spinner.Jump,
		spinner.Pulse,
		spinner.Points,
		spinner.Globe,
		spinner.Moon,
		spinner.Monkey,
	}
	modelStyle = lipgloss.NewStyle().
			Width(15).
			Height(5).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())
	focusedModelStyle = lipgloss.NewStyle().
				Width(15).
				Height(5).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("69"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

type Model struct {
	state    sessionState
	s        *sequence.Sequence
	instance []*InstModel
	index    int
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.Next()
		}
		u, cmd := m.instance[m.index].Update(msg)
		cmds = append(cmds, cmd)
		m.instance[m.index] = u.(*InstModel)

	}
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var s string
	if len(m.instance) == 0 {
		return helpStyle.Render("no devices")
	}
	model := m.currentFocusedModel()
	ss := make([]string, len(m.instance))
	for i, inst := range m.instance {
		if i == m.index {
			ss[i] = focusedModelStyle.Render(inst.View())
		}
		ss[i] = modelStyle.Render(inst.View())
	}
	// split into groups of 2 and make line for each group of 2
	if len(ss)%2 != 0 {
		ss = append(ss, "")
	}
	gg := make([]string, len(ss)/2)
	for i := 0; i < len(ss); i += 2 {
		gg[i/2] = lipgloss.JoinHorizontal(lipgloss.Top, ss[i], ss[i+1])
	}
	// then join all the lines together
	s = lipgloss.JoinVertical(lipgloss.Left, gg...)

	s += helpStyle.Render(fmt.Sprintf("\ntab: focus next • n: new %s • q: exit\n", model))
	return "\n" + s
}

func (m *Model) currentFocusedModel() string {
	return m.instance[m.index].Name
}

func (m *Model) Next() {
	if m.index == len(spinners)-1 {
		m.index = 0
	} else {
		m.index++
	}
}

func New(s *sequence.Sequence) *Model {
	ii := make([]*InstModel, len(s.Devices()))
	for i, inst := range s.Devices() {
		ii[i] = InstanceModel(inst.Instances)
	}
	return &Model{
		state:    0,
		s:        s,
		instance: ii,
	}

}
