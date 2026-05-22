package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/soakhan/cremio/internal/config"
	"github.com/soakhan/cremio/internal/stremio"
)

type videoItem struct {
	video stremio.Video
}

func (v videoItem) Title() string {
	if v.video.Season > 0 {
		return fmt.Sprintf("S%02dE%02d - %s", v.video.Season, v.video.Episode, v.video.Title)
	}
	return v.video.Title
}
func (v videoItem) Description() string { return v.video.Overview }
func (v videoItem) FilterValue() string { return v.video.Title }

type DetailModel struct {
	meta        *stremio.Meta
	list        list.Model
	spinner     spinner.Model
	client      *stremio.Client
	config      *config.Config
	loading     bool
	err         error
	width       int
	height      int
	contentType string
}

type metaLoadedMsg struct {
	meta *stremio.Meta
}
type metaErrorMsg struct {
	err error
}

func NewDetailModel(client *stremio.Client, cfg *config.Config) DetailModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Episodes"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	return DetailModel{
		list:    l,
		spinner: s,
		client:  client,
		config:  cfg,
	}
}

func (m *DetailModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.list.SetSize(w, h-10)
}

func (m DetailModel) LoadMeta(nav NavigateToDetailMsg) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// Try the source addon first, then all others
		tried := make(map[string]bool)
		addons := []string{nav.BaseURL}
		for _, u := range m.config.Addons {
			if u != nav.BaseURL {
				addons = append(addons, u)
			}
		}

		var lastErr error
		for _, addonURL := range addons {
			if tried[addonURL] {
				continue
			}
			tried[addonURL] = true
			resp, err := m.client.FetchMeta(ctx, addonURL, nav.Type, nav.ID)
			if err != nil {
				lastErr = err
				continue
			}
			if resp.Meta.ID == "" && resp.Meta.Name == "" {
				lastErr = fmt.Errorf("empty meta response")
				continue
			}
			// Skip addon error responses (e.g. AIOStreams error meta)
			if strings.HasPrefix(resp.Meta.ID, "aiostreamserror") {
				lastErr = fmt.Errorf("%s", resp.Meta.Description)
				continue
			}
			return metaLoadedMsg{meta: &resp.Meta}
		}

		if lastErr != nil {
			return metaErrorMsg{err: fmt.Errorf("could not load metadata: %w", lastErr)}
		}
		return metaErrorMsg{err: fmt.Errorf("no addon supports meta for this item")}
	}
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case metaLoadedMsg:
		m.loading = false
		m.meta = msg.meta
		m.contentType = msg.meta.Type

		if len(msg.meta.Videos) > 0 {
			items := make([]list.Item, len(msg.meta.Videos))
			for i, v := range msg.meta.Videos {
				items[i] = videoItem{video: v}
			}
			m.list.SetItems(items)
		}
		return m, nil

	case metaErrorMsg:
		m.loading = false
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.meta == nil {
				return m, nil
			}
			// For movies (no videos or single video), go to streams directly
			if m.meta.Type == "movie" || len(m.meta.Videos) == 0 {
				return m, func() tea.Msg {
					return NavigateToStreamsMsg{
						ID:   m.meta.ID,
						Type: m.meta.Type,
					}
				}
			}
			// For series, use selected episode's video ID
			if item, ok := m.list.SelectedItem().(videoItem); ok {
				return m, func() tea.Msg {
					return NavigateToStreamsMsg{
						ID:   item.video.ID,
						Type: m.meta.Type,
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m DetailModel) View() string {
	if m.loading {
		return m.spinner.View() + " Loading details..."
	}
	if m.err != nil {
		return ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}
	if m.meta == nil {
		return ""
	}

	var sections []string

	// Title
	sections = append(sections, TitleStyle.Render(m.meta.Name))

	// Info line
	var info []string
	if m.meta.ReleaseInfo != "" {
		info = append(info, m.meta.ReleaseInfo)
	}
	if m.meta.Runtime != "" {
		info = append(info, m.meta.Runtime)
	}
	if m.meta.IMDBRating != "" {
		info = append(info, "⭐ "+m.meta.IMDBRating)
	}
	if len(info) > 0 {
		sections = append(sections, SubtitleStyle.Render(strings.Join(info, " • ")))
	}

	// Genres
	if len(m.meta.Genres) > 0 {
		sections = append(sections, DetailLabelStyle.Render("Genres: ")+DetailValueStyle.Render(strings.Join(m.meta.Genres, ", ")))
	}

	// Cast
	if len(m.meta.Cast) > 0 {
		cast := m.meta.Cast
		if len(cast) > 5 {
			cast = cast[:5]
		}
		sections = append(sections, DetailLabelStyle.Render("Cast: ")+DetailValueStyle.Render(strings.Join(cast, ", ")))
	}

	// Description
	if m.meta.Description != "" {
		desc := m.meta.Description
		if len(desc) > 300 {
			desc = desc[:300] + "..."
		}
		sections = append(sections, "")
		sections = append(sections, DetailValueStyle.Render(desc))
	}

	sections = append(sections, "")

	// For movies, show prompt to play
	if m.meta.Type == "movie" || len(m.meta.Videos) == 0 {
		sections = append(sections, HelpStyle.Render("enter: find streams • esc: back"))
	} else {
		// For series, show episode list
		sections = append(sections, m.list.View())
		sections = append(sections, HelpStyle.Render("enter: find streams for episode • esc: back"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
