package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type notice struct {
	Title      string `json:"notice_title"`
	FilePath   string `json:"file_path"`
	UploadedOn string `json:"notice_date"`
}

type kbmsg struct {
	key string
	msg string
}

type keybind struct {
	up    kbmsg
	down  kbmsg
	enter kbmsg
	save  kbmsg
	png   kbmsg
	quit  kbmsg
}

var keybinds keybind = keybind{
	quit:  kbmsg{"q", "to quit"},
	up:    kbmsg{"↑/k", "move up"},
	down:  kbmsg{"↓/j", "move up"},
	enter: kbmsg{"↵ Enter", "to select"},
	save:  kbmsg{"s", "to save pdf"},
	png:   kbmsg{"p", "to save as png"},
}

type model struct {
	// loader
	loader     spinner.Model
	isLoading  bool
	loadingMsg string
	// table
	table     table.Model
	showTable bool
	// viewport
	viewport     viewport.Model
	windowHeight int
	// meta
	exiting  bool
	data     any
	dataSize int
	showData bool
	// toast
	showToast bool
	err       error
}

// this is only for to show pdf in text
type display struct {
	show bool
	text string
	err  error
}

type toast struct {
	show bool
	msg  string
	err  error
}

var (
	wg         sync.WaitGroup
	tmpDir     = os.TempDir()
	homeDir, _ = os.UserHomeDir()
)

func (m model) Init() tea.Cmd {
	return tea.Batch(m.loader.Tick, getData())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		/* (4+4) for padding + (1+1) for borders */
		m.viewport.Width = msg.Width - 10
		/* 3 for help text + (1+1) for padding +
		   (1+1) for borders + 1 for bottom space
		*/
		m.viewport.Height = msg.Height - 8

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.showData = false
			m.showTable = false
			m.isLoading = false
			m.exiting = true
			return m, tea.Quit

		case "enter":
			m.showTable = false
			m.isLoading = true
			m.loadingMsg = "Downloading pdf..."

			selected := m.table.SelectedRow()
			return m, tea.Batch(
				m.loader.Tick,
				showData(selected[3], selected[1]),
			)

		case "s":
			if m.showTable || m.isLoading {
				return m, nil
			}
			selected := m.table.SelectedRow()
			return m, saveFileCMD(selected[1])

		case "p":
			// if m.showTable || m.isLoading {
			// 	return m, nil
			// }
			m.isLoading = true
			m.loadingMsg = "Convertig pdf to png..."

			selected := m.table.SelectedRow()
			return m, tea.Batch(
				m.loader.Tick,
				saveAsPngCMD(selected[1]),
			)
		}

	case model:
		var titleWidth int
		var data = msg.data.([]notice)
		var rows = make([]table.Row, 0, 12)

		for i, v := range data {
			titleWidth += len(v.Title)
			slno := toString(i + 1)
			if i+1 < 10 {
				slno = "0" + slno
			}
			rows = append(rows, []string{slno, v.Title, v.UploadedOn, v.FilePath})
		}

		m.dataSize = len(data)
		cols := []table.Column{
			{Title: "ID", Width: 3},
			{Title: "Notice title", Width: titleWidth / m.dataSize},
			{Title: "Uploaded On", Width: len("uploaded on")},
			{Title: "Link", Width: 4},
		}
		m.table.SetColumns(cols)
		m.table.SetRows(rows)
		m.table.Focus()
		m.isLoading = false
		m.showTable = true

		return m, nil

	case display:
		if msg.err != nil {
			return m, tea.Batch(
				tea.Println(msg.err.Error()),
				tea.Quit,
			)
		}
		m.data = msg.text
		m.isLoading = false
		m.showTable = false
		m.showData = true

		return m, nil

	case toast:
		if msg.err != nil {
			m.err = msg.err
		}
		m.data = msg.msg
		m.isLoading = false
		m.showData = false
		m.showTable = false
		m.showToast = true

		return m, nil
	}

	if m.isLoading {
		var cmd tea.Cmd
		m.loader, cmd = m.loader.Update(msg)
		return m, cmd
	}

	if m.showTable {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}

	if m.showData {
		var cmd tea.Cmd
		var cmds []tea.Cmd

		m.viewport.SetContent(m.data.(string))
		m.viewport, cmd = m.viewport.Update(msg)

		cmds = append(cmds, cmd)

		return m, tea.Batch(cmds...)
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	withBorder := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("0"))

	withHelp := func(kbs ...kbmsg) {
		s.WriteString("\n\n")
		for i, v := range kbs {
			helpText := showHelpKey(v)
			s.WriteString(helpText)

			if i < len(kbs)-1 {
				s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Render("  •  "))
			}
		}
	}

	if m.exiting {
		s.Reset()
		s.WriteString("Bye bye!")
	}

	if m.isLoading {
		s.Reset()
		if m.loadingMsg == "" {
			m.loadingMsg = "Fetching data, please wait..."
		}
		s.WriteString(fmt.Sprintf(" %s %s", m.loader.View(), m.loadingMsg))
		withHelp(keybinds.quit)
	}

	if m.showTable {
		s.Reset()
		data := fmt.Sprintf("%s\nTotal %s",
			withBorder.Render(m.table.View()),
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Bold(true).
				Render(toString(m.dataSize)),
		)
		s.WriteString(data)
		withHelp(keybinds.up, keybinds.down, keybinds.enter, keybinds.quit)
	}

	if m.showData {
		s.Reset()
		data := lipgloss.NewStyle().
			Padding(1, 4).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("0")).
			Render(m.viewport.View())
		s.WriteString(data)
		withHelp(keybinds.save, keybinds.png, keybinds.quit)
	}

	if m.showToast {
		s.Reset()
		if m.err != nil {
			s.WriteString(m.err.Error())
		}
		data := lipgloss.NewStyle().
			Padding(1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("0")).
			Render("File saved to " + m.data.(string))
		s.WriteString(data)
		withHelp(keybinds.quit)
	}

	return "\n" + s.String()
}

