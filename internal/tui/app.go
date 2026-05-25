package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/soakhan/cremio/internal/config"
	"github.com/soakhan/cremio/internal/stremio"
)

type Screen int

const (
	ScreenHome Screen = iota
	ScreenSearch
	ScreenAddons
	ScreenDetail
	ScreenStreams
)

type App struct {
	screen     Screen
	prevScreen Screen
	width      int
	height     int

	client *stremio.Client
	config *config.Config

	home    HomeModel
	search  SearchModel
	addons  AddonsModel
	detail  DetailModel
	streams StreamsModel
}

func NewApp(cfg *config.Config) App {
	client := stremio.NewClient()
	return App{
		screen:  ScreenHome,
		client:  client,
		config:  cfg,
		home:    NewHomeModel(client, cfg),
		search:  NewSearchModel(client, cfg),
		addons:  NewAddonsModel(client, cfg),
		detail:  NewDetailModel(client, cfg),
		streams: NewStreamsModel(client, cfg),
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(a.home.Init(), a.addons.Init())
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		contentHeight := msg.Height - 4 // tab bar + padding
		a.home.SetSize(msg.Width-4, contentHeight)
		a.search.SetSize(msg.Width-4, contentHeight)
		a.addons.SetSize(msg.Width-4, contentHeight)
		a.detail.SetSize(msg.Width-4, contentHeight)
		a.streams.SetSize(msg.Width-4, contentHeight)
		return a, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if a.screen == ScreenAddons && a.addons.inputActive {
				break
			}
			if a.screen == ScreenSearch && a.search.inputFocused {
				break
			}
			if a.screen == ScreenStreams && a.streams.filterActive {
				break
			}
			return a, tea.Quit
		case "esc", "escape":
			if a.screen == ScreenStreams && a.streams.filterActive {
				break
			}
			if a.screen == ScreenDetail {
				a.screen = a.prevScreen
				return a, nil
			}
			if a.screen == ScreenStreams {
				a.screen = ScreenDetail
				return a, nil
			}
		case "tab":
			if a.screen == ScreenDetail || a.screen == ScreenStreams {
				break
			}
			if a.screen == ScreenAddons && a.addons.inputActive {
				break
			}
			if a.screen == ScreenSearch && a.search.inputFocused {
				break
			}
			switch a.screen {
			case ScreenHome:
				a.screen = ScreenSearch
				return a, nil
			case ScreenSearch:
				a.screen = ScreenAddons
				return a, nil
			case ScreenAddons:
				a.screen = ScreenHome
				return a, a.home.Init()
			}
			return a, nil
		}

	case NavigateToDetailMsg:
		a.prevScreen = a.screen
		a.screen = ScreenDetail
		a.detail.loading = true
		a.detail.meta = nil
		a.detail.err = nil
		return a, tea.Batch(a.detail.spinner.Tick, a.detail.LoadMeta(msg))

	case NavigateToStreamsMsg:
		a.screen = ScreenStreams
		a.streams.loading = true
		a.streams.err = nil
		a.streams.allItems = nil
		a.streams.filterInput.SetValue("")
		a.streams.list.SetItems(nil)
		return a, tea.Batch(a.streams.spinner.Tick, a.streams.LoadStreams(msg))

	case NavigateToAllStreamsMsg:
		a.screen = ScreenStreams
		a.streams.loading = true
		a.streams.err = nil
		a.streams.allItems = nil
		a.streams.filterInput.SetValue("")
		a.streams.list.SetItems(nil)
		return a, tea.Batch(a.streams.spinner.Tick, a.streams.LoadAllStreams(msg))

	case AddonAddedMsg:
		// Refresh home catalogs when addon is added
		a.home = NewHomeModel(a.client, a.config)
		a.search = NewSearchModel(a.client, a.config)
		return a, nil

	case AddonRemovedMsg:
		a.home = NewHomeModel(a.client, a.config)
		a.search = NewSearchModel(a.client, a.config)
		return a, nil

	case addonsRefreshedMsg:
		a.addons, _ = a.addons.Update(msg)
		return a, nil
	}

	var cmd tea.Cmd
	switch a.screen {
	case ScreenHome:
		a.home, cmd = a.home.Update(msg)
	case ScreenSearch:
		a.search, cmd = a.search.Update(msg)
	case ScreenAddons:
		a.addons, cmd = a.addons.Update(msg)
	case ScreenDetail:
		// Don't pass esc to sub-model
		if keyMsg, ok := msg.(tea.KeyMsg); ok && (keyMsg.String() == "esc" || keyMsg.String() == "escape") {
			a.screen = a.prevScreen
			return a, nil
		}
		a.detail, cmd = a.detail.Update(msg)
	case ScreenStreams:
		a.streams, cmd = a.streams.Update(msg)
	}
	return a, cmd
}

func (a App) View() string {
	tabs := a.renderTabs()
	var content string
	switch a.screen {
	case ScreenHome:
		content = a.home.View()
	case ScreenSearch:
		content = a.search.View()
	case ScreenAddons:
		content = a.addons.View()
	case ScreenDetail:
		content = a.detail.View()
	case ScreenStreams:
		content = a.streams.View()
	}
	return AppStyle.Render(lipgloss.JoinVertical(lipgloss.Left, tabs, content))
}

func (a App) renderTabs() string {
	tabs := []struct {
		name   string
		screen Screen
	}{
		{"Home", ScreenHome},
		{"Search", ScreenSearch},
		{"Addons", ScreenAddons},
	}

	var rendered []string
	for _, t := range tabs {
		if t.screen == a.screen || (a.screen == ScreenDetail && t.screen == a.prevScreen) || (a.screen == ScreenStreams && t.screen == a.prevScreen) {
			rendered = append(rendered, TabActiveStyle.Render(t.name))
		} else {
			rendered = append(rendered, TabInactiveStyle.Render(t.name))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// Navigation messages
type NavigateToDetailMsg struct {
	ID      string
	Type    string
	BaseURL string
}

type NavigateToStreamsMsg struct {
	ID   string
	Type string
}

type NavigateToAllStreamsMsg struct {
	Videos []stremio.Video
	Type   string
}

type AddonAddedMsg struct{}
type AddonRemovedMsg struct{}
