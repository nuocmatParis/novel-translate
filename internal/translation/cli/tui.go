package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/term"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/translation/engine"
	"github.com/voocel/ainovel-cli/internal/translation/model"
	"github.com/voocel/ainovel-cli/internal/translation/state"
	"github.com/voocel/ainovel-cli/internal/translation/workspace"
)

const tuiHelp = `Mở giao diện terminal dịch EPUB Nhật → Việt.

novel-translate-cli tui [--workspace <path>] [--volume <id>]
novel-translate-cli --tui [--workspace <path>]

Không truyền workspace: mở workspace tại thư mục hiện tại nếu có,
hoặc chọn Mở/Tạo workspace trên màn hình chào.
↑/↓ chọn thao tác; ←/→ chọn tập; Enter thực hiện; PgUp/PgDn cuộn log.
Esc hủy hộp thoại; Ctrl+C dừng tác vụ; q thoát khi rảnh.
Mở TUI không gọi API. Dịch/review yêu cầu xác nhận trước khi gọi model.
`

func runTUI(ctx context.Context, args []string, out, errout io.Writer, options Options) error {
	f := flag.NewFlagSet("tui", flag.ContinueOnError)
	f.SetOutput(errout)
	path := f.String("workspace", "", "workspace path")
	volume := f.String("volume", "", "selected volume ID")
	f.Usage = func() { fmt.Fprint(errout, tuiHelp) }
	if err := f.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if f.NArg() != 0 {
		return errors.New("tui chỉ nhận --workspace và --volume")
	}
	if *path == "" {
		if _, err := os.Stat("workspace.yaml"); err == nil {
			*path = "."
		}
	}
	m := newTUI(ctx, options)
	if *path != "" {
		c, err := workspace.Load(*path)
		if err != nil {
			return err
		}
		m.config = &c
		if *volume != "" {
			if _, err := c.Volume(*volume); err != nil {
				return err
			}
			for i, v := range c.Volumes {
				if v.ID == *volume {
					m.volume = i
				}
			}
		}
	} else if *volume != "" {
		return errors.New("--volume cần một workspace")
	}
	input := options.Input
	if input == nil {
		if !term.IsTerminal(os.Stdin.Fd()) {
			return errors.New("TUI cần terminal tương tác; mở PowerShell/Terminal rồi chạy novel-translate-cli tui")
		}
		input = os.Stdin
	}
	// Child jobs are also canceled if the terminal/program closes unexpectedly.
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	m.ctx = runCtx
	_, err := tea.NewProgram(m, tea.WithInput(input), tea.WithOutput(out), tea.WithContext(runCtx), tea.WithAltScreen()).Run()
	return err
}

type tuiAction struct {
	label   string
	command []string
	form    string
	paid    bool
	confirm bool
}

func (m *translationTUI) actions() []tuiAction {
	common := []tuiAction{{label: "Mở workspace", form: "open"}, {label: "Tạo workspace mới", form: "create"}}
	if m.config == nil {
		return common
	}
	actions := []tuiAction{
		{label: "Làm mới tiến độ", command: []string{"status"}},
		{label: "Thêm tập / EPUB nguồn", form: "add"},
		{label: "Chọn file cấu hình API", form: "provider"},
	}
	if len(m.config.Volumes) > 0 {
		actions = append(actions,
			tuiAction{label: "Nhập EPUB của tập đang chọn", command: []string{"import"}},
			tuiAction{label: "Dịch / tiếp tục tập đang chọn", command: []string{"resume"}, paid: true},
			tuiAction{label: "Review bản dịch", command: []string{"review"}, paid: true},
			tuiAction{label: "Đồng bộ bản sửa tay", command: []string{"sync"}},
			tuiAction{label: "Kiểm tra độ đầy đủ", command: []string{"verify"}},
			tuiAction{label: "Xuất EPUB xem trước", command: []string{"export", "--preview"}},
			tuiAction{label: "Xuất EPUB hoàn chỉnh", command: []string{"export"}, confirm: true},
			tuiAction{label: "Danh sách ảnh / trạng thái xử lý chữ", command: []string{"image", "list"}},
		)
	}
	actions = append(actions, tuiAction{label: "Xem skills / quy ước", command: []string{"skill", "list"}})
	return append(actions, common...)
}

