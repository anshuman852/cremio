package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itssoap/cremio/internal/config"
	"github.com/itssoap/cremio/internal/player"
	"github.com/itssoap/cremio/internal/stremio"
)

type streamItem struct {
	stream       stremio.Stream
	episodeLabel string
	videoID      string
}

func (s streamItem) Title() string {
	name := s.stream.DisplayName()
	if s.episodeLabel != "" {
		return s.episodeLabel + " | " + name
	}
	return name
}
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
	list          list.Model
	spinner       spinner.Model
	filterInput   textinput.Model
	client        *stremio.Client
	config        *config.Config
	allItems      []streamItem
	pendingVideos []stremio.Video
	pendingType   string
	contentID     string
	contentType   string
	filterActive  bool
	loading       bool
	launching     bool
	launched      bool
	err           error
	playErr       error
	width         int
	height        int
}

type streamsLoadedMsg struct {
	streams []stremio.Stream
}
type streamsErrorMsg struct {
	err error
}
type mpvLaunchedMsg struct {
	videoID   string
	videoType string
}
type mpvErrorMsg struct {
	err error
}
type clearLaunchedMsg struct{}

func NewStreamsModel(client *stremio.Client, cfg *config.Config) StreamsModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Streams"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	fi := textinput.New()
	fi.Placeholder = "Filter: +include -exclude ..."
	fi.CharLimit = 200

	return StreamsModel{
		list:        l,
		spinner:     s,
		filterInput: fi,
		client:      client,
		config:      cfg,
	}
}

func (m *StreamsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.filterInput.Width = w - 4
	m.list.SetSize(w, h-7) // account for filter input (2 lines), help (1 line), spacing
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

func (m StreamsModel) LoadAllStreams(nav NavigateToAllStreamsMsg, filter Filter) tea.Cmd {
	return func() tea.Msg {
		var allStreams []labeledStream
		ctx := context.Background()

		for _, video := range nav.Videos {
			label := fmt.Sprintf("S%02dE%02d", video.Season, video.Episode)
			for _, addonURL := range m.config.Addons {
				resp, err := m.client.FetchStreams(ctx, addonURL, nav.Type, video.ID)
				if err != nil {
					continue
				}
				for _, s := range resp.Streams {
					if filter.IsEmpty() || filter.Match(s.Name, s.Title) {
						allStreams = append(allStreams, labeledStream{stream: s, label: label, videoID: video.ID})
					}
				}
			}
		}

		if len(allStreams) == 0 {
			return streamsErrorMsg{err: fmt.Errorf("no streams found matching filter")}
		}
		return allStreamsLoadedMsg{streams: allStreams}
	}
}

type labeledStream struct {
	stream  stremio.Stream
	label   string
	videoID string
}

type allStreamsLoadedMsg struct {
	streams []labeledStream
}

func (m *StreamsModel) applyFilter() {
	f := ParseFilter(m.filterInput.Value())
	var filtered []list.Item
	for _, item := range m.allItems {
		if f.IsEmpty() || f.Match(item.stream.Name, item.stream.Title) {
			filtered = append(filtered, item)
		}
	}
	m.list.SetItems(filtered)
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
		m.allItems = make([]streamItem, len(msg.streams))
		for i, s := range msg.streams {
			m.allItems[i] = streamItem{stream: s}
		}
		m.applyFilter()
		return m, nil

	case allStreamsLoadedMsg:
		m.loading = false
		m.allItems = make([]streamItem, len(msg.streams))
		for i, ls := range msg.streams {
			m.allItems[i] = streamItem{stream: ls.stream, episodeLabel: ls.label, videoID: ls.videoID}
		}
		m.applyFilter()
		return m, nil

	case streamsErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case mpvLaunchedMsg:
		m.launching = false
		m.launched = true
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return clearLaunchedMsg{}
		})

	case mpvErrorMsg:
		m.launching = false
		m.launched = false
		m.playErr = msg.err
		return m, nil

	case clearLaunchedMsg:
		m.launched = false
		return m, nil

	case tea.KeyMsg:
		if m.filterActive {
			switch msg.String() {
			case "enter":
				m.filterActive = false
				m.filterInput.Blur()
				// If we have pending videos (batch series mode), fetch now with filter
				if len(m.pendingVideos) > 0 {
					f := ParseFilter(m.filterInput.Value())
					if f.IsEmpty() {
						return m, nil
					}
					m.loading = true
					m.err = nil
					nav := NavigateToAllStreamsMsg{Videos: m.pendingVideos, Type: m.pendingType}
					return m, tea.Batch(m.spinner.Tick, m.LoadAllStreams(nav, f))
				}
				m.applyFilter()
				return m, nil
			case "esc":
				m.filterActive = false
				m.filterInput.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "/":
			m.filterActive = true
			return m, m.filterInput.Focus()
		case "c":
			m.filterInput.SetValue("")
			m.applyFilter()
			return m, nil
		case "enter":
			if m.launching {
				return m, nil
			}
			if item, ok := m.list.SelectedItem().(streamItem); ok {
				m.launching = true
				m.playErr = nil
				url := item.stream.PlayableURL()
				videoID := item.videoID
				contentType := m.contentType
				if videoID == "" {
					videoID = m.contentID
				}
				return m, func() tea.Msg {
					err := player.PlayWithMPV(url)
					if err != nil {
						return mpvErrorMsg{err: err}
					}
					return mpvLaunchedMsg{videoID: videoID, videoType: contentType}
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

	var sections []string
	sections = append(sections, m.filterInput.View())

	// If waiting for filter input in batch mode, show hint
	if len(m.pendingVideos) > 0 && len(m.allItems) == 0 {
		sections = append(sections, HelpStyle.Render("Type a filter and press enter to search all episodes"))
		sections = append(sections, HelpStyle.Render("/ filter • esc: back"))
		return lipgloss.JoinVertical(lipgloss.Left, sections...)
	}

	view := m.list.View()
	if m.launching {
		view += "\n" + SubtitleStyle.Render("▶ Launching mpv...")
	} else if m.launched {
		view += "\n" + SubtitleStyle.Render("▶ Launched")
	}
	if m.playErr != nil {
		view += "\n" + ErrorStyle.Render(fmt.Sprintf("MPV error: %v", m.playErr))
	}
	sections = append(sections, view)
	sections = append(sections, HelpStyle.Render("/ filter • c clear • enter: play • esc: back"))
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
