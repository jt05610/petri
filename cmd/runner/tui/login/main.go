package login

import (
	"context"
	"errors"
	"fmt"
	"github.com/jt05610/petri/prisma/db"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

const (
	username = iota
	password
)

const (
	hotPink  = lipgloss.Color("#FF06B7")
	darkGray = lipgloss.Color("#767676")
)

var (
	inputStyle    = lipgloss.NewStyle().Foreground(hotPink)
	continueStyle = lipgloss.NewStyle().Foreground(darkGray)
)

type Model struct {
	db      *db.PrismaClient
	inputs  []textinput.Model
	focused int
	err     error
}

func New(client *db.PrismaClient) *Model {
	var inputs = make([]textinput.Model, 2)
	inputs[username] = textinput.New()
	inputs[username].Placeholder = "username@example.com"
	inputs[username].Focus()
	inputs[username].CharLimit = 254
	inputs[username].Width = 30
	inputs[username].Prompt = ""

	inputs[password] = textinput.New()
	inputs[password].Placeholder = "password"
	inputs[password].CharLimit = 128
	inputs[password].Width = 30
	inputs[password].Prompt = ""
	inputs[password].EchoMode = textinput.EchoPassword
	inputs[password].EchoCharacter = '•'

	return &Model{
		inputs:  inputs,
		db:      client,
		focused: 0,
		err:     nil,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

type LoginResult struct {
	ID    string
	Error error
}

func (m *Model) Login() tea.Msg {
	id, err := validateLogin(m.db, m.inputs[username].Value(), m.inputs[password].Value())
	return LoginResult{id, err}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds = make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				return m, m.Login
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func writeLine(bld *strings.Builder, l string) *strings.Builder {
	_, err := bld.WriteString(l + "\n")
	if err != nil {
		panic(err)
	}
	return bld
}

func (m *Model) View() string {
	bld := &strings.Builder{}
	writeLine(bld, "\n")
	writeLine(bld, inputStyle.Width(30).Render("Username"))
	writeLine(bld, m.inputs[username].View())
	writeLine(bld, "\n")
	writeLine(bld, inputStyle.Width(30).Render("Password"))
	writeLine(bld, m.inputs[password].View())
	return bld.String()
}

func (m *Model) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

func (m *Model) prevInput() {
	m.focused--
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}

func validateLogin(c *db.PrismaClient, email, password string) (string, error) {
	u, err := c.User.FindUnique(db.User.Email.Equals(email)).With(db.User.Password.Fetch()).Exec(context.Background())
	if err != nil {
		panic(err)
	}
	if m, found := u.Password(); !found {
		return "", errors.New("user has no password")
	} else if bcrypt.CompareHashAndPassword([]byte(m.Hash), []byte(password)) != nil {
		return "", errors.New("password does not match")
	}
	return u.ID, nil
}

func saveLogin(id string) {
	fmt.Println("Saving Login")
	df, err := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := df.Close()
		if err != nil {
			panic(err)
		}
	}()
	_, err = df.WriteString(fmt.Sprintf("\nAUTHOR_ID=%s\n", id))
	if err != nil {
		panic(err)
	}
}
