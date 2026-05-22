package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/soakhan/cremio/internal/config"
	"github.com/soakhan/cremio/internal/player"
	"github.com/soakhan/cremio/internal/stremio"
)

type streamItem struct {
	stream stremio.Stream
}

func (s streamItem) Title() string { return s.stream.DisplayName() }
func (s streamItem) Description() string {
	if s.stream.Description != "" {
		return s.stream.Description
	}
	if s.stream.Title != "" && s.stream.Title != s.stream.Name {
		return s.stream.Title
	}
	url := s.stream.PlayableURL()
	if len(url) > 60 {
		return url[:60] + "..."
	}
	return url
}
func (s streamItem) FilterValue() string { return s.stream.DisplayName() }

type StreamsModel struct {
	list    list.Model
	spinner spinner.Model
	client  *stremio.Client
	config  *config.Config
	loading bool
	err     error
	playErr error
	width   int
	height  int
}

type streamsLoadedMsg struct {
	streams []stremio.Stream
}
type streamsErrorMsg struct {
	err error
}
type mpvLaunchedMsg struct{}
type mpvErrorMsg struct {
	err error
}

func NewStreamsModel(client *stremio.Client, cfg *config.Config) StreamsModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Streams"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	return StreamsModel{
		list:    l,
		spinner: s,
		client:  client,
		config:  cfg,
	}
}

func (m *StreamsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.list.SetSize(w, h-2)
}

func (m StreamsModel) LoadStreams(nav NavigateToStreamsMsg) tea.Cmd {
	return func() tea.Msg {
		var allStreams []stremio.Stream
		ctx := context.Background()

		for _, addonURL := range m.config.Addons {
			resp, err := m.client.FetchStreams(ctx, addonURL, nav.Type, nav.ID)
			if err != nil {
				continue
			}
			allStreams = append(allStreams, resp.Streams...)
		}

		if len(allStreams) == 0 {
			return streamsErrorMsg{err: fmt.Errorf("no streams found")}
		}
		return streamsLoadedMsg{streams: allStreams}
	}
}

func (m StreamsModel) Update(msg tea.Msg) (StreamsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case streamsLoadedMsg:
		m.loading = false
		items := make([]list.Item, len(msg.streams))
		for i, s := range msg.streams {
			items[i] = streamItem{stream: s}
		}
		m.list.SetItems(items)
		return m, nil

	case streamsErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case mpvLaunchedMsg:
		return m, nil

	case mpvErrorMsg:
		m.playErr = msg.err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(streamItem); ok {
				url := item.stream.PlayableURL()
				return m, func() tea.Msg {
					err := player.PlayWithMPV(url)
					if err != nil {
						return mpvErrorMsg{err: err}
					}
					return mpvLaunchedMsg{}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m StreamsModel) View() string {
	if m.loading {
		return m.spinner.View() + " Loading streams..."
	}
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	view := m.list.View()
	if m.playErr != nil {
		view += "\n" + ErrorStyle.Render(fmt.Sprintf("MPV error: %v", m.playErr))
	}
	view += "\n" + HelpStyle.Render("enter: play with mpv • esc: back")
	return view
}
