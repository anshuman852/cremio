package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/itssoap/cremio/internal/config"
	"github.com/itssoap/cremio/internal/stremio"
)

type SearchModel struct {
	input        textinput.Model
	results      list.Model
	spinner      spinner.Model
	client       *stremio.Client
	config       *config.Config
	inputFocused bool
	searching    bool
	err          error
	width        int
	height       int
}

type searchResultsMsg struct {
	items []catalogItem
}
type searchErrorMsg struct {
	err error
}

func NewSearchModel(client *stremio.Client, cfg *config.Config) SearchModel {
	ti := textinput.New()
	ti.Placeholder = "Search movies & series..."
	ti.CharLimit = 100

	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Results"
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))

	return SearchModel{
		input:   ti,
		results: l,
		spinner: s,
		client:  client,
		config:  cfg,
	}
}

func (m *SearchModel) SetSize(w, h int) {
	m.width = w
	m.height = h
	m.input.Width = w - 4
	m.results.SetSize(w, h-4)
}

func (m SearchModel) search(query string) tea.Cmd {
	return func() tea.Msg {
		var allItems []catalogItem
		ctx := context.Background()
		queryLower := strings.ToLower(query)

		for _, addonURL := range m.config.Addons {
			manifest, err := m.client.FetchManifest(ctx, addonURL)
			if err != nil {
				continue
			}

			for _, cat := range manifest.Catalogs {
				if cat.SupportsSearch() {
					// Use the addon's search endpoint
					resp, err := m.client.SearchCatalog(ctx, addonURL, cat.Type, cat.ID, query)
					if err != nil {
						continue
					}
					for _, meta := range resp.Metas {
						allItems = append(allItems, catalogItem{meta: meta, baseURL: addonURL})
					}
				} else {
					// Skip catalogs that require extra params we can't provide
					hasRequired := false
					for _, e := range cat.Extra {
						if e.IsRequired {
							hasRequired = true
							break
						}
					}
					if hasRequired {
						continue
					}
					// Fallback: fetch catalog and filter client-side
					resp, err := m.client.FetchCatalog(ctx, addonURL, cat.Type, cat.ID)
					if err != nil {
						continue
					}
					for _, meta := range resp.Metas {
						if strings.Contains(strings.ToLower(meta.Name), queryLower) {
							allItems = append(allItems, catalogItem{meta: meta, baseURL: addonURL})
						}
					}
				}
			}
		}

		if len(allItems) == 0 {
			return searchErrorMsg{err: fmt.Errorf("no results found")}
		}
		return searchResultsMsg{items: allItems}
	}
}

func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	switch msg := msg.(type) {
	case searchResultsMsg:
		m.searching = false
		items := make([]list.Item, len(msg.items))
		for i, item := range msg.items {
			items[i] = item
		}
		m.results.SetItems(items)
		return m, nil

	case searchErrorMsg:
		m.searching = false
		m.err = msg.err
		return m, nil

	case spinner.TickMsg:
		if m.searching {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if m.inputFocused {
			switch msg.String() {
			case "enter":
				query := m.input.Value()
				if query != "" {
					m.inputFocused = false
					m.input.Blur()
					m.searching = true
					m.err = nil
					return m, tea.Batch(m.spinner.Tick, m.search(query))
				}
				return m, nil
			case "esc":
				m.inputFocused = false
				m.input.Blur()
				return m, nil
			}
		} else {
			switch msg.String() {
			case "/":
				m.inputFocused = true
				return m, m.input.Focus()
			case "enter":
				if item, ok := m.results.SelectedItem().(catalogItem); ok {
					return m, func() tea.Msg {
						return NavigateToDetailMsg{
							ID:      item.meta.ID,
							Type:    item.meta.Type,
							BaseURL: item.baseURL,
						}
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	if m.inputFocused {
		m.input, cmd = m.input.Update(msg)
	} else {
		m.results, cmd = m.results.Update(msg)
	}
	return m, cmd
}

func (m SearchModel) View() string {
	var sections []string

	sections = append(sections, m.input.View())

	if m.searching {
		sections = append(sections, "\n"+m.spinner.View()+" Searching...")
	} else if m.err != nil {
		sections = append(sections, "\n"+ErrorStyle.Render(m.err.Error()))
	} else {
		sections = append(sections, m.results.View())
	}

	help := HelpStyle.Render("/ focus search • enter submit • esc unfocus")
	sections = append(sections, help)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