func main() {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	t := table.New()
	ts := table.DefaultStyles()

	ts.Header = ts.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("8")).
		Foreground(lipgloss.Color("8")).
		Bold(true)
	ts.Selected = ts.Selected.
		Foreground(lipgloss.Color("1")).
		Bold(true)
	t.SetStyles(ts)
	t.SetHeight(16)

	myModel := model{
		loader:    s,
		isLoading: true,
		table:     t,
	}

	expect(
		func() (tea.Model, error) {
			program := tea.NewProgram(myModel, tea.WithAltScreen(), tea.WithMouseCellMotion())
			return program.Run()
		},
		"Could no start new tea program",
	)
}

func downloadFile(link string, filename string) bool {
	isFileCreated := expect(func() (bool, error) {
		res, err := http.Get(link)
		if err != nil {
			return false, err
		}
		defer res.Body.Close()

		f, err := os.Create("/tmp/" + filename + ".pdf")
		if err != nil {
			return false, err
		}
		defer f.Close()

		_, terr := io.Copy(f, res.Body)
		return true, terr
	}, "Could not download file")

	return isFileCreated
}

func convertPDFtoText(filename string) bool {
	err := exec.Command("pdftotext", "/tmp/"+filename+".pdf").Run()
	if err != nil {
		return false
	}
	return true
}

func readFile(filename string) string {
	data := expect(
		func() ([]byte, error) {
			return os.ReadFile("/tmp/" + filename + ".txt")
		},
		"Cannot read converted text file",
	)
	return string(data)
}

func getData() tea.Cmd {
	type result struct {
		Data []notice `json:"data"`
	}
	var data result

	wg.Add(1)
	go func() {
		data = fetch[result]("makaut1.ucanapply.com/smartexam/public/api/notice-data")
		wg.Done()
	}()

	return func() tea.Msg {
		wg.Wait()
		return model{data: data.Data}
	}
}

func showData(link, filename string) tea.Cmd {
	var err error
	var data string

	wg.Add(1)
	go func() {
		defer wg.Done()

		ok := downloadFile(link, filename)
		if !ok {
			err = fmt.Errorf("Could not download file")
		}

		ok = convertPDFtoText(filename)
		if !ok {
			err = fmt.Errorf("Could not found/convert pdf")
		}

		data = readFile(filename)
	}()

	return func() tea.Msg {
		wg.Wait()
		return display{text: data, err: err}
	}
}

