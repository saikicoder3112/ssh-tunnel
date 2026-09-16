package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"mytunnel/internal/config"
	"mytunnel/internal/ssh"
	"mytunnel/internal/tui"
)

func main() {
	cfg, path, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	defer ssh.ReapAll(cfg)

	model := tui.New(cfg, path)
	program := tea.NewProgram(model, tea.WithAltScreen())

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigc
		ssh.ReapAll(cfg)
		program.Quit()
	}()

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "run tui: %v\n", err)
		os.Exit(1)
	}
}
