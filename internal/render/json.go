package render

import (
	"encoding/json"

	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/config"
)

type jsonItem struct {
	Label  string   `json:"label"`
	Value  string   `json:"value"`
	Detail string   `json:"detail,omitempty"`
	Meter  *float64 `json:"meter,omitempty"`
	Level  string   `json:"level,omitempty"`
}

type jsonSection struct {
	Key   string         `json:"key"`
	Title string         `json:"title"`
	Items []jsonItem     `json:"items,omitempty"`
	Lines []string       `json:"lines"`
	Data  map[string]any `json:"data,omitempty"`
}

type jsonBanner struct {
	Hostname string        `json:"hostname"`
	Header   []string      `json:"header"`
	Sections []jsonSection `json:"sections"`
}

// RenderJSON emits the banner as structured JSON for scripting. Sections
// follow the configured layout order, same as the text output.
func RenderJSON(out banner.Output, cfg config.Config) (string, error) {
	doc := jsonBanner{
		Hostname: out.Header.Hostname,
		Header:   out.Header.Lines,
		Sections: []jsonSection{},
	}
	if doc.Header == nil {
		doc.Header = []string{}
	}
	for _, section := range orderSections(out.Sections, cfg.Layout.Sections) {
		if len(section.Lines) == 0 {
			continue
		}
		js := jsonSection{
			Key:   section.Key,
			Title: section.Title,
			Lines: section.Lines,
			Data:  section.Data,
		}
		for _, it := range section.Items {
			ji := jsonItem{Label: it.Label, Value: it.Value, Detail: it.Detail, Meter: it.Meter}
			if it.Level != banner.LevelNormal {
				ji.Level = it.Level.String()
			}
			js.Items = append(js.Items, ji)
		}
		doc.Sections = append(doc.Sections, js)
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
