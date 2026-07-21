// Package table is a themed, interactive data table widget for Bubble Tea apps.
//
// Rendering, pagination, highlighting, and live filtering are delegated to
// github.com/evertras/bubble-table — the de-facto standard Bubble Tea table —
// while this wrapper keeps the features it lacks: full mouse support (header
// clicks sort, row clicks select, double-click opens, wheel scrolls), a
// 3-state column sort (asc → desc → unsorted) that understands numeric cells
// whose display text differs from their sort value (a duration shown as
// "2h 8m" but sorted by milliseconds), theme-hook styling from a
// styles.AppStyle passed to View so the table recolors live, and a status
// footer. The table renders borderless and compact: one header line, then
// data rows.
//
// Mouse interaction is cooperative: the host routes mouse events to a page's
// View().OnMouse with page-relative coordinates, and the page forwards clicks
// to HandleClick / wheels to HandleWheel. Clicking a header sorts that column;
// double-clicking a row (or pressing the Open key) emits an OpenDetailMsg the
// host page can act on (e.g. open a detail overlay).
package table

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	btable "github.com/evertras/bubble-table/table"

	"github.com/jarvisfriends/snap/keys"
	"github.com/jarvisfriends/snap/styles"
)

// doubleClickWindow is how close two clicks on the same row must land to count
// as a double-click (which opens the row's details).
const doubleClickWindow = 450 * time.Millisecond

// rowDataKey is the reserved bubble-table row-data key carrying our Row.
// Column keys are decimal indices, so it can never collide.
const rowDataKey = "__row"

// chromeRows is the number of non-data lines View draws around the data rows:
// the header line and the external footer line.
const chromeRows = 2

// minColWidth is the narrowest a column may be squeezed to when the natural
// widths exceed the available width (room for padding plus a few characters).
const minColWidth = 5

// Column describes one column header.
type Column struct {
	Title  string
	Filter bool // included in the `/` text filter
}

// Cell is one rendered cell. Numeric cells carry a real value so a column sorts
// by magnitude rather than lexically (e.g. "9" before "10", "2h 8m" by ms).
type Cell struct {
	Text  string
	Num   float64
	IsNum bool
}

// Text returns a string cell. Num returns a numeric cell whose display text and
// sort value can differ (e.g. a duration shown as "2h 8m" but sorted by ms).
func Text(s string) Cell                 { return Cell{Text: s} }
func Num(display string, v float64) Cell { return Cell{Text: display, Num: v, IsNum: true} }

// Row is one data row plus a stable Key the host uses to identify it (e.g. when
// opening details).
type Row struct {
	Key   string
	Cells []Cell
}

// OpenDetailMsg is emitted when the user opens a row (Enter or double-click).
// The host page handles it (e.g. by opening a detail overlay for Key).
type OpenDetailMsg struct{ Key string }

var _ help.KeyMap = (*keys.AppKeyMap)(nil)

// TableModel is the table widget. Construct it with New.
type TableModel struct {
	KeyMap *keys.AppKeyMap

	// HideFooterHint suppresses the footer's right-aligned key-hint text
	// (the cursor/total, sort, and live-filter readouts stay). Hosts that
	// surface the table's keys in their own help/status bar set this so the
	// hints aren't shown twice.
	HideFooterHint bool

	// bt does the rendering, pagination, highlight, and filter input.
	// Sorting stays here in the wrapper: bubble-table compares raw data
	// values, so it cannot sort a Num cell by magnitude while displaying
	// different text. The wrapper orders rows and hands them over in final
	// display order.
	bt btable.Model

	cols []Column
	rows []Row // sorted in place by the active sort

	sortCol    int
	sortAsc    bool
	sortActive bool

	width         int
	pageSize      int
	fixedPageSize bool

	// Geometry recorded by View (page-content coordinates) for HandleClick.
	colWidths  []int
	headerY    int
	dataStartY int

	lastClickRow  int
	lastClickTime time.Time
}

// Option configures a Model at construction.
type Option func(*TableModel)

// WithKeyMap overrides the default key bindings.
func WithKeyMap(km *keys.AppKeyMap) Option { return func(m *TableModel) { m.KeyMap = km } }

// WithPageSize sets a fixed page size (otherwise it's derived from the height
// passed to SetSize).
func WithPageSize(n int) Option {
	return func(m *TableModel) { m.pageSize, m.fixedPageSize = n, true }
}

// WithSort sets the initial sort column and direction.
func WithSort(col int, asc bool) Option {
	return func(m *TableModel) {
		if col >= 0 {
			m.sortCol, m.sortAsc, m.sortActive = col, asc, true
		}
	}
}

