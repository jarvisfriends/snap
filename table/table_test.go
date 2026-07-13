package table

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/jarvisfriends/snap/styles"
)

func sampleCols() []Column {
	return []Column{{Title: "Name", Filter: true}, {Title: "N"}}
}

func sampleRows() []Row {
	return []Row{
		{Key: "a", Cells: []Cell{Text("Apple"), Num("3", 3)}},
		{Key: "b", Cells: []Cell{Text("Banana"), Num("1", 1)}},
		{Key: "c", Cells: []Cell{Text("Cherry"), Num("2", 2)}},
	}
}

func keyText(s string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
}

// TestNumericSortDesc verifies numeric cells sort by magnitude (not lexically)
// and that the default descending sort orders highest-first.
func TestNumericSortDesc(t *testing.T) {
	m := New(sampleCols(), WithSort(1, false))
	m.SetRows(sampleRows())
	if got := []string{
		m.rows[0].Key,
		m.rows[1].Key,
		m.rows[2].Key,
	}; got[0] != "a" || got[1] != "c" ||
		got[2] != "b" {
		t.Fatalf("desc numeric sort wrong: %v", got)
	}
	if r, ok := m.SelectedRow(); !ok || r.Key != "a" {
		t.Fatalf("selected row should be the top (a): %+v ok=%v", r, ok)
	}
}

// TestSortByColCycle verifies the 3-state header cycle: asc → desc → unsorted.
func TestSortByColCycle(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())

	m.sortByCol(1)
	if !m.sortActive || !m.sortAsc {
		t.Fatalf("first click should sort ascending")
	}
	m.sortByCol(1)
	if !m.sortActive || m.sortAsc {
		t.Fatalf("second click should sort descending")
	}
	m.sortByCol(1)
	if m.sortActive {
		t.Fatalf("third click should clear the sort")
	}
}

// TestFilterThroughKeys drives the `/` filter through the real input path:
// `/` focuses the filter input, typed text narrows the rows live, Enter blurs
// the input keeping the filter applied, and esc (blurred) clears it.
func TestFilterThroughKeys(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())

	m.Update(keyText("/"))
	if !m.Filtering() {
		t.Fatal("'/' should focus the filter input")
	}
	m.Update(keyText("a"))
	m.Update(keyText("n"))
	if got := m.bt.GetCurrentFilter(); got != "an" {
		t.Fatalf("filter text = %q, want %q", got, "an")
	}
	if got := m.bt.TotalRows(); got != 1 {
		t.Fatalf("filter 'an' should match 1 row, got %d", got)
	}
	if r, ok := m.SelectedRow(); !ok || r.Key != "b" {
		t.Fatalf("filtered selection should be Banana, got %q ok=%v", r.Key, ok)
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if m.Filtering() {
		t.Fatal("Enter should blur the filter input")
	}
	if got := m.bt.TotalRows(); got != 1 {
		t.Fatalf("filter should stay applied after blur, got %d rows", got)
	}

	m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if got := m.bt.TotalRows(); got != 3 {
		t.Fatalf("esc should clear the filter, got %d rows", got)
	}
}

// TestFilterOnlyMatchesFilterColumns: column N is not filterable, so its cell
// text never matches (filtering "2" matches nothing even though Cherry's N=2).
func TestFilterOnlyMatchesFilterColumns(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())
	m.Update(keyText("/"))
	m.Update(keyText("2"))
	if got := m.bt.TotalRows(); got != 0 {
		t.Fatalf("filter should only match Filter:true columns, got %d rows", got)
	}
}

// TestOpenEmitsMsg verifies Enter emits an OpenDetailMsg for the selected row.
func TestOpenEmitsMsg(t *testing.T) {
	m := New(sampleCols(), WithSort(1, false))
	m.SetRows(sampleRows())
	cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should return a command")
	}
	msg, ok := cmd().(OpenDetailMsg)
	if !ok {
		t.Fatalf("expected OpenDetailMsg, got %T", cmd())
	}
	if msg.Key != "a" {
		t.Fatalf("OpenDetailMsg should carry the selected key (a), got %q", msg.Key)
	}
}

