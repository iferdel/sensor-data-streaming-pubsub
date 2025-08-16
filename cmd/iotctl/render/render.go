package render

import (
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/iferdel/sensor-data-streaming-server/cmd/iotctl/cmd"
)

type model struct {
	status int
	err    error
}

type errMsg struct {
	err error
}

func (e errMsg) Error() string {
	return e.err.Error()
}

type statusMsg int

func (m model) healthCheck() tea.Msg {

	c := &http.Client{Timeout: 10 * time.Second}
	res, err := c.Get(cmd.API_URL)

	if err != nil {
		return errMsg{err}
	}

	return statusMsg(res.StatusCode)
}

func (m model) Init() tea.Cmd {
	return m.healthCheck
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statusMsg:
		m.status = int(msg)
		return m, tea.Quit

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {

	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble: %v\n\n", m.err)
	}

	s := fmt.Sprintf("Checking %s ... ", cmd.API_URL)

	if m.status > 0 {
		s += fmt.Sprintf("%d %s!", m.status, http.StatusText(m.status))
	}

	return "\n" + s + "\n\n"
}
