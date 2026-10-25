package loginpage

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/dece2183/go-clipboard"
	"github.com/dece2183/yamusic-tui/api"
	"github.com/dece2183/yamusic-tui/config"
	"github.com/dece2183/yamusic-tui/ui/style"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	err           error
	program       *tea.Program
	width, height int

	input     textinput.Model
	help      help.Model
	helpMap   *helpKeyMap
	clipboard *clipboard.Clipboard

	ctx         context.Context
	cancel      context.CancelFunc
	deviceLogin bool
	code        *api.DeviceCode
	poll        time.Duration
	status      string
	notice      string
}

// loginpage.Model constructor.
func New() *Model {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Model{
		input:       textinput.New(),
		help:        help.New(),
		helpMap:     newHelpMap(),
		clipboard:   clipboard.New(),
		ctx:         ctx,
		cancel:      cancel,
		deviceLogin: config.Current.DeviceLogin,
	}

	m.input.Width = 64
	m.input.CharLimit = 0
	m.input.Placeholder = "paste an access token"
	if !m.deviceLogin {
		m.input.Focus()
	}

	// No mouse capture: the code and the URL must stay selectable for copying.
	p := tea.NewProgram(m, tea.WithAltScreen())
	m.program = p

	return m
}

//
// model.Model interface implementation
//

func (m *Model) Run() error {
	defer m.cancel()
	_, err := m.program.Run()
	if err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *Model) Send(msg tea.Msg) {
	go m.program.Send(msg)
}

//
// tea.Model interface implementation
//

func (m *Model) Init() tea.Cmd {
	if !m.deviceLogin {
		return textinput.Blink
	}
	return tea.Batch(textinput.Blink, m.requestDeviceCodeCmd())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
		return m, tea.ClearScreen

	case deviceCodeMsg:
		if msg.err != nil {
			m.status = "Can't start browser login: " + msg.err.Error()
			return m, nil
		}
		m.code = msg.code
		m.notice = ""
		m.poll = time.Duration(msg.code.Interval) * time.Second
		m.status = "Or sign in at " + msg.code.VerificationURL + " and enter the code:"
		openBrowser(msg.code.VerificationURL)
		m.copyCode(true)
		return m, m.pollTokenCmd()

	case devicePollMsg:
		switch {
		case msg.err == nil:
			config.Current.Token = msg.token
			m.err = config.Save()
			return m, tea.Quit
		case errors.Is(msg.err, api.ErrAuthorizationPending):
			return m, m.pollTokenCmd()
		case errors.Is(msg.err, api.ErrSlowDown):
			m.poll += 5 * time.Second
			return m, m.pollTokenCmd()
		default:
			m.status = "Browser login failed: " + msg.err.Error()
			return m, nil
		}

	case tea.KeyMsg:
		controls := config.Current.Controls
		keypress := msg.String()

		switch {
		case keypress == "ctrl+y":
			m.copyCode(false)
			return m, nil
		case keypress == "ctrl+o":
			if m.code != nil {
				openBrowser(m.code.VerificationURL)
			}
			return m, nil
		case controls.Quit.Contains(keypress):
			return m, tea.Quit
		case !m.deviceLogin && controls.Apply.Contains(keypress):
			token := strings.TrimSpace(m.input.Value())
			if token == "" {
				m.status = "Paste a token or confirm the browser login"
				return m, nil
			}
			config.Current.Token = token
			m.err = config.Save()
			return m, tea.Quit
		default:
			if !m.deviceLogin {
				m.input, cmd = m.input.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	default:
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	title := "Enter your token"
	if m.deviceLogin {
		title = "Sign in to Yandex Music"
	}
	title = style.DialogTitleStyle.Render(title)
	lines := []string{title}
	statusLine := ""
	if m.status != "" {
		statusLine = style.DialogHelpStyle.Render(m.status)
		lines = append(lines, statusLine)
	}
	if m.code != nil {
		width := lipgloss.Width(statusLine)
		code := lipgloss.NewStyle().
			Foreground(style.AccentColor).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(style.AccentColor).
			Padding(0, 2).
			Render(m.code.UserCode)
		lines = append(lines, "", lipgloss.PlaceHorizontal(width, lipgloss.Center, code), "")
	}
	if m.notice != "" {
		lines = append(lines, style.DialogHelpStyle.Render(m.notice))
	}
	if !m.deviceLogin {
		lines = append(lines, m.input.View())
	}
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	content = style.DialogBoxStyle.Render(content)
	content = lipgloss.JoinVertical(lipgloss.Left, content, style.DialogHelpStyle.Render(m.help.View(m.helpMap)))

	dialog := lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)

	return dialog
}

//
// private methods
//

func (m *Model) resize(width, height int) {
	m.width, m.height = width, height
}

func (m *Model) copyCode(quiet bool) {
	if m.code == nil {
		return
	}
	if err := m.clipboard.CopyText(m.code.UserCode); err != nil {
		if !quiet {
			m.notice = "Can't copy the code: " + err.Error()
		}
		return
	}
	m.notice = "Code copied to clipboard"
}

func (m *Model) requestDeviceCodeCmd() tea.Cmd {
	ctx := m.ctx
	return func() tea.Msg {
		code, err := api.RequestDeviceCode(ctx, randomDeviceID(), "yamusic-tui")
		return deviceCodeMsg{code: code, err: err}
	}
}

// pollTokenCmd waits for the polling interval and asks the OAuth service once.
// Polling is driven by the UI instead of a long-running goroutine, so quitting
// the login page stops it immediately.
func (m *Model) pollTokenCmd() tea.Cmd {
	ctx := m.ctx
	code := m.code
	delay := m.poll
	return func() tea.Msg {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return devicePollMsg{err: ctx.Err()}
		case <-timer.C:
		}

		token, err := api.PollDeviceToken(ctx, code.DeviceCode)
		if err != nil {
			return devicePollMsg{err: err}
		}
		if _, err := api.NewClient(config.DirName, token); err != nil {
			return devicePollMsg{err: err}
		}
		return devicePollMsg{token: token}
	}
}

