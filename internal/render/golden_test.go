package render

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/veteranbv/sysgreet/internal/config"
	"github.com/veteranbv/sysgreet/internal/terminal"
)

var update = flag.Bool("update", false, "rewrite golden files")

// TestGoldenLayouts pins the exact banner body at the widths that matter:
// piped (stacked), a standard 80-column terminal, and a wide one. Run
// `go test ./internal/render -run Golden -update` after an intended change.
func TestGoldenLayouts(t *testing.T) {
	for _, width := range []int{0, 80, 140} {
		name := filepath.Join("testdata", "demo_"+widthName(width)+".golden")
		got := NewRenderer(terminal.Env{Width: width}).Render(demoOutput(), config.Default()) + "\n"
		if *update {
			if err := os.WriteFile(name, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("missing golden file (run with -update): %v", err)
		}
		if got != string(want) {
			t.Errorf("%s changed; rerun with -update if intended.\n--- got ---\n%s--- want ---\n%s", name, got, want)
		}
	}
}

func widthName(w int) string {
	if w == 0 {
		return "piped"
	}
	return map[int]string{80: "80", 140: "140"}[w]
}