type tuiVolumeStatus struct {
	ID             string   `json:"id"`
	Accepted       int      `json:"accepted"`
	Eligible       int      `json:"eligible"`
	Translated     int      `json:"translated"`
	Imported       bool     `json:"source_imported"`
	Stale          []string `json:"stale"`
	FreshnessError string   `json:"freshness_error"`
}

type tuiSnapshot struct {
	Volumes []tuiVolumeStatus  `json:"volumes"`
	Usage   model.UsageSummary `json:"usage"`
}

type tuiField struct{ label, value string }
type tuiForm struct {
	kind   string
	fields []tuiField
	index  int
}
type tuiLogMsg string
type tuiDoneMsg struct {
	config      *workspace.Config
	snapshot    tuiSnapshot
	err         error
	snapshotErr error
}
type tuiTickMsg struct{}

type translationTUI struct {
	ctx        context.Context
	options    Options
	config     *workspace.Config
	snapshot   tuiSnapshot
	volume     int
	selected   int
	width      int
	height     int
	logs       []string
	scroll     int
	form       *tuiForm
	pending    *tuiAction
	busy       bool
	jobLabel   string
	notice     string
	cancel     context.CancelFunc
	events     chan tea.Msg
	frame      int
	quitOnDone bool
}

func newTUI(ctx context.Context, options Options) *translationTUI {
	return &translationTUI{ctx: ctx, options: options, width: 90, height: 30, notice: "Chọn workspace để bắt đầu. Mở giao diện không gọi API."}
}

func (m *translationTUI) Init() tea.Cmd {
	if m.config != nil {
		return m.start("Đọc tiến độ", []string{"status"}, m.config.Root)
	}
	return nil
}

func (m *translationTUI) addLog(text string) {
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if line = cleanTUIText(line); strings.TrimSpace(line) != "" {
			m.logs = append(m.logs, ansi.Truncate(line, 2000, "…"))
		}
	}
	if len(m.logs) > 500 {
		m.logs = append([]string(nil), m.logs[len(m.logs)-500:]...)
	}
}

// Source text, provider errors and file names must never inject terminal escapes.
func cleanTUIText(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, ansi.Strip(s))
}

func (m *translationTUI) wait() tea.Cmd {
	events, ctx := m.events, m.ctx
	return func() tea.Msg {
		select {
		case event := <-events:
			return event
		case <-ctx.Done():
			return nil
		}
	}
}

func tuiTick() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(time.Time) tea.Msg { return tuiTickMsg{} })
}

// stdout is bounded: import may print an entire book. Progress is streamed to
// the UI instead of writing directly over Bubble Tea's renderer.
type tuiOutput struct{ data []byte }

func (w *tuiOutput) Write(p []byte) (int, error) {
	const limit = 24000
	n := len(p)
	if len(p) > limit {
		p = p[len(p)-limit:]
	}
	w.data = append(w.data, p...)
	if len(w.data) > limit {
		w.data = append([]byte(nil), w.data[len(w.data)-limit:]...)
	}
	return n, nil
}

type tuiProgress struct {
	ctx    context.Context
	events chan tea.Msg
}

func (w tuiProgress) Write(p []byte) (int, error) {
	select {
	case w.events <- tuiLogMsg(string(p)):
		return len(p), nil
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	}
}

func (m *translationTUI) start(label string, args []string, path string) tea.Cmd {
	command := append([]string(nil), args...)
	if len(command) > 0 && command[0] == "status" {
		return m.startOperation(label, path, nil)
	}
	if len(command) > 0 && command[0] != "init" {
		command = append(command, "--workspace", path)
	}
	options := m.options
	return m.startOperation(label, path, func(ctx context.Context, out, progress io.Writer) error {
		return execute(ctx, command, out, progress, options)
	})
}

