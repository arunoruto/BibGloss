package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const url = "https://charm.sh/"

type (
	statusMsg int
	// errMsg    error
	errMsg struct{ error }
)

type model struct {
	textInput textinput.Model
	status    int
	err       error
}

// Default values
func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "10.1016/j.icarus.2016.12.026"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	return model{
		textInput: ti,
		status:    0,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	// catch key presses
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
			// This will result in catching any keys
			// default:
			// 	return m, nil
		}

	// handle the status message of the request
	case statusMsg:
		m.status = int(msg)
		return m, tea.Quit

	// handle the error messages
	case errMsg:
		m.err = msg
		return m, nil

		// render the text input
		// default:
		// 	var cmd tea.Cmd
		// 	m.textInput, cmd = m.textInput.Update(msg)
		// 	return m, cmd
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.textInput.View() != "" {
		return fmt.Sprintf(
			"What’s your favorite Pokémon?\n\n%s\n\n%s",
			m.textInput.View(),
			"(esc to quit)",
		) + "\n"
	}
	return ""
}

func checkServer() tea.Msg {
	c := &http.Client{
		Timeout: 10 * time.Second,
	}
	res, err := c.Get(url)
	if err != nil {
		return errMsg{err}
	}
	defer res.Body.Close() // nolint:errcheck

	return statusMsg(res.StatusCode)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
