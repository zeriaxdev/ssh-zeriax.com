package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

const (
	host = "0.0.0.0"
	port = "22"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519-1"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	m := model{
		Width:       pty.Window.Width,
		Height:      pty.Window.Height,
		connectedAt: time.Now(),
		r:           bubbletea.MakeRenderer(s),
	}
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

type tickMsg time.Time

type model struct {
	Width       int
	Height      int
	connectedAt time.Time
	r           *lipgloss.Renderer
}

func (m model) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tickMsg:
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	accent := lipgloss.Color("150") // soft sage green
	muted := lipgloss.Color("242")  // medium gray
	dim := lipgloss.Color("238")    // dark gray
	label := lipgloss.Color("246")  // light gray
	link := lipgloss.Color("255")   // idfk

	nameStyle := m.r.NewStyle().Bold(true).Foreground(accent)
	bodyStyle := m.r.NewStyle().Foreground(label)
	labelStyle := m.r.NewStyle().Foreground(muted)
	hintStyle := m.r.NewStyle().Foreground(dim)
	linkStyle := m.r.NewStyle().Foreground(link)
	pad := m.r.NewStyle().MarginLeft(4)

	elapsed := time.Since(m.connectedAt).Round(time.Second)
	h := int(elapsed.Hours())
	min := int(elapsed.Minutes()) % 60
	sec := int(elapsed.Seconds()) % 60
	uptime := fmt.Sprintf("%d:%02d:%02d", h, min, sec)

	lines := []string{
		"",
		nameStyle.Render("zeriax"),
		"",
		bodyStyle.Render("student, developer, designer"),
		bodyStyle.Render("studying electricity and automation at Omnia"),
		bodyStyle.Render("discord bot and react experience"),
		"",
		row(labelStyle, bodyStyle, "languages ", "typescript  go  python  rust (wip)"),
		row(labelStyle, bodyStyle, "os        ", "macos, arch linux, windows"),
		row(labelStyle, bodyStyle, "interests ", "frontend, APIs, self-hosting, UI/UX, MCP"),
		"",
		row(labelStyle, linkStyle, "github    ", "https://github.com/zeriaxdev"),
		row(labelStyle, linkStyle, "linkedin  ", "https://linkedin.com/in/egor-gaynutdinov"),
		row(labelStyle, linkStyle, "twitter   ", "https://x.com/zeriaxdev"),
		row(labelStyle, linkStyle, "email     ", "mailto:contact@zeriax.com"),
		row(labelStyle, bodyStyle, "matrix    ", "@egorrg:matrix.org"),
		"",
		"",
		hintStyle.Render(fmt.Sprintf("session %s    q/ctrl+c to quit", uptime)),
		"",
	}

	out := ""
	for _, l := range lines {
		out += pad.Render(l) + "\n"
	}
	return out
}

func row(labelStyle, bodyStyle lipgloss.Style, key, val string) string {
	return labelStyle.Render(key) + "  " + bodyStyle.Render(val)
}