func (m *translationTUI) startOperation(label, path string, operation func(context.Context, io.Writer, io.Writer) error) tea.Cmd {
	if m.busy {
		return nil
	}
	m.busy, m.jobLabel, m.notice, m.scroll = true, label, "", 0
	m.addLog("› " + label)
	jobCtx, cancel := context.WithCancel(m.ctx)
	m.cancel = cancel
	m.events = make(chan tea.Msg, 64)
	events, rootCtx := m.events, m.ctx
	go func() {
		defer cancel()
		var output tuiOutput
		var err error
		progress := tuiProgress{ctx: jobCtx, events: events}
		if operation != nil {
			err = operation(jobCtx, &output, progress)
			if len(output.data) > 0 {
				_, _ = progress.Write(output.data)
			}
		}
		result := tuiDoneMsg{err: err}
		if err == nil || mPathExists(path) {
			c, loadErr := workspace.Load(path)
			if loadErr == nil {
				result.config = &c
				var raw bytes.Buffer
				result.snapshotErr = status(&raw, c, state.Store{Root: c.Root}, true)
				if result.snapshotErr == nil {
					result.snapshotErr = json.Unmarshal(raw.Bytes(), &result.snapshot)
				}
			} else {
				result.snapshotErr = loadErr
			}
		}
		// Even canceled jobs report completion: q waits for the writer to stop.
		select {
		case events <- result:
		case <-rootCtx.Done():
		}
	}()
	return tea.Batch(m.wait(), tuiTick())
}

func mPathExists(path string) bool {
	name, err := workspace.ConfigPath(path)
	if err != nil {
		return false
	}
	_, err = os.Stat(name)
	return err == nil
}

func (m *translationTUI) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tuiTickMsg:
		if m.busy {
			m.frame++
			return m, tuiTick()
		}
	case tuiLogMsg:
		m.addLog(string(msg))
		return m, m.wait()
	case tuiDoneMsg:
		m.busy = false
		if m.cancel != nil {
			m.cancel()
			m.cancel = nil
		}
		if msg.config != nil {
			m.config, m.snapshot = msg.config, msg.snapshot
			m.volume = min(m.volume, max(0, len(m.config.Volumes)-1))
			m.selected = min(m.selected, len(m.actions())-1)
		}
		m.notice = "Đã xong: " + m.jobLabel
		if msg.err != nil {
			m.notice = "Lỗi: " + msg.err.Error()
			if errors.Is(msg.err, engine.ErrUnresolved) {
				m.notice = "Còn vấn đề cần xử lý: " + msg.err.Error()
			}
		} else if msg.snapshotErr != nil {
			m.notice = "Không đọc được tiến độ: " + msg.snapshotErr.Error()
		}
		m.addLog(m.notice)
		if m.quitOnDone {
			return m, tea.Quit
		}
	case tea.KeyMsg:
		key := msg.String()
		if m.busy {
			switch key {
			case "ctrl+c", "esc", "q":
				m.cancel()
				m.notice = "Đang dừng và chờ tác vụ kết thúc…"
				m.quitOnDone = key == "q"
			case "pgup":
				m.scroll += 5
			case "pgdown":
				m.scroll = max(0, m.scroll-5)
			}
			return m, nil
		}
		if m.form != nil {
			return m, m.updateForm(msg)
		}
		if m.pending != nil {
			switch key {
			case "y":
				action := *m.pending
				m.pending = nil
				return m, m.runAction(action)
			case "n", "esc", "ctrl+c":
				m.pending = nil
			}
			return m, nil
		}
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.selected = (m.selected + len(m.actions()) - 1) % len(m.actions())
		case "down", "j", "tab":
			m.selected = (m.selected + 1) % len(m.actions())
		case "left", "right":
			if m.config != nil && len(m.config.Volumes) > 0 {
				delta := 1
				if key == "left" {
					delta = -1
				}
				m.volume = (m.volume + delta + len(m.config.Volumes)) % len(m.config.Volumes)
			}
		case "pgup":
			m.scroll += 5
		case "pgdown":
			m.scroll = max(0, m.scroll-5)
		case "r":
			if m.config != nil {
				return m, m.start("Đọc tiến độ", []string{"status"}, m.config.Root)
			}
		case "enter":
			action := m.actions()[m.selected]
			if action.form != "" {
				m.openForm(action.form)
			} else if action.paid || action.confirm {
				m.pending = &action
			} else {
				return m, m.runAction(action)
			}
		}
	}
	return m, nil
}

func (m *translationTUI) runAction(action tuiAction) tea.Cmd {
	args := append([]string(nil), action.command...)
	if args[0] != "status" && args[0] != "skill" {
		args = append(args, "--volume", m.config.Volumes[m.volume].ID)
	}
	return m.start(action.label, args, m.config.Root)
}

