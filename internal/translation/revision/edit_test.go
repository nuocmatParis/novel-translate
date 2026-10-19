package revision

import (
	"os"
	"testing"

	"github.com/voocel/ainovel-cli/internal/translation/domain"
	"github.com/voocel/ainovel-cli/internal/translation/skills"
	"github.com/voocel/ainovel-cli/internal/translation/workspace"
)

func TestRecoverCommittedViewAndPreserveConcurrentEdit(t *testing.T) {
	c, e := workspace.Create(t.TempDir(), "test", "Test")
	if e != nil {
		t.Fatal(e)
	}
	v := workspace.Volume{Directory: "volume"}
	ch := domain.Chapter{ID: "c001", Title: "Chapter", BlockIDs: []string{"one"}}
	book := domain.Book{Chapters: []domain.Chapter{ch}}
	g := domain.Generation{Units: map[string]domain.Unit{"one": {ID: "one", Target: "Trước"}}}
	if e = CaptureViews(c, v, book, &g); e != nil {
		t.Fatal(e)
	}
	if e = Recover(c, v, book, g); e != nil {
		t.Fatal(e)
	}
	if e = CaptureViews(c, v, book, &g); e != nil {
		t.Fatal(e)
	}
	u := g.Units["one"]
	u.Target = "Sau"
	g.Units["one"] = u
	// Simulate committed generation with old view still on disk.
	if e = Recover(c, v, book, g); e != nil {
		t.Fatal(e)
	}
	p, _ := ChapterPath(c, v, ch.ID)
	raw, _ := os.ReadFile(p)
	if string(raw) != Render(ch, g) {
		t.Fatal("crash replay failed")
	}
	if e = CaptureViews(c, v, book, &g); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(p, []byte("Independent manual edit"), 0600); e != nil {
		t.Fatal(e)
	}
	u.Target = "New generation"
	g.Units["one"] = u
	if e = Recover(c, v, book, g); e == nil {
		t.Fatal("manual edit overwritten")
	}
	raw, _ = os.ReadFile(p)
	if string(raw) != "Independent manual edit" {
		t.Fatal("manual data lost")
	}
}

func TestMarkersAndImpactPreserveManual(t *testing.T) {
	ch := domain.Chapter{Title: "Chapter", BlockIDs: []string{"one", "two"}}
	g := domain.Generation{ID: "g", SourceHash: "source", Units: map[string]domain.Unit{"one": {ID: "one", Target: "Một", Origin: "manual", Dependencies: domain.Dependencies{Skills: "old"}}, "two": {ID: "two", Target: "Hai", Dependencies: domain.Dependencies{Skills: "old"}}}}
	text := Render(ch, g)
	parsed, e := parse(text, ch.BlockIDs)
	if e != nil || parsed["one"] != "Một" {
		t.Fatal(parsed, e)
	}
	if _, e = parse(text, []string{"two", "one"}); e == nil {
		t.Fatal("reorder accepted")
	}
	book := domain.Book{SourceHash: "source", Blocks: []domain.Block{{ID: "one", SourceHash: "", Sequence: 0}, {ID: "two", SourceHash: "", Sequence: 1}}}
	p := Plan(workspace.Volume{ID: "v1"}, book, g, skills.Snapshot{Digest: "new"})
	if len(p.Units) != 2 || p.Units[0].Action != "review" {
		t.Fatal(p)
	}
	if e = Apply(&g, p); e != nil {
		t.Fatal(e)
	}
	if g.Units["one"].Target != "Một" {
		t.Fatal("manual overwritten")
	}
	if _, ok := g.Units["two"]; ok {
		t.Fatal("affected AI unit not invalidated")
	}
}