func saveFile(filename string) (string, error) {
	f, err := os.Stat("/tmp/" + filename + ".pdf")

	if f == nil || err != nil {
		return "", fmt.Errorf("File is not downloaded to save")
	}

	downloadDest := fmt.Sprintf("%s/Downloads/%s",
		homeDir,
		f.Name(),
	)

	err = moveFile("/tmp/"+f.Name(), downloadDest)

	if err != nil {
		return "", fmt.Errorf("Could not save file to %s\n", downloadDest)
	}

	return f.Name(), nil
}

func saveFileCMD(filename string) tea.Cmd {
	f, err := saveFile(filename)

	return func() tea.Msg {
		return toast{msg: f, err: err}
	}
}

func saveAsPng(filename string) (string, error) {
	cmd := "pdftocairo"
	file := fmt.Sprintf("%s/%s.pdf", tmpDir, filename)
	dest := fmt.Sprintf("%s/Downloads/%s", homeDir, filename)

	f, err := os.Stat(file)
	if f == nil || err != nil {
		return "", fmt.Errorf("File is not downloaded to save")
	}

	if !which(cmd) {
		return "", fmt.Errorf("`%s` is not installed", cmd)
	}

	err = exec.Command(cmd, file, dest, "-png", "-singlefile").Run()
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	return dest + ".png", nil
}

func saveAsPngCMD(filename string) tea.Cmd {
	var msg string
	var err error

	wg.Add(1)
	go func() {
		msg, err = saveAsPng(filename)
		wg.Done()
	}()

	return func() tea.Msg {
		wg.Wait()
		return toast{msg: msg, err: err}
	}
}

func showHelpKey(kb kbmsg) string {
	keyColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("1")).
		Bold(true).
		Render
	msgColor := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Render
	text := fmt.Sprintf("%s %s", keyColor(kb.key), msgColor(kb.msg))
	return text
}

// utilities
func which(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func fetch[T any](url_and_params ...string) T {
	var data T
	var s strings.Builder

	url := url_and_params[0]
	if url[:4] != "http" {
		url = "https://" + url
	}
	s.WriteString(url)

	params := url_and_params[1:]
	paramsLen := len(params)

	if paramsLen > 1 {
		s.WriteString("?")
		for i, v := range params {
			s.WriteString(v)
			if i < paramsLen-1 {
				s.WriteString("&")
			}
		}
	}

	res := expect(func() (*http.Response, error) {
		return http.Get(s.String())
	})
	defer res.Body.Close()

	s.Reset()
	maybe(
		json.NewDecoder(res.Body).Decode(&data),
		"Could not decode respose:body",
	)
	return data
}

func expect[T any](fn func() (T, error), msg ...string) T {
	var errmsg string

	data, err := fn()
	if err != nil {
		if len(msg) > 0 {
			errmsg = "[EXCEPTION]: " + err.Error() + "\n" +
				/**/ "  caused by: " + msg[0]
		} else {
			errmsg = "[EXCEPTION]: " + err.Error()
		}
		fmt.Println(errmsg)
		os.Exit(1)
	}
	return data
}

func maybe(err error, msg string) {
	if err != nil {
		fmt.Println("[ERROR]", msg)
		fmt.Println("REASON:", err.Error())
		os.Exit(1)
	}
}

func toString(s any) string {
	switch s := s.(type) {
	case int:
		return strconv.Itoa(s)
	case int64:
		return strconv.FormatInt(s, 10)
	case byte:
		return string(s)
	default:
		return ""
	}
}

/*
   GoLang: os.Rename() give error "invalid cross-device link" for Docker container with Volumes.
   MoveFile(source, destination) will work moving file between folders

 see: https://gist.github.com/var23rav/23ae5d0d4d830aff886c3c970b8f6c6b
*/

func moveFile(sourcePath, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}
	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}
	// The copy was successful, so now delete the original file
	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