func (m *translationTUI) openForm(kind string) {
	f := &tuiForm{kind: kind}
	switch kind {
	case "open":
		f.fields = []tuiField{{label: "Đường dẫn workspace"}}
	case "create":
		f.fields = []tuiField{{label: "Thư mục workspace mới"}, {label: "Series ID (vd: rezero)"}, {label: "Tên bộ truyện"}}
	case "add":
		order := 1
		previous := ""
		for _, v := range m.config.Volumes {
			if v.Order >= order {
				order, previous = v.Order+1, v.ID
			}
		}
		f.fields = []tuiField{{label: "ID tập", value: fmt.Sprintf("vol%02d", order)}, {label: "Thứ tự tập", value: strconv.Itoa(order)}, {label: "Đường dẫn EPUB Nhật"}, {label: "Tập trước (để trống nếu chưa có)", value: previous}}
	case "provider":
		path := m.config.ProviderConfig
		if path == "" {
			path = bootstrap.DefaultConfigPath()
		}
		f.fields = []tuiField{{label: "File cấu hình provider JSON/JSONC", value: path}}
	}
	m.form = f
}

func (m *translationTUI) updateForm(key tea.KeyMsg) tea.Cmd {
	f := m.form
	switch key.String() {
	case "esc", "ctrl+c":
		m.form = nil
	case "tab", "down":
		f.index = (f.index + 1) % len(f.fields)
	case "shift+tab", "up":
		f.index = (f.index + len(f.fields) - 1) % len(f.fields)
	case "backspace", "ctrl+h":
		runes := []rune(f.fields[f.index].value)
		if len(runes) > 0 {
			f.fields[f.index].value = string(runes[:len(runes)-1])
		}
	case "ctrl+u":
		f.fields[f.index].value = ""
	case "enter":
		if f.index < len(f.fields)-1 {
			f.index++
			return nil
		}
		return m.submitForm()
	default:
		if key.Type == tea.KeyRunes || key.Type == tea.KeySpace {
			text := string(key.Runes)
			if key.Type == tea.KeySpace {
				text = " "
			}
			f.fields[f.index].value += cleanTUIText(text)
		}
	}
	return nil
}

func (m *translationTUI) submitForm() tea.Cmd {
	f := m.form
	value := func(i int) string { return strings.Trim(strings.TrimSpace(f.fields[i].value), "\"'") }
	var args []string
	path := ""
	switch f.kind {
	case "open":
		if value(0) == "" || !mPathExists(value(0)) {
			m.notice = "Không tìm thấy workspace.yaml tại đường dẫn đã nhập."
			return nil
		}
		path, args = value(0), []string{"status"}
	case "create":
		if value(0) == "" || !workspace.ValidID(value(1)) || value(2) == "" {
			m.notice = "Cần đường dẫn, Series ID hợp lệ và tên bộ truyện."
			return nil
		}
		path = value(0)
		args = []string{"init", path, "--series", value(1), "--title", value(2)}
	case "add":
		path = m.config.Root
		args = []string{"volume", "add", "--id", value(0), "--order", value(1), "--source", value(2)}
		if value(3) != "" {
			args = append(args, "--previous", value(3))
		}
	case "provider":
		// Never read/display keys in the TUI. Only persist the config file path,
		// under the same workspace lock as other configuration mutations.
		path := m.config.ReadPath(value(0))
		if value(0) == "" {
			m.notice = "Cần đường dẫn file cấu hình."
			return nil
		}
		if info, err := os.Stat(path); err != nil || info.IsDir() {
			m.notice = "Không tìm thấy file cấu hình provider."
			return nil
		}
		root := m.config.Root
		m.form = nil
		return m.startOperation("Lưu đường dẫn cấu hình API", root, func(ctx context.Context, out, progress io.Writer) error {
			unlock, err := (state.Store{Root: root}).Lock(ctx)
			if err != nil {
				return err
			}
			defer unlock()
			c, err := workspace.Load(root)
			if err != nil {
				return err
			}
			c.ProviderConfig = filepath.Clean(path)
			return c.Save()
		})
	}
	m.form = nil
	return m.start("Workspace: "+f.kind, args, path)
}

