package stremio

type Manifest struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Resources   []any     `json:"resources"`
	Types       []string  `json:"types"`
	Catalogs    []Catalog `json:"catalogs"`
	IDPrefixes  []string  `json:"idPrefixes,omitempty"`
}

type Catalog struct {
	Type  string       `json:"type"`
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Extra []CatalogExtra `json:"extra,omitempty"`
}

type CatalogExtra struct {
	Name       string   `json:"name"`
	IsRequired bool     `json:"isRequired,omitempty"`
	Options    []string `json:"options,omitempty"`
}

func (c Catalog) SupportsSearch() bool {
	for _, e := range c.Extra {
		if e.Name == "search" {
			return true
		}
	}
	return false
}

type CatalogResponse struct {
	Metas []MetaPreview `json:"metas"`
}

type MetaPreview struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Poster      string `json:"poster,omitempty"`
	PosterShape string `json:"posterShape,omitempty"`
	Description string `json:"description,omitempty"`
	ReleaseInfo string `json:"releaseInfo,omitempty"`
	IMDBRating  string `json:"imdbRating,omitempty"`
}

type MetaResponse struct {
	Meta Meta `json:"meta"`
}

type Meta struct {
	ID          string   `json:"id"`
	Type        string   `json:"type"`
	Name        string   `json:"name"`
	Poster      string   `json:"poster,omitempty"`
	Background  string   `json:"background,omitempty"`
	Description string   `json:"description,omitempty"`
	ReleaseInfo string   `json:"releaseInfo,omitempty"`
	IMDBRating  string   `json:"imdbRating,omitempty"`
	Runtime     string   `json:"runtime,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Director    []string `json:"director,omitempty"`
	Cast        []string `json:"cast,omitempty"`
	Videos      []Video  `json:"videos,omitempty"`
}

type Video struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Released  string `json:"released,omitempty"`
	Season    int    `json:"season,omitempty"`
	Episode   int    `json:"episode,omitempty"`
	Overview  string `json:"overview,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

type StreamResponse struct {
	Streams []Stream `json:"streams"`
}

type Stream struct {
	URL         string       `json:"url,omitempty"`
	YtID        string       `json:"ytId,omitempty"`
	InfoHash    string       `json:"infoHash,omitempty"`
	FileIdx     *int         `json:"fileIdx,omitempty"`
	ExternalURL string       `json:"externalUrl,omitempty"`
	Name        string       `json:"name,omitempty"`
	Title       string       `json:"title,omitempty"`
	Description string       `json:"description,omitempty"`
	BehaviorHints *StreamBehaviorHints `json:"behaviorHints,omitempty"`
}

type StreamBehaviorHints struct {
	NotWebReady      bool   `json:"notWebReady,omitempty"`
	BingeGroup       string `json:"bingeGroup,omitempty"`
	ProxyHeaders     *ProxyHeaders `json:"proxyHeaders,omitempty"`
}

type ProxyHeaders struct {
	Request  map[string]string `json:"request,omitempty"`
	Response map[string]string `json:"response,omitempty"`
}

func (s Stream) PlayableURL() string {
	if s.URL != "" {
		return s.URL
	}
	if s.InfoHash != "" {
		url := "magnet:?xt=urn:btih:" + s.InfoHash
		if s.FileIdx != nil {
			// mpv can't select file index from magnet, but we pass it anyway
		}
		return url
	}
	if s.YtID != "" {
		return "https://www.youtube.com/watch?v=" + s.YtID
	}
	if s.ExternalURL != "" {
		return s.ExternalURL
	}
	return ""
}

func (s Stream) DisplayName() string {
	if s.Name != "" {
		return s.Name
	}
	if s.Title != "" {
		return s.Title
	}
	if s.Description != "" {
		return s.Description
	}
	if s.URL != "" {
		return "HTTP Stream"
	}
	if s.InfoHash != "" {
		return "Torrent: " + s.InfoHash[:8] + "..."
	}
	return "Unknown Stream"
}