// TestKeyboardSortKey drives the sort key through Update (the real input path),
// constructing the KeyPressMsg the way the terminal does (Text set), and checks
// the sort actually changes.
func TestKeyboardSortKey(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())
	if m.sortActive {
		t.Fatal("table should start unsorted")
	}
	if cmd := m.Update(keyText("s")); cmd != nil {
		t.Errorf("sort key should not return a command, got %T", cmd)
	}
	if !m.sortActive || m.sortCol != 0 || !m.sortAsc {
		t.Fatalf(
			"'s' should sort column 0 ascending; got active=%v col=%d asc=%v",
			m.sortActive,
			m.sortCol,
			m.sortAsc,
		)
	}
	m.Update(keyText("s"))
	if !m.sortActive || m.sortAsc {
		t.Fatalf(
			"second 's' should flip to descending; got active=%v asc=%v",
			m.sortActive,
			m.sortAsc,
		)
	}
}

// TestKeyboardSortReordersRows confirms a keyboard sort actually reorders the
// underlying rows, not just the sort flags.
func TestKeyboardSortReordersRows(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())
	m.Update(keyText("s")) // sort column 0 (Name) ascending
	if m.rows[0].Key != "a" || m.rows[2].Key != "c" {
		t.Fatalf(
			"ascending name sort wrong order: %s,%s,%s",
			m.rows[0].Key,
			m.rows[1].Key,
			m.rows[2].Key,
		)
	}
}

// TestMouseHeaderSort clicks a column header (after a render records geometry)
// and checks the 3-state cycle runs.
func TestMouseHeaderSort(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())
	m.SetSize(60, 20)
	_ = m.View(styles.Active(), 1) // records headerY + column widths

	if len(m.colWidths) != len(sampleCols()) {
		t.Fatal("no column widths recorded; cannot locate header columns")
	}
	x := m.colWidths[0] + 1 // a point inside column 1

	m.HandleClick(x, m.headerY)
	if !m.sortActive || m.sortCol != 1 || !m.sortAsc {
		t.Fatalf(
			"header click should sort col 1 ascending; got active=%v col=%d asc=%v",
			m.sortActive,
			m.sortCol,
			m.sortAsc,
		)
	}
	m.HandleClick(x, m.headerY)
	if !m.sortActive || m.sortAsc {
		t.Fatalf(
			"second header click should be descending; got active=%v asc=%v",
			m.sortActive,
			m.sortAsc,
		)
	}
	m.HandleClick(x, m.headerY)
	if m.sortActive {
		t.Fatal("third header click should clear the sort")
	}
}

// TestMouseClickSelects verifies a click on a data row moves the selection
// there.
func TestMouseClickSelects(t *testing.T) {
	m := New(sampleCols(), WithSort(1, false))
	m.SetRows(sampleRows())
	m.SetSize(60, 20)
	_ = m.View(styles.Active(), 0)

	if cmd := m.HandleClick(3, m.dataStartY+1); cmd != nil {
		t.Fatal("a single click should select, not open details")
	}
	if r, ok := m.SelectedRow(); !ok || r.Key != "c" {
		t.Fatalf("click on second data row should select c, got %q ok=%v", r.Key, ok)
	}
}

// TestMouseDoubleClickOpens checks a quick second click on a row opens details.
func TestMouseDoubleClickOpens(t *testing.T) {
	m := New(sampleCols(), WithSort(1, false))
	m.SetRows(sampleRows())
	m.SetSize(60, 20)
	_ = m.View(styles.Active(), 0)

	y := m.dataStartY // first data row
	if cmd := m.HandleClick(3, y); cmd != nil {
		t.Fatal("a single click should select, not open details")
	}
	cmd := m.HandleClick(3, y) // immediate second click → double-click
	if cmd == nil {
		t.Fatal("double-click should open details")
	}
	if _, ok := cmd().(OpenDetailMsg); !ok {
		t.Fatalf("double-click should emit OpenDetailMsg, got %T", cmd())
	}
}

// TestWheelMovesAndClamps: the wheel moves the selection one row per notch
// and clamps at both ends instead of wrapping.
func TestWheelMovesAndClamps(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())

	m.HandleWheel(false) // down 1 from row 0
	if got := m.bt.GetHighlightedRowIndex(); got != 1 {
		t.Fatalf("wheel down should move one row, got %d", got)
	}
	m.HandleWheel(false)
	m.HandleWheel(false) // past the last row → clamped to index 2
	if got := m.bt.GetHighlightedRowIndex(); got != 2 {
		t.Fatalf("wheel down should clamp to last row, got %d", got)
	}
	for range 4 { // back past the first row → clamped to 0
		m.HandleWheel(true)
	}
	if got := m.bt.GetHighlightedRowIndex(); got != 0 {
		t.Fatalf("wheel up should clamp to first row, got %d", got)
	}
}