func (m *translationTUI) View() string {
	width := max(1, m.width-2)
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	selected := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24"))
	line := func(s string) string { return ansi.Truncate(s, width, "…") }
	var content []string
	content = append(content, accent.Render("NOVEL TRANSLATE  /  Nhật → Việt"))
	if m.config != nil {
		content = append(content, line(cleanTUIText(m.config.Series.Title)+"  ·  "+cleanTUIText(m.config.Root)))
		provider := m.config.ProviderConfig
		if provider == "" {
			provider = bootstrap.DefaultConfigPath()
		}
		content = append(content, muted.Render(line("API config: "+cleanTUIText(provider))))
		if len(m.config.Volumes) == 0 {
			content = append(content, "Chưa có tập. Chọn Thêm tập / EPUB nguồn.")
		} else {
			v := m.config.Volumes[m.volume]
			text := fmt.Sprintf("◀ Tập %d · %s ▶  [%d/%d]  Chưa nhập EPUB", v.Order, v.ID, m.volume+1, len(m.config.Volumes))
			for _, s := range m.snapshot.Volumes {
				if s.ID == v.ID && s.Imported {
					text = fmt.Sprintf("◀ Tập %d · %s ▶  Dịch %d/%d · Đối chiếu %d · Cần xem lại %d", v.Order, v.ID, s.Translated, s.Eligible, s.Accepted, len(s.Stale))
				}
			}
			content = append(content, accent.Render(line(text)))
		}
		u := m.snapshot.Usage
		content = append(content, muted.Render(line(fmt.Sprintf("%d requests · Chi phí đã biết $%.4f · %d requests chưa rõ giá", u.Requests, u.KnownCostUSD, u.UnknownCostRequests))))
	} else {
		content = append(content, "Mỗi workspace là một bộ truyện, chứa nhiều tập EPUB.")
	}
	content = append(content, "")
	if m.form != nil {
		content = append(content, accent.Render("Nhập thông tin"))
		for i, f := range m.form.fields {
			text := "  " + f.label + ": " + cleanTUIText(f.value)
			if i == m.form.index {
				text = "› " + f.label + ": " + cleanTUIText(f.value) + "▏"
				text = selected.Render(line(text))
			} else {
				text = line(text)
			}
			content = append(content, text)
		}
		content = append(content, muted.Render("Tab/↑/↓ đổi ô · Enter tiếp/lưu · Ctrl+U xóa ô · Esc hủy"))
	} else if m.pending != nil {
		content = append(content, accent.Render(line("Xác nhận: "+m.pending.label)))
		if m.pending.paid {
			content = append(content, line("Gọi API thật, có thể tính phí; tự chạy tiếp theo giới hạn workspace."))
		} else {
			content = append(content, line("Xuất vào Output; có thể thay file xuất trước đó khi mọi kiểm tra đạt."))
		}
		content = append(content, "Nhấn y để chạy · n/Esc hủy")
	} else {
		actions := m.actions()
		visible := max(3, min(len(actions), m.height-17))
		start := max(0, min(m.selected-visible+1, len(actions)-visible))
		for i := start; i < min(len(actions), start+visible); i++ {
			text := "  " + actions[i].label
			if i == m.selected {
				text = selected.Render(line("› " + actions[i].label))
			} else {
				text = line(text)
			}
			content = append(content, text)
		}
		if visible < len(actions) {
			content = append(content, muted.Render(fmt.Sprintf("  Thao tác %d/%d · ↑/↓ để xem thêm", m.selected+1, len(actions))))
		}
	}
	content = append(content, "")
	if m.busy {
		content = append(content, accent.Render(line([]string{"⠋", "⠙", "⠹", "⠸"}[m.frame%4]+" "+m.jobLabel)))
	}
	if m.notice != "" {
		content = append(content, line(cleanTUIText(m.notice)))
	}
	logHeight := max(1, m.height-len(content)-3)
	content = append(content, muted.Render("── Nhật ký ──"))
	end := max(0, len(m.logs)-min(m.scroll, max(0, len(m.logs)-logHeight)))
	for _, entry := range m.logs[max(0, end-logHeight):end] {
		content = append(content, line(entry))
	}
	footer := "↑/↓ chọn · ←/→ đổi tập · Enter chạy · r làm mới · PgUp/PgDn log · q thoát"
	if m.busy {
		footer = "Ctrl+C/Esc dừng tác vụ · q dừng rồi thoát · PgUp/PgDn cuộn nhật ký"
	}
	content = append(content, muted.Render(line(footer)))
	for i := range content {
		content[i] = line(content[i])
	}
	return strings.Join(content, "\n")
}