type deviceCodeMsg struct {
	code *api.DeviceCode
	err  error
}

type devicePollMsg struct {
	token string
	err   error
}

func randomDeviceID() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, 10)
	if _, err := rand.Read(id); err != nil {
		return "yamusic-tui"
	}
	for i := range id {
		id[i] = alphabet[int(id[i])%len(alphabet)]
	}
	return string(id)
}

// openBrowser is best effort: on headless/SSH systems it fails silently and the
// user can open the URL shown on the login screen manually. A command from
// $BROWSER takes precedence over the system default browser.
func openBrowser(url string) {
	if runBrowserCommand(os.Getenv("BROWSER"), url) {
		return
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	startBrowser(cmd)
}

// runBrowserCommand tries every entry of $BROWSER (':' separated) until one
// starts. An entry may contain %s for the URL, otherwise the URL is appended,
// e.g. BROWSER="chromium" or BROWSER="firefox -private-window".
func runBrowserCommand(browsers, url string) bool {
	for _, entry := range strings.Split(browsers, ":") {
		fields := strings.Fields(strings.TrimSpace(entry))
		if len(fields) == 0 {
			continue
		}

		placed := false
		args := make([]string, 0, len(fields)+1)
		for _, field := range fields {
			if strings.Contains(field, "%s") {
				field = strings.ReplaceAll(field, "%s", url)
				placed = true
			}
			args = append(args, field)
		}
		if !placed {
			args = append(args, url)
		}

		if startBrowser(exec.Command(args[0], args[1:]...)) {
			return true
		}
	}
	return false
}

func startBrowser(cmd *exec.Cmd) bool {
	if err := cmd.Start(); err != nil {
		return false
	}
	go cmd.Wait()
	return true
}