// New builds a table for the given columns.
func New(cols []Column, opts ...Option) *TableModel {
	m := &TableModel{
		KeyMap:       keys.DefaultKeyMap(),
		cols:         cols,
		pageSize:     20,
		sortCol:      -1,
		lastClickRow: -1,
	}
	for _, o := range opts {
		o(m)
	}
	m.bt = btable.New(nil).
		Border(btable.Border{}).
		WithOuterBorder(false).
		WithFooterVisibility(false).
		Filtered(true).
		Focused(true).
		WithPageSize(m.pageSize).
		WithKeyMap(m.btKeyMap())
	m.syncColumns()
	return m
}

// btKeyMap maps our public KeyMap onto bubble-table's bindings. Sort and Open
// never reach bubble-table (handleKey intercepts them), and the built-in
// row-select toggle is disabled so it can't swallow the Open key.
func (m *TableModel) btKeyMap() btable.KeyMap {
	km := btable.DefaultKeyMap()
	km.RowUp = m.KeyMap.Up
	km.RowDown = m.KeyMap.Down
	km.PageUp = m.KeyMap.PageUp
	km.PageDown = m.KeyMap.PageDown
	km.PageFirst = m.KeyMap.Top
	km.PageLast = m.KeyMap.Bottom
	km.Filter = m.KeyMap.Filter
	km.FilterClear = m.KeyMap.Cancel
	km.RowSelectToggle = key.NewBinding(key.WithDisabled())
	return km
}

// ─── data ────────────────────────────────────────────────────────────────────

// SetRows replaces the data, re-applying the active sort (the live filter, if
// any, keeps applying on the bubble-table side).
func (m *TableModel) SetRows(rows []Row) {
	m.rows = rows
	m.doSort()
	m.syncColumns()
	m.syncRows()
}

// SetSize informs the table of its available width/height. The page size is
// derived from height unless WithPageSize fixed it.
func (m *TableModel) SetSize(w, h int) {
	m.width = w
	if !m.fixedPageSize {
		m.pageSize = max(h-chromeRows, 3)
	}
	m.bt = m.bt.WithPageSize(m.pageSize)
	m.syncColumns()
}

// SelectedRow returns the highlighted row, if any.
func (m *TableModel) SelectedRow() (Row, bool) {
	if r, ok := m.bt.HighlightedRow().Data[rowDataKey].(Row); ok {
		return r, true
	}
	return Row{}, false
}

// Filtering reports whether the `/` filter input is active (the host should
// claim keyboard focus while it is, e.g. via navigation.KeyCapturer).
func (m *TableModel) Filtering() bool { return m.bt.GetIsFilterInputFocused() }

// ─── input ───────────────────────────────────────────────────────────────────

// Update handles a message (currently key presses) and may return a Cmd — e.g.
// an OpenDetailMsg-producing Cmd when the user opens a row.
func (m *TableModel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return nil
}

func (m *TableModel) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	// KeyMap is a public field hosts may rebind at runtime; re-map it onto
	// bubble-table before each dispatch so rebinds take effect immediately.
	m.bt = m.bt.WithKeyMap(m.btKeyMap())

	// While the filter input is focused every key belongs to it (Enter/esc
	// blur the input; the filter text stays applied until esc clears it).
	if !m.Filtering() {
		switch {
		case key.Matches(msg, m.KeyMap.Sort):
			m.cycleSort()
			return nil
		case key.Matches(msg, m.KeyMap.Open):
			return m.openSelected()
		}
	}

	var cmd tea.Cmd
	m.bt, cmd = m.bt.Update(msg)
	return cmd
}

// ShortHelp returns the most relevant keybindings for the table context.
func (m *TableModel) ShortHelp() []key.Binding {
	if m.KeyMap == nil {
		return nil
	}
	return []key.Binding{m.KeyMap.Up, m.KeyMap.Down, m.KeyMap.Sort, m.KeyMap.Filter, m.KeyMap.Open}
}

// FullHelp returns all table keybindings organized into groups.
func (m *TableModel) FullHelp() [][]key.Binding {
	if m.KeyMap == nil {
		return nil
	}
	return [][]key.Binding{
		{m.KeyMap.Up, m.KeyMap.Down, m.KeyMap.PageUp, m.KeyMap.PageDown, m.KeyMap.Top, m.KeyMap.Bottom},
		{m.KeyMap.Sort, m.KeyMap.Filter, m.KeyMap.Open, m.KeyMap.Cancel},
	}
}

