package render

import (
	"encoding/json"

	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/config"
)

type jsonSection struct {
	Key   string         `json:"key"`
	Title string         `json:"title"`
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
		doc.Sections = append(doc.Sections, jsonSection{
			Key:   section.Key,
			Title: section.Title,
			Lines: section.Lines,
			Data:  section.Data,
		})
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
