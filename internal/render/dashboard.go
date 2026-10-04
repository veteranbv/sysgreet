package render

import (
	"math"
	"strings"

	"github.com/veteranbv/sysgreet/internal/banner"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

const (
	columnGap   = 4 // spaces between side-by-side sections
	labelGap    = 2 // spaces between a label and its value
	meterWidth  = 10
	meterNarrow = 5
)

// blockLayout is the set of choices that decide how wide a section renders.
// fit walks it from roomiest to tightest until the section fits.
type blockLayout struct {
	meterW     int
	showDetail bool
}

var layoutLadder = []blockLayout{
	{meterW: meterWidth, showDetail: true},
	{meterW: meterWidth, showDetail: false},
	{meterW: meterNarrow, showDetail: false},
	{meterW: 0, showDetail: false},
}

type block struct {
	section banner.Section
	labelW  int // widest label
	valueW  int // widest value among meter rows, which are right-aligned
}

func newBlock(s banner.Section) block {
	b := block{section: s}
	for _, it := range s.Items {
		b.labelW = max(b.labelW, terminal.DisplayWidth(it.Label))
		if it.Meter != nil {
			b.valueW = max(b.valueW, terminal.DisplayWidth(it.Value))
		}
	}
	return b
}

// rowWidth is the plain display width of one item under layout l.
func (b block) rowWidth(it banner.Item, l blockLayout) int {
	w := len(bodyIndent) + b.labelW + labelGap
	if it.Meter != nil && l.meterW > 0 {
		w += l.meterW + 1 + b.valueW
	} else {
		w += terminal.DisplayWidth(it.Value)
	}
	if l.showDetail && it.Detail != "" {
		w += labelGap + terminal.DisplayWidth(it.Detail)
	}
	return w
}

func (b block) width(l blockLayout) int {
	w := terminal.DisplayWidth(b.section.Title)
	for _, it := range b.section.Items {
		w = max(w, b.rowWidth(it, l))
	}
	return w
}

// fit picks the roomiest layout whose width is within maxW. When even the
// tightest is too wide, it returns the tightest; lines are then clipped.
func (b block) fit(maxW int) blockLayout {
	if maxW <= 0 {
		return layoutLadder[0]
	}
	for _, l := range layoutLadder {
		if b.width(l) <= maxW {
			return l
		}
	}
	return layoutLadder[len(layoutLadder)-1]
}

// render styles the block under layout l, clipping plain text so every
// line stays within maxW (0 means unbounded).
func (r Renderer) renderBlock(b block, l blockLayout, maxW int) []string {
	title := b.section.Title
	if maxW > 0 {
		title = clipTo(title, maxW)
	}
	lines := []string{r.title(title)}
	for _, it := range b.section.Items {
		lines = append(lines, r.renderItem(b, it, l, maxW))
	}
	return lines
}

func (r Renderer) renderItem(b block, it banner.Item, l blockLayout, maxW int) string {
	label := padRight(it.Label, b.labelW)
	prefixW := len(bodyIndent) + b.labelW + labelGap
	if maxW > 0 && prefixW >= maxW {
		// Degenerate widths: the label alone does not fit.
		return clipTo(bodyIndent+it.Label, maxW)
	}

	var meter, value string
	if it.Meter != nil && l.meterW > 0 {
		meter = r.meter(*it.Meter, l.meterW, it.Level) + " "
		value = padLeft(it.Value, b.valueW)
		prefixW += l.meterW + 1
	} else {
		value = it.Value
	}

	detail := ""
	if l.showDetail && it.Detail != "" {
		detail = strings.Repeat(" ", labelGap) + it.Detail
	}

	if maxW > 0 {
		room := maxW - prefixW
		value = clipTo(value, room)
		detail = clipTo(detail, room-terminal.DisplayWidth(value))
	}

	var sb strings.Builder
	sb.WriteString(bodyIndent)
	sb.WriteString(r.label(label))
	sb.WriteString(strings.Repeat(" ", labelGap))
	sb.WriteString(meter)
	sb.WriteString(r.value(value, it.Level))
	if detail != "" {
		sb.WriteString(r.label(detail))
	}
	return strings.TrimRight(sb.String(), " ")
}

// meter draws a usage bar filled to fraction, colored by level.
func (r Renderer) meter(fraction float64, width int, level banner.Level) string {
	fraction = math.Max(0, math.Min(1, fraction))
	filled := int(math.Round(fraction * float64(width)))
	color := "green"
	switch level {
	case banner.LevelWarn:
		color = "yellow"
	case banner.LevelCrit:
		color = "red"
	}
	return r.colorizer.Wrap(color, strings.Repeat("█", filled)) +
		terminal.Dim(r.profile, strings.Repeat("░", width-filled))
}

func (r Renderer) title(text string) string {
	return terminal.Bold(r.profile, r.colorizer.Wrap(r.accent, text))
}

func (r Renderer) label(text string) string {
	return terminal.Dim(r.profile, text)
}

func (r Renderer) value(text string, level banner.Level) string {
	switch level {
	case banner.LevelWarn:
		return r.colorizer.Wrap("yellow", text)
	case banner.LevelCrit:
		return terminal.Bold(r.profile, r.colorizer.Wrap("red", text))
	default:
		return text
	}
}

// renderBody lays sections out left to right, wrapping to a new row when
// the next section would overflow the width. An unknown width (0) stacks
// every section, keeping piped output stable.
func (r Renderer) renderBody(sections []banner.Section) []string {
	blocks := make([]block, 0, len(sections))
	for _, s := range sections {
		if len(s.Items) > 0 {
			blocks = append(blocks, newBlock(s))
		}
	}

	var out []string
	for i, row := range packRows(blocks, r.width) {
		if i > 0 {
			out = append(out, "")
		}
		out = append(out, r.renderRow(row)...)
	}
	return out
}

func packRows(blocks []block, maxW int) [][]block {
	var rows [][]block
	var cur []block
	curW := 0
	for _, b := range blocks {
		w := b.width(layoutLadder[0])
		need := w
		if len(cur) > 0 {
			need += columnGap
		}
		if len(cur) > 0 && (maxW <= 0 || curW+need > maxW) {
			rows = append(rows, cur)
			cur, curW, need = nil, 0, w
		}
		cur = append(cur, b)
		curW += need
	}
	if len(cur) > 0 {
		rows = append(rows, cur)
	}
	return rows
}

func (r Renderer) renderRow(row []block) []string {
	if len(row) == 1 {
		b := row[0]
		return r.renderBlock(b, b.fit(r.width), r.width)
	}

	// Side-by-side sections were packed at their roomiest width, so they
	// fit as-is.
	cols := make([][]string, len(row))
	widths := make([]int, len(row))
	height := 0
	for i, b := range row {
		l := layoutLadder[0]
		cols[i] = r.renderBlock(b, l, 0)
		widths[i] = b.width(l)
		height = max(height, len(cols[i]))
	}

	lines := make([]string, height)
	for y := range lines {
		var sb strings.Builder
		for i, col := range cols {
			cell := ""
			if y < len(col) {
				cell = col[y]
			}
			sb.WriteString(cell)
			if i < len(cols)-1 {
				pad := widths[i] - terminal.DisplayWidth(terminal.Strip(cell)) + columnGap
				sb.WriteString(strings.Repeat(" ", pad))
			}
		}
		lines[y] = strings.TrimRight(sb.String(), " ")
	}
	return lines
}

func padRight(s string, w int) string {
	if gap := w - terminal.DisplayWidth(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

func padLeft(s string, w int) string {
	if gap := w - terminal.DisplayWidth(s); gap > 0 {
		return strings.Repeat(" ", gap) + s
	}
	return s
}

// clipTo truncates plain text to w display columns. Callers pass the room
// left under a positive width bound, so w <= 0 means nothing fits.
func clipTo(s string, w int) string {
	if w <= 0 {
		return ""
	}
	return terminal.Clip(s, w)
}