// HandleClick processes a left click at page-relative (x, y): a header click
// sorts that column; a data-row click selects it, and a quick second click on
// the same row opens its details.
func (m *TableModel) HandleClick(x, y int) tea.Cmd {
	if y == m.headerY {
		if col := m.columnAtX(x); col >= 0 {
			m.sortByCol(col)
		}
		return nil
	}
	idx := y - m.dataStartY
	if idx < 0 {
		return nil
	}
	start, end := m.bt.VisibleIndices()
	row := start + idx
	if row > end {
		return nil
	}
	now := time.Now()
	double := row == m.lastClickRow && now.Sub(m.lastClickTime) < doubleClickWindow
	m.bt = m.bt.WithHighlightedRow(row)
	m.lastClickRow, m.lastClickTime = row, now
	if double {
		m.lastClickTime = time.Time{} // a third click starts a fresh pair
		return m.openSelected()
	}
	return nil
}

// SelectRowAt highlights the data row at page-relative y (as recorded by the
// last View) and returns it. It ignores the header row and out-of-range
// coordinates. Use it to select the row under a right-click before opening a
// context menu, mirroring the single-click selection HandleClick performs.
func (m *TableModel) SelectRowAt(y int) (Row, bool) {
	idx := y - m.dataStartY
	if idx < 0 {
		return Row{}, false
	}
	start, end := m.bt.VisibleIndices()
	row := start + idx
	if row > end {
		return Row{}, false
	}
	m.bt = m.bt.WithHighlightedRow(row)
	return m.SelectedRow()
}

// HandleWheel scrolls the selection by one row per wheel notch.
func (m *TableModel) HandleWheel(up bool) {
	delta := 1
	if up {
		delta = -1
	}
	// WithHighlightedRow clamps to the visible rows and follows pages.
	m.bt = m.bt.WithHighlightedRow(m.bt.GetHighlightedRowIndex() + delta)
}

func (m *TableModel) openSelected() tea.Cmd {
	r, ok := m.SelectedRow()
	if !ok {
		return nil
	}
	rowKey := r.Key
	return func() tea.Msg { return OpenDetailMsg{Key: rowKey} }
}

// ─── sorting ─────────────────────────────────────────────────────────────────

func (m *TableModel) doSort() {
	if !m.sortActive || m.sortCol < 0 || m.sortCol >= len(m.cols) {
		return
	}
	col, asc := m.sortCol, m.sortAsc
	sort.SliceStable(m.rows, func(i, j int) bool {
		return lessCell(m.rows[i], m.rows[j], col, asc)
	})
}

func lessCell(a, b Row, col int, asc bool) bool {
	var ca, cb Cell
	if col < len(a.Cells) {
		ca = a.Cells[col]
	}
	if col < len(b.Cells) {
		cb = b.Cells[col]
	}
	if ca.IsNum && cb.IsNum {
		if asc {
			return ca.Num < cb.Num
		}
		return ca.Num > cb.Num
	}
	la, lb := strings.ToLower(ca.Text), strings.ToLower(cb.Text)
	if asc {
		return la < lb
	}
	return la > lb
}

// sortByCol cycles a column: asc → desc → unsorted. Clearing restores insertion
// order for equal keys only approximately: rows keep whatever stable order the
// previous sorts left them in, matching the old in-place behavior.
func (m *TableModel) sortByCol(col int) {
	if col < 0 || col >= len(m.cols) {
		return
	}
	switch {
	case !m.sortActive || m.sortCol != col:
		m.sortCol, m.sortAsc, m.sortActive = col, true, true
	case m.sortAsc:
		m.sortAsc = false
	default:
		m.sortActive = false
	}
	m.doSort()
	m.syncColumns()
	m.syncRows()
}

// cycleSort walks the current column asc → desc, then advances to the next.
func (m *TableModel) cycleSort() {
	if len(m.cols) == 0 {
		return
	}
	switch {
	case !m.sortActive:
		m.sortByCol(max(m.sortCol, 0))
	case m.sortAsc:
		m.sortByCol(m.sortCol)
	default:
		next := (m.sortCol + 1) % len(m.cols)
		m.sortActive = false
		m.sortByCol(next)
	}
}

// ─── bubble-table sync ───────────────────────────────────────────────────────

// syncColumns rebuilds the bubble-table columns: computed widths, the sort
// indicator on the sorted column's title, and per-column cell padding.
func (m *TableModel) syncColumns() {
	m.colWidths = fitWidths(m.naturalWidths(), m.width)
	cols := make([]btable.Column, len(m.cols))
	pad := lipgloss.NewStyle().Padding(0, 1)
	for i, c := range m.cols {
		title := c.Title
		if m.sortActive && i == m.sortCol {
			if m.sortAsc {
				title += " ▲"
			} else {
				title += " ▼"
			}
		}
		cols[i] = btable.NewColumn(strconv.Itoa(i), title, m.colWidths[i]).
			WithFiltered(c.Filter).
			WithStyle(pad)
	}
	m.bt = m.bt.WithColumns(cols)
}