// TestBorderlessCompactRender pins the compact layout HandleClick's geometry
// relies on: the header is the first line, data rows follow immediately (no
// border or separator lines), and the footer is the last line.
func TestBorderlessCompactRender(t *testing.T) {
	m := New(sampleCols())
	m.SetRows(sampleRows())
	m.SetSize(40, 20)

	out := m.View(styles.Active(), 0)
	lines := strings.Split(ansi.Strip(out), "\n")
	if len(lines) != 1+len(sampleRows())+1 {
		t.Fatalf("want header + %d rows + footer lines, got %d:\n%s",
			len(sampleRows()), len(lines), ansi.Strip(out))
	}
	for _, glyph := range []string{"│", "─", "┌", "╭", "┼", "┬"} {
		if strings.Contains(lines[0]+lines[1], glyph) {
			t.Fatalf("borderless render should not contain %q:\n%s", glyph, ansi.Strip(out))
		}
	}
	if !strings.Contains(lines[m.headerY], "Name") {
		t.Errorf("headerY=%d does not land on the header row: %q", m.headerY, lines[m.headerY])
	}
	if !strings.Contains(lines[m.dataStartY], "Apple") {
		t.Errorf("dataStartY=%d does not land on the first data row: %q", m.dataStartY, lines[m.dataStartY])
	}
}

// TestViewRecordsGeometry: headerY/dataStartY are page-relative offsets from
// the originY the host passes in.
func TestViewRecordsGeometry(t *testing.T) {
	m := New(sampleCols(), WithSort(1, false))
	m.SetRows(sampleRows())
	m.SetSize(60, 20)

	out := m.View(styles.Active(), 5)
	if out == "" {
		t.Fatal("View returned empty output")
	}
	if m.headerY != 5 {
		t.Errorf("headerY = %d, want 5 (originY)", m.headerY)
	}
	if m.dataStartY != 6 {
		t.Errorf("dataStartY = %d, want 6 (originY + header line)", m.dataStartY)
	}
}

// TestColumnWidthsAlignWithRender confirms the widths columnAtX hit-tests with
// are the widths the render actually uses: the second column's title starts
// exactly one padding cell past the first column's width.
func TestColumnWidthsAlignWithRender(t *testing.T) {
	m := New([]Column{{Title: "Alpha"}, {Title: "Beta"}})
	m.SetRows([]Row{{Key: "x", Cells: []Cell{Text("one"), Text("two")}}})
	m.SetSize(30, 10)

	out := m.View(styles.Active(), 0)
	header := ansi.Strip(strings.Split(out, "\n")[0])
	wantX := m.colWidths[0] + 1 // second cell's 1-cell padding
	if got := strings.Index(header, "Beta"); got != wantX {
		t.Fatalf("Beta starts at x=%d, want %d (colWidths=%v): %q",
			got, wantX, m.colWidths, header)
	}
	if m.columnAtX(wantX) != 1 {
		t.Fatalf("columnAtX(%d) = %d, want 1", wantX, m.columnAtX(wantX))
	}
}

// TestFitWidthsStretchAndSqueeze: widths stretch round-robin to fill the total
// and squeeze the widest column first, never below the minimum.
func TestFitWidthsStretchAndSqueeze(t *testing.T) {
	if got := fitWidths([]int{5, 5}, 14); got[0]+got[1] != 14 {
		t.Fatalf("stretch: %v does not fill 14", got)
	}
	got := fitWidths([]int{20, 6}, 16)
	if got[0]+got[1] != 16 {
		t.Fatalf("squeeze: %v does not fit 16", got)
	}
	if got[1] < minColWidth {
		t.Fatalf("squeeze took the narrow column below min: %v", got)
	}
	// Unknown width (0) leaves natural widths untouched.
	if got := fitWidths([]int{7, 9}, 0); got[0] != 7 || got[1] != 9 {
		t.Fatalf("width 0 should keep natural widths, got %v", got)
	}
}

// TestEmptyTableIsSafe: no rows renders header + footer only, and clicks and
// selection queries are harmless no-ops.
func TestEmptyTableIsSafe(t *testing.T) {
	m := New(sampleCols())
	m.SetSize(40, 10)
	out := m.View(styles.Active(), 0)
	if !strings.Contains(ansi.Strip(out), "0/0") {
		t.Errorf("empty table footer should show 0/0: %q", ansi.Strip(out))
	}
	if _, ok := m.SelectedRow(); ok {
		t.Error("empty table should have no selected row")
	}
	if cmd := m.HandleClick(2, m.dataStartY); cmd != nil {
		t.Error("click below an empty table should be a no-op")
	}
	if cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter}); cmd != nil {
		t.Error("Enter on an empty table should be a no-op")
	}
}