// syncRows hands the wrapper-sorted rows to bubble-table in display order.
// Each bubble-table row carries the display strings under the column keys
// (which is also what the `/` filter matches) and our Row under rowDataKey.
func (m *TableModel) syncRows() {
	rows := make([]btable.Row, len(m.rows))
	for i, r := range m.rows {
		data := btable.RowData{rowDataKey: r}
		for c := range m.cols {
			if c < len(r.Cells) {
				data[strconv.Itoa(c)] = r.Cells[c].Text
			}
		}
		rows[i] = btable.NewRow(data)
	}
	m.bt = m.bt.WithRows(rows)
}

// naturalWidths measures each column's content: the widest cell or the title
// (plus room for a sort indicator), plus the cell padding.
func (m *TableModel) naturalWidths() []int {
	w := make([]int, len(m.cols))
	for i, c := range m.cols {
		w[i] = lipgloss.Width(c.Title) + 2 // room for " ▲" / " ▼"
		for _, r := range m.rows {
			if i < len(r.Cells) {
				w[i] = max(w[i], lipgloss.Width(r.Cells[i].Text))
			}
		}
		w[i] += 2 // Padding(0, 1)
	}
	return w
}

// fitWidths stretches or squeezes natural column widths to fill total exactly
// (when a width is known): extra cells are distributed round-robin, deficits
// are taken from the widest column first, never below minColWidth.
func fitWidths(natural []int, total int) []int {
	widths := make([]int, len(natural))
	copy(widths, natural)
	if total <= 0 || len(widths) == 0 {
		return widths
	}
	sum := 0
	for _, w := range widths {
		sum += w
	}
	for sum < total {
		for i := range widths {
			if sum == total {
				break
			}
			widths[i]++
			sum++
		}
	}
	for sum > total {
		widest := 0
		for i, w := range widths {
			if w > widths[widest] {
				widest = i
			}
		}
		if widths[widest] <= minColWidth {
			break // can't squeeze further; bubble-table truncates cells
		}
		widths[widest]--
		sum--
	}
	return widths
}

// columnAtX maps a page-relative X to a column index using the computed
// column widths (the borderless layout has no separator cells between them).
func (m *TableModel) columnAtX(x int) int {
	acc := 0
	for i, w := range m.colWidths {
		acc += w
		if x < acc {
			return i
		}
	}
	return -1
}

// ─── rendering (delegated to bubble-table) ───────────────────────────────────

// View renders the table for the given palette, at vertical origin originY
// within the page content (the number of lines drawn above it). It records
// geometry for HandleClick and appends a status footer below the table.
func (m *TableModel) View(c *styles.AppStyle, originY int) string {
	if len(m.cols) == 0 {
		return ""
	}
	st := c.Styles

	selectedStyle := st.SelectedItem
	cellStyle := st.TextOnBg
	zebraStyle := st.Subtitle

	bt := m.bt.
		WithBaseStyle(lipgloss.NewStyle().Align(lipgloss.Left)).
		HeaderStyle(st.Title).
		WithRowStyleFunc(func(in btable.RowStyleFuncInput) lipgloss.Style {
			if in.IsHighlighted {
				return selectedStyle
			}
			if in.Index%2 == 1 {
				return zebraStyle
			}
			return cellStyle
		})
	out := bt.View()

	// Borderless layout: the header is the first line, data rows follow.
	m.headerY = originY
	m.dataStartY = originY + 1

	return lipgloss.JoinVertical(lipgloss.Left, out, m.footer(c, lipgloss.Width(out)))
}

func (m *TableModel) footer(c *styles.AppStyle, w int) string {
	st := c.Styles
	var left string
	if total := m.bt.TotalRows(); total > 0 {
		left = " " + strconv.Itoa(m.bt.GetHighlightedRowIndex()+1) + "/" + strconv.Itoa(total) + " "
	} else {
		left = " 0/0 "
	}
	if m.sortActive {
		dir := "▲"
		if !m.sortAsc {
			dir = "▼"
		}
		left += "· sorted by " + m.cols[m.sortCol].Title + " " + dir + " "
	}

	filter := m.bt.GetCurrentFilter()
	var right string
	switch {
	case m.Filtering():
		right = "/" + filter + "▏"
	case filter != "":
		right = "filter: " + filter + " (esc clears) "
	case !m.HideFooterHint:
		right = "s/click header: sort · enter/double-click: details · / filter "
	}

	// Right-align the hint text across the remaining cells (at least one so
	// the sides never touch) instead of hand-building a space filler.
	gap := max(1, w-lipgloss.Width(left)-lipgloss.Width(right))
	return st.Subtitle.Render(left) + lipgloss.PlaceHorizontal(
		gap+lipgloss.Width(right),
		lipgloss.Right,
		st.Subtitle.Render(right),
	)
}
