package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Embed all template files into the compiled binary so Seppy runs anywhere standalone
//
//go:embed all:template
var embeddedTemplates embed.FS

// User Config Schema (~/.seppy/config.json)
type CustomSkillCommand struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type UserConfig struct {
	DefaultUnselectAll   bool                 `json:"default_unselect_all"`
	CustomSkillsCommands []CustomSkillCommand `json:"custom_skills_commands"`
	CustomNpmPackages    []string             `json:"custom_npm_packages"`
}

// Windows API DLLs & Procedures
var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	user32                         = syscall.NewLazyDLL("user32.dll")
	procBeep                       = kernel32.NewProc("Beep")
	procGetConsoleWindow           = kernel32.NewProc("GetConsoleWindow")
	procGetSystemMetrics           = user32.NewProc("GetSystemMetrics")
	procGetWindowRect              = user32.NewProc("GetWindowRect")
	procSetWindowPos               = user32.NewProc("SetWindowPos")
	procGetStdHandle               = kernel32.NewProc("GetStdHandle")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")

	lastNavSoundTime time.Time
)

type coord struct {
	X, Y int16
}

type smallRect struct {
	Left, Top, Right, Bottom int16
}

type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

func getTerminalSize() (int, int) {
	if runtime.GOOS == "windows" {
		handle, _, _ := procGetStdHandle.Call(uintptr(0xFFFFFFF5)) // STD_OUTPUT_HANDLE = -11
		if handle != 0 && handle != uintptr(syscall.InvalidHandle) {
			var csbi consoleScreenBufferInfo
			r, _, _ := procGetConsoleScreenBufferInfo.Call(handle, uintptr(unsafe.Pointer(&csbi)))
			if r != 0 {
				w := int(csbi.Window.Right - csbi.Window.Left + 1)
				h := int(csbi.Window.Bottom - csbi.Window.Top + 1)
				if w > 10 && h > 5 {
					return w, h
				}
			}
		}
	}
	return 80, 24
}

// Audio Worker Queue (Eliminates unbounded goroutine leaks on keypresses)
type beepSound int

const (
	soundNav beepSound = iota
	soundToggle
	soundConfirm
)

var beepChan = make(chan beepSound, 32)

func startAudioWorker() {
	go func() {
		for sound := range beepChan {
			if runtime.GOOS == "windows" {
				switch sound {
				case soundNav:
					procBeep.Call(uintptr(550), uintptr(12))
				case soundToggle:
					procBeep.Call(uintptr(784), uintptr(14))
					procBeep.Call(uintptr(987), uintptr(18))
				case soundConfirm:
					procBeep.Call(uintptr(1046), uintptr(15))
					procBeep.Call(uintptr(1318), uintptr(22))
				}
			}
		}
	}()
}

func playNavSound() {
	now := time.Now()
	if now.Sub(lastNavSoundTime) < 90*time.Millisecond {
		return
	}
	lastNavSoundTime = now

	select {
	case beepChan <- soundNav:
	default:
	}
}

func playToggleSound() {
	select {
	case beepChan <- soundToggle:
	default:
	}
}

func playConfirmSound() {
	select {
	case beepChan <- soundConfirm:
	default:
	}
}

type rect struct {
	Left, Top, Right, Bottom int32
}

// Center the console window on the primary display screen
func centerWindow() {
	if runtime.GOOS != "windows" {
		return
	}

	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd == 0 {
		return
	}

	screenWidth, _, _ := procGetSystemMetrics.Call(uintptr(0))
	screenHeight, _, _ := procGetSystemMetrics.Call(uintptr(1))

	var r rect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))

	windowWidth := r.Right - r.Left
	windowHeight := r.Bottom - r.Top

	x := (int32(screenWidth) - windowWidth) / 2
	y := (int32(screenHeight) - windowHeight) / 2

	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	procSetWindowPos.Call(hwnd, 0, uintptr(x), uintptr(y), 0, 0, uintptr(0x0005))
}

// Styles matching the red header design
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	instructionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242"))

	activeItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	inactiveItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	checkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	uncheckStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	spinnerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	pendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	barFilledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	barUnfilledStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("238"))

	stepActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	stepInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))
)

const indent = "    "

func countSelected(selectedMap map[int]bool) int {
	c := 0
	for _, v := range selectedMap {
		if v {
			c++
		}
	}
	return c
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func getCatComment(m model) string {
	switch m.state {
	case stateBoot:
		return "Yawn... booting up Seppy!"

	case stateAppName:
		currentAppName := strings.TrimSpace(m.textInput)
		if currentAppName == "" {
			if m.collisionCount > 0 {
				targetDir := filepath.Join(m.parentDir, m.appName)
				if dirExists(targetDir) {
					if m.collisionCount >= 3 {
						return fmt.Sprintf("Seriously? '%s' is STILL taken. Read the room!", m.appName)
					}
					return fmt.Sprintf("'%s' exists! Pick a new name or I'll knock it off!", m.appName)
				}
			}
			return "Name this app? Make it pawsome!"
		}
		
		targetDir := filepath.Join(m.parentDir, currentAppName)
		if dirExists(targetDir) {
			if m.collisionCount >= 3 {
				return fmt.Sprintf("Seriously? '%s' is STILL taken. Read the room!", currentAppName)
			} else if m.collisionCount >= 1 {
				return fmt.Sprintf("'%s' exists! Pick a new name or I'll knock it off!", currentAppName)
			}
			return fmt.Sprintf("Hmm... '%s' smells like it already exists.", currentAppName)
		}
		return fmt.Sprintf("'%s'? Purr-fect choice!", currentAppName)

	case stateAddCustom:
		totalCustom := len(m.customSkillItems) + len(m.customDocPaths)
		if totalCustom == 0 {
			return "Got custom kibble? (skills/docs)"
		}
		return fmt.Sprintf("%d custom goodies stashed!", totalCustom)

	case stateSkills:
		cnt := countSelected(m.selectedSkills)
		total := len(m.availableSkills)
		if cnt == 0 {
			return "Zero NPMs? Living dangerously minimalist!"
		}
		if cnt == total && total > 0 {
			return "SELECT ALL?! Downloading the whole internet!"
		}
		if cnt <= 3 {
			return fmt.Sprintf("%d packages! Keeping it lean and mean!", cnt)
		}
		return fmt.Sprintf("%d packages chosen! Heavy duty stack!", cnt)

	case stateAgentSkills:
		cnt := countSelected(m.selectedAgentSkills)
		total := len(m.availableAgentSkills)
		if cnt == 0 {
			return "No AI skills? Going full manual mode!"
		}
		if cnt == total && total > 0 {
			return "ALL AI skills! Over 9000 power!"
		}
		return fmt.Sprintf("%d AI skills active! Automating everything!", cnt)

	case stateDocs:
		cnt := countSelected(m.selectedDocs)
		total := len(m.availableDocs)
		if cnt == 0 {
			return "No docs? Who reads manuals anyway!"
		}
		if cnt == total && total > 0 {
			return "ALL docs included! Knowledge overload!"
		}
		return fmt.Sprintf("%d markdown guides bundled!", cnt)

	case stateLocations:
		return "Sniffing out system config paths..."

	case stateExec:
		switch m.currentExecStep {
		case execCreateApp:
			return "Creating Next.js app... step 1!"
		case execCopyTemplates:
			return "Copying templates & skills like a ninja!"
		case execInstallDeps:
			return "npm install... time for a cat nap"
		case execRunCustomCmds:
			return "Running custom scripts... zoomies!"
		default:
			return "Coding super fast!"
		}

	case stateDone:
		if m.stepError != nil {
			return "Ouch! Setup hit a hairball."
		}
		return "All done! Now go build something pawsome!"

	default:
		return "Purr..."
	}
}

func getHeaderArt(m model) string {
	eyes := "-.-"
	switch (m.catTick / 4) % 8 {
	case 0, 1, 2:
		eyes = "-.-"
	case 3:
		eyes = "u.u"
	case 4, 5:
		eyes = "o.o"
	case 6:
		eyes = "o.x"
	case 7:
		eyes = "-.-"
	}

	zFrame := "zZZ"
	switch (m.catTick / 3) % 4 {
	case 0:
		zFrame = "z  "
	case 1:
		zFrame = "zZ "
	case 2:
		zFrame = "zZZ"
	case 3:
		zFrame = " Zz"
	}

	comment := getCatComment(m)

	showChars := m.catCommentTick * 2
	if showChars > len(comment) {
		showChars = len(comment)
	}
	
	var rightText string
	if m.idleTicks >= 100 {
		rightText = zFrame
	} else if showChars < len(comment) {
		rightText = comment[:showChars] + "█"
	} else {
		rightText = comment
	}

	catHead1 := "     /\\_/\\  "
	catHead2 := fmt.Sprintf("    ( %s )_", eyes)
	catHead3 := "    (   \"   )_"

	// 3 spaces gap between header logo and cat ASCII as requested
	gap := 3

	line1 := fmt.Sprintf("    ▄████▄ ██████ █████▄ █████▄ ██  ██%s%s%s", strings.Repeat(" ", gap), catHead1, rightText)
	line2 := fmt.Sprintf("    ██▄▄▄▄ ██▄▄   ██▄▄██ ██▄▄██ ▀████▀%s%s", strings.Repeat(" ", gap), catHead2)
	line3 := fmt.Sprintf("    ▄▄▄▄██ ██████ ██     ██       ██  %s%s", strings.Repeat(" ", gap), catHead3)

	return line1 + "\n" + line2 + "\n" + line3
}

type sessionState int

const (
	stateBoot sessionState = iota
	stateAppName
	stateAddCustom
	stateSkills
	stateAgentSkills
	stateDocs
	stateLocations
	stateConfirm
	stateExec
	stateDone
)

type execStep int

const (
	execCreateApp execStep = iota
	execCopyTemplates
	execInstallDeps
	execRunCustomCmds
	execFinished
)

type stepDoneMsg struct {
	step execStep
	err  error
}

type tickBootMsg time.Time

func doBootTick() tea.Cmd {
	return tea.Tick(35*time.Millisecond, func(t time.Time) tea.Msg {
		return tickBootMsg(t)
	})
}

type model struct {
	state                sessionState
	prevState            sessionState
	bootProgress         int
	appName              string
	cursor               int
	width                int
	height               int
	quitting             bool
	catTick              int
	catCommentTick       int
	lastCatComment       string
	idleTicks            int
	collisionCount       int

	availableSkills      []string
	selectedSkills       map[int]bool

	availableAgentSkills []string
	selectedAgentSkills  map[int]bool
	allSpeckitSkills     []string
	customCmdMap         map[string]string
	customNpmCmdMap      map[string]string

	availableDocs        []string
	selectedDocs         map[int]bool
	userDocsPathMap      map[string]string

	rootDir              string
	parentDir            string
	targetDir            string

	textInput            string

	// ADD CUSTOM SOURCES inline section
	customSkillInput     string
	customDocInput       string
	customNpmInput       string
	customActiveField    int
	customSkillItems     []string
	customSkillCmds      []string
	customDocPaths       []string
	customNpmPackages    []string

	// Viewport component for 100% full tab scrolling
	viewport             viewport.Model

	// Spinners
	spinner              spinner.Model
	rowSweepSpinner      spinner.Model
	currentExecStep      execStep
	stepStatus           map[execStep]string
	stepError            error

	config               UserConfig
	configPath           string
	cachePath            string
	docsPath             string
	exePath              string
}

func loadUserConfig() (UserConfig, string, string, string) {
	homeDir, err := os.UserHomeDir()
	seppyDir := filepath.Join(homeDir, ".seppy")
	if err != nil {
		seppyDir = ".seppy"
	}

	docsDir := filepath.Join(seppyDir, "docs")
	cacheDir := filepath.Join(seppyDir, "cache", "skills")
	os.MkdirAll(docsDir, 0755)
	os.MkdirAll(cacheDir, 0755)

	syncEmbeddedToSeppy(cacheDir, docsDir)

	configFile := filepath.Join(seppyDir, "config.json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		cfg := defaultUserConfig()
		saveUserConfig(configFile, cfg)
		return cfg, configFile, cacheDir, docsDir
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		cfg := defaultUserConfig()
		return cfg, configFile, cacheDir, docsDir
	}

	var cfg UserConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		cfg = defaultUserConfig()
		return cfg, configFile, cacheDir, docsDir
	}
	return cfg, configFile, cacheDir, docsDir
}

func syncEmbeddedToSeppy(cacheDir, docsDir string) {
	if entries, err := embeddedTemplates.ReadDir("template/.agents/skills"); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				srcEmbed := filepath.Join("template/.agents/skills", e.Name())
				dst := filepath.Join(cacheDir, e.Name())
				if _, err := os.Stat(dst); os.IsNotExist(err) {
					copyEmbeddedDir(embeddedTemplates, srcEmbed, dst)
				}
			}
		}
	}
	if entries, err := embeddedTemplates.ReadDir("template/docs"); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				srcEmbed := filepath.Join("template/docs", e.Name())
				dst := filepath.Join(docsDir, e.Name())
				if _, err := os.Stat(dst); os.IsNotExist(err) {
					copyEmbeddedFile(embeddedTemplates, srcEmbed, dst)
				}
			}
		}
	}
}

func saveUserConfig(configFile string, cfg UserConfig) {
	if data, err := json.MarshalIndent(cfg, "", "  "); err == nil {
		os.WriteFile(configFile, data, 0644)
	}
}

func defaultUserConfig() UserConfig {
	return UserConfig{
		DefaultUnselectAll:   false,
		CustomSkillsCommands: []CustomSkillCommand{},
		CustomNpmPackages:    []string{},
	}
}

func initialModel() model {
	userCfg, cfgPath, cachePath, docsPath := loadUserConfig()
	exePath, _ := os.Executable()
	initW, initH := getTerminalSize()

	parentDir, err := os.UserHomeDir()
	if err == nil {
		desktopPath := filepath.Join(parentDir, "Desktop")
		if info, err := os.Stat(desktopPath); err == nil && info.IsDir() {
			seppyFolder := filepath.Join(desktopPath, "SEPPY")
			os.MkdirAll(seppyFolder, 0755)
			parentDir = seppyFolder
		}
	} else {
		parentDir, _ = os.Getwd()
	}

	var agentSkills []string
	customCmdMap := make(map[string]string)
	var customSkillItems []string
	var customSkillCmds []string

	// 1. Load Custom Skills Commands from userConfig (~/.seppy/config.json)
	for _, cmd := range userCfg.CustomSkillsCommands {
		cleanCmd := sanitizeString(cmd.Command)
		cleanName := sanitizeString(cmd.Name)

		displayName := extractSkillName(cleanCmd)
		if (displayName == "skill" || strings.HasPrefix(displayName, "http")) && cleanName != "" {
			displayName = extractSkillName(cleanName)
		}

		agentSkills = append(agentSkills, displayName)
		customCmdMap[displayName] = cleanCmd
		customSkillItems = append(customSkillItems, displayName)
		customSkillCmds = append(customSkillCmds, cleanCmd)
	}

	// 2. Dynamically scan local cached skills from ~/.seppy/cache/skills/
	if entries, err := os.ReadDir(cachePath); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				name := e.Name()
				alreadyAdded := false
				for _, existing := range customSkillItems {
					if existing == name || sanitizeSlug(existing) == sanitizeSlug(name) {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					label := name
					cmdStr := filepath.Join(cachePath, name)
					agentSkills = append(agentSkills, label)
					customCmdMap[label] = cmdStr
					customSkillItems = append(customSkillItems, label)
					customSkillCmds = append(customSkillCmds, cmdStr)
				}
			}
		}
	}

	// 3. Dynamically scan user markdown docs from ~/.seppy/docs/
	var docsFiles []string
	userDocsPathMap := make(map[string]string)
	var customDocPaths []string

	if entries, err := os.ReadDir(docsPath); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
				name := e.Name()
				fullPath := filepath.Join(docsPath, name)
				docsFiles = append(docsFiles, name)
				userDocsPathMap[name] = fullPath
				customDocPaths = append(customDocPaths, fullPath)
			}
		}
	}

	var skills []string
	customNpmCmdMap := make(map[string]string)
	for _, raw := range userCfg.CustomNpmPackages {
		clean := sanitizeString(raw)
		if clean == "" {
			continue
		}
		displayName := extractNpmDisplayName(clean)
		skills = append(skills, displayName)
		customNpmCmdMap[displayName] = clean
	}

	s1 := spinner.New()
	s1.Spinner = spinner.Spinner{
		Frames: []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"},
		FPS:    time.Second / 10,
	}
	s1.Style = spinnerStyle

	s2 := spinner.New()
	s2.Spinner = spinner.Spinner{
		Frames: []string{"▸ ", "► ", "▶ ", "▹ "},
		FPS:    time.Second / 8,
	}
	s2.Style = spinnerStyle

	initVpH := initH - 12
	if initVpH < 4 {
		initVpH = 4
	}

	vp := viewport.New(initW, initVpH)
	vp.MouseWheelEnabled = true
	vp.MouseWheelDelta = 2

	m := model{
		state:                stateBoot,
		bootProgress:         0,
		appName:              "my-awesome-app",
		textInput:            "",
		customSkillInput:     "",
		customDocInput:       "",
		customActiveField:    0,
		customSkillItems:     customSkillItems,
		customSkillCmds:      customSkillCmds,
		customDocPaths:       customDocPaths,
		customNpmPackages:    userCfg.CustomNpmPackages,
		viewport:             vp,
		width:                initW,
		height:               initH,
		quitting:             false,
		catTick:              0,
		catCommentTick:       0,
		lastCatComment:       "",
		idleTicks:            0,
		availableSkills:      skills,
		selectedSkills:       make(map[int]bool),
		availableAgentSkills: agentSkills,
		selectedAgentSkills:  make(map[int]bool),
		customCmdMap:         customCmdMap,
		customNpmCmdMap:      customNpmCmdMap,
		availableDocs:        docsFiles,
		selectedDocs:         make(map[int]bool),
		userDocsPathMap:      userDocsPathMap,
		parentDir:            parentDir,
		spinner:              s1,
		rowSweepSpinner:      s2,
		currentExecStep:      execCreateApp,
		stepStatus: map[execStep]string{
			execCreateApp:     "pending",
			execCopyTemplates: "pending",
			execInstallDeps:   "pending",
			execRunCustomCmds: "pending",
		},
		config:     userCfg,
		configPath: cfgPath,
		cachePath:  cachePath,
		docsPath:   docsPath,
		exePath:    exePath,
	}

	defaultValue := !userCfg.DefaultUnselectAll
	for i := range skills {
		m.selectedSkills[i] = defaultValue
	}
	for i := range agentSkills {
		m.selectedAgentSkills[i] = defaultValue
	}
	for i := range docsFiles {
		m.selectedDocs[i] = defaultValue
	}

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.rowSweepSpinner.Tick, doBootTick())
}

func computeViewportHeight(h int) int {
	compact := h < 24
	veryCompact := h < 18
	used := 20
	if veryCompact {
		used = 10
	} else if compact {
		used = 13
	}
	
	vpH := h - used
	if vpH < 3 {
		vpH = 3
	}
	return vpH
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.viewport.Width = m.width
	m.viewport.Height = computeViewportHeight(m.height)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = computeViewportHeight(msg.Height)
		if m.state == stateAddCustom {
			m = updateCustomViewportContent(m)
		}

	case tea.MouseMsg:
		m.idleTicks = 0
		if msg.Type == tea.MouseWheelUp {
			switch m.state {
			case stateSkills:
				if m.cursor > 0 {
					m.cursor--
					playNavSound()
				}
			case stateAgentSkills:
				if m.cursor > 0 {
					m.cursor--
					playNavSound()
				}
			case stateDocs:
				if m.cursor > 0 {
					m.cursor--
					playNavSound()
				}
			}
		} else if msg.Type == tea.MouseWheelDown {
			switch m.state {
			case stateSkills:
				if m.cursor < len(m.availableSkills)-1 {
					m.cursor++
					playNavSound()
				}
			case stateAgentSkills:
				if m.cursor < len(m.availableAgentSkills)-1 {
					m.cursor++
					playNavSound()
				}
			case stateDocs:
				if m.cursor < len(m.availableDocs)-1 {
					m.cursor++
					playNavSound()
				}
			}
		}

		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd

	case tickBootMsg:
		if m.state == stateBoot {
			m.bootProgress += 2
			if m.bootProgress >= 100 {
				m.bootProgress = 100
				m.state = stateAppName
				playConfirmSound()
				return m, nil
			}
			return m, doBootTick()
		}

	case spinner.TickMsg:
		var cmd1, cmd2 tea.Cmd
		m.spinner, cmd1 = m.spinner.Update(msg)
		m.rowSweepSpinner, cmd2 = m.rowSweepSpinner.Update(msg)
		m.catTick++
		m.idleTicks++

		comment := getCatComment(m)
		if comment != m.lastCatComment {
			m.lastCatComment = comment
			m.catCommentTick = 0
		} else {
			m.catCommentTick++
		}
		if m.state == stateAddCustom {
			m = updateCustomViewportContent(m)
		}
		return m, tea.Batch(cmd1, cmd2)

	case stepDoneMsg:
		if msg.err != nil {
			m.stepStatus[msg.step] = "failed"
			m.stepError = msg.err
			m.state = stateDone
			return m, nil
		}

		m.stepStatus[msg.step] = "success"
		switch msg.step {
		case execCreateApp:
			m.currentExecStep = execCopyTemplates
			m.stepStatus[execCopyTemplates] = "running"
			return m, m.runStepCopyTemplates()

		case execCopyTemplates:
			m.currentExecStep = execInstallDeps
			m.stepStatus[execInstallDeps] = "running"
			return m, m.runStepInstallDeps()

		case execInstallDeps:
			m.currentExecStep = execRunCustomCmds
			m.stepStatus[execRunCustomCmds] = "running"
			return m, m.runStepRunCustomCmds()

		case execRunCustomCmds:
			m.currentExecStep = execFinished
			m.state = stateDone
			playConfirmSound()
			return m, nil
		}

	case tea.KeyMsg:
		m.idleTicks = 0
		switch msg.String() {
		case "ctrl+c":
			if m.state == stateLocations {
				m.state = m.prevState
				if m.state == stateAddCustom {
					m = updateCustomViewportContent(m)
				}
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit

		case "ctrl+l", "ctrl+L":
			if m.state != stateBoot && m.state != stateExec && m.state != stateLocations {
				m.prevState = m.state
				m.state = stateLocations
				playToggleSound()
				return m, nil
			}

		case "esc":
			switch m.state {
			case stateLocations:
				m.state = m.prevState
				if m.state == stateAddCustom {
					m = updateCustomViewportContent(m)
				}
				playNavSound()
				return m, nil
			case stateAppName:
				m.quitting = true
				return m, tea.Quit
			case stateAddCustom:
				m.state = stateAppName
				m.cursor = 0
				playNavSound()
				return m, nil
			case stateSkills:
				m.state = stateAppName
				m.cursor = 0
				playNavSound()
				return m, nil
			case stateAgentSkills:
				m.state = stateSkills
				m.cursor = 0
				playNavSound()
			case stateDocs:
				m.state = stateAgentSkills
				m.cursor = 0
				playNavSound()
				return m, nil
			case stateConfirm:
				m.state = stateDocs
				m.cursor = 0
				playNavSound()
				return m, nil
			}
		}

		switch m.state {
		case stateBoot:
			if msg.Type == tea.KeyEnter || msg.String() == " " {
				m.bootProgress = 100
				m.state = stateAppName
				return m, nil
			}

		case stateLocations:
			if msg.Type == tea.KeyEnter || msg.String() == "esc" {
				m.state = m.prevState
				if m.state == stateAddCustom {
					m = updateCustomViewportContent(m)
				}
				playNavSound()
			}

		case stateAppName:
			switch msg.String() {
			case "tab":
				if strings.TrimSpace(m.textInput) != "" {
					m.appName = strings.TrimSpace(m.textInput)
				}
				m.targetDir = filepath.Join(m.parentDir, m.appName)
				m.state = stateAddCustom
				m.customActiveField = 0
				m = updateCustomViewportContent(m)
				playToggleSound()
			default:
				switch msg.Type {
				case tea.KeyEnter:
					if strings.TrimSpace(m.textInput) != "" {
						m.appName = strings.TrimSpace(m.textInput)
					}
					m.targetDir = filepath.Join(m.parentDir, m.appName)
					if dirExists(m.targetDir) {
						m.idleTicks = 0
						m.catCommentTick = 0
						m.collisionCount++
						playToggleSound()
						return m, nil
					}
					m.state = stateSkills
					m.cursor = 0
					playConfirmSound()
				case tea.KeyBackspace, tea.KeyDelete:
					m.collisionCount = 0
					if len(m.textInput) > 0 {
						m.textInput = m.textInput[:len(m.textInput)-1]
					}
				case tea.KeyRunes:
					m.collisionCount = 0
					m.textInput += string(msg.Runes)
				}
			}

		case stateAddCustom:
			switch msg.String() {
			case "tab":
				m.customActiveField = (m.customActiveField + 1) % 3
				if m.customActiveField == 0 {
					m.viewport.GotoTop()
				} else if m.customActiveField == 1 {
					offset := 6
					if len(m.customSkillCmds) > 0 {
						offset += 1 + len(m.customSkillCmds)
					}
					m.viewport.SetYOffset(offset)
				} else {
					m.viewport.GotoBottom()
				}
				m = updateCustomViewportContent(m)
				playToggleSound()

			case "up", "k":
				m.viewport.LineUp(1)
				return m, nil

			case "down", "j":
				m.viewport.LineDown(1)
				return m, nil

			case "pgup":
				m.viewport.HalfViewUp()
				return m, nil

			case "pgdown":
				m.viewport.HalfViewDown()
				return m, nil

			case "enter":
				if m.customActiveField == 0 {
					trimmed := strings.TrimSpace(m.customSkillInput)
					if trimmed != "" {
						label := extractSkillName(trimmed)
						m.customSkillItems = append(m.customSkillItems, label)
						m.customSkillCmds = append(m.customSkillCmds, trimmed)
						offset := len(m.availableAgentSkills)
						m.availableAgentSkills = append(m.availableAgentSkills, label)
						m.customCmdMap[label] = trimmed
						m.selectedAgentSkills[offset] = true

						m.config.CustomSkillsCommands = append(m.config.CustomSkillsCommands, CustomSkillCommand{
							Name:    label,
							Command: trimmed,
						})
						saveUserConfig(m.configPath, m.config)

						m.customSkillInput = ""
						m = updateCustomViewportContent(m)
						playToggleSound()
					} else if strings.TrimSpace(m.customDocInput) == "" && strings.TrimSpace(m.customNpmInput) == "" {
						m.state = stateSkills
						m.cursor = 0
						playConfirmSound()
					}
				} else if m.customActiveField == 1 {
					trimmed := strings.TrimSpace(m.customDocInput)
					if trimmed != "" {
						docName := filepath.Base(trimmed)
						m.customDocPaths = append(m.customDocPaths, trimmed)
						offset := len(m.availableDocs)
						m.availableDocs = append(m.availableDocs, docName)
						m.userDocsPathMap[docName] = trimmed
						m.selectedDocs[offset] = true
						m.customDocInput = ""
						m = updateCustomViewportContent(m)
						playToggleSound()
					} else if strings.TrimSpace(m.customSkillInput) == "" && strings.TrimSpace(m.customNpmInput) == "" {
						m.state = stateSkills
						m.cursor = 0
						playConfirmSound()
					}
				} else {
					trimmed := sanitizeString(m.customNpmInput)
					if trimmed != "" {
						displayName := extractNpmDisplayName(trimmed)
						alreadyAdded := false
						for _, pkg := range m.availableSkills {
							if strings.EqualFold(pkg, displayName) {
								alreadyAdded = true
								break
							}
						}
						if !alreadyAdded {
							m.customNpmPackages = append(m.customNpmPackages, trimmed)
							offset := len(m.availableSkills)
							m.availableSkills = append(m.availableSkills, displayName)
							m.customNpmCmdMap[displayName] = trimmed
							m.selectedSkills[offset] = true

							m.config.CustomNpmPackages = append(m.config.CustomNpmPackages, trimmed)
							saveUserConfig(m.configPath, m.config)
						}
						m.customNpmInput = ""
						m = updateCustomViewportContent(m)
						playToggleSound()
					} else if strings.TrimSpace(m.customSkillInput) == "" && strings.TrimSpace(m.customDocInput) == "" {
						m.state = stateSkills
						m.cursor = 0
						playConfirmSound()
					}
				}

			default:
				switch msg.Type {
				case tea.KeyBackspace, tea.KeyDelete:
					if m.customActiveField == 0 {
						if len(m.customSkillInput) > 0 {
							m.customSkillInput = m.customSkillInput[:len(m.customSkillInput)-1]
							m = updateCustomViewportContent(m)
						}
					} else if m.customActiveField == 1 {
						if len(m.customDocInput) > 0 {
							m.customDocInput = m.customDocInput[:len(m.customDocInput)-1]
							m = updateCustomViewportContent(m)
						}
					} else {
						if len(m.customNpmInput) > 0 {
							m.customNpmInput = m.customNpmInput[:len(m.customNpmInput)-1]
							m = updateCustomViewportContent(m)
						}
					}
				case tea.KeyRunes:
					if m.customActiveField == 0 {
						m.customSkillInput += string(msg.Runes)
					} else if m.customActiveField == 1 {
						m.customDocInput += string(msg.Runes)
					} else {
						m.customNpmInput += string(msg.Runes)
					}
					m = updateCustomViewportContent(m)
				}
			}

		case stateSkills:
			m.handleMenuNavigation(msg, m.availableSkills, m.selectedSkills, stateAgentSkills)

		case stateAgentSkills:
			m.handleMenuNavigation(msg, m.availableAgentSkills, m.selectedAgentSkills, stateDocs)

		case stateDocs:
			m.handleMenuNavigation(msg, m.availableDocs, m.selectedDocs, stateConfirm)
			
		case stateConfirm:
			switch msg.String() {
			case "enter", "y", "Y":
				m.state = stateExec
				m.stepStatus[execCreateApp] = "running"
				playConfirmSound()
				return m, m.runStepCreateApp()
			case "esc":
				m.state = stateDocs
				m.cursor = 0
				playNavSound()
				return m, nil
			}

		case stateDone:
			switch msg.String() {
			case " ":
				if runtime.GOOS == "windows" {
					exec.Command("explorer.exe", m.targetDir).Start()
				} else {
					exec.Command("open", m.targetDir).Start()
				}
				playConfirmSound()
				return m, nil
			case "enter":
				closeParentTerminal()
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func closeParentTerminal() {
	if runtime.GOOS == "windows" {
		ppid := os.Getppid()
		exec.Command("cmd", "/c", fmt.Sprintf("taskkill /F /PID %d 2>NUL", ppid)).Run()
	}
}

func (m *model) handleMenuNavigation(msg tea.KeyMsg, items []string, selectedMap map[int]bool, nextState sessionState) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
			playNavSound()
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
			playNavSound()
		}
	case " ":
		selectedMap[m.cursor] = !selectedMap[m.cursor]
		playToggleSound()
	case "a", "A":
		for i := range items {
			selectedMap[i] = true
		}
		playToggleSound()
	case "n", "N":
		for i := range items {
			selectedMap[i] = false
		}
		playToggleSound()
	case "enter":
		m.state = nextState
		m.cursor = 0
		playConfirmSound()
	}
}

// Execution commands with non-interactive timeouts and central caching
func (m model) runStepCreateApp() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		cmdApp := exec.CommandContext(ctx, "npx", "--yes", "create-next-app@latest", m.targetDir, "--yes")
		cmdApp.Stdin = strings.NewReader("\n")
		cmdApp.Env = append(os.Environ(), "CI=1", "NONINTERACTIVE=1")
		err := cmdApp.Run()
		return stepDoneMsg{step: execCreateApp, err: err}
	}
}

func (m model) runStepCopyTemplates() tea.Cmd {
	return func() tea.Msg {
		targetSkillsDir := filepath.Join(m.targetDir, ".agents", "skills")
		targetDocsDir := filepath.Join(m.targetDir, "docs")

		// Copy selected agent skills dynamically from ~/.seppy/cache/skills
		for i, skill := range m.availableAgentSkills {
			if m.selectedAgentSkills[i] {
				if cachedPath, exists := m.customCmdMap[skill]; exists {
					if _, err := os.Stat(cachedPath); err == nil {
						skillSlug := sanitizeSlug(skill)
						targetSkillDir := filepath.Join(targetSkillsDir, skillSlug)
						copyDir(cachedPath, targetSkillDir)
					}
				}
			}
		}

		// Copy selected markdown docs dynamically from ~/.seppy/docs
		for i, docFile := range m.availableDocs {
			if m.selectedDocs[i] {
				if userPath, isUserDoc := m.userDocsPathMap[docFile]; isUserDoc {
					copyFile(userPath, filepath.Join(targetDocsDir, docFile))
				}
			}
		}

		return stepDoneMsg{step: execCopyTemplates, err: nil}
	}
}

func (m model) runStepInstallDeps() tea.Cmd {
	return func() tea.Msg {
		os.Chdir(m.targetDir)

		for i, s := range m.availableSkills {
			if m.selectedSkills[i] {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				var cmd *exec.Cmd

				switch {
				case strings.HasPrefix(s, "Prettier"):
					cmd = exec.CommandContext(ctx, "npm", "install", "-D", "prettier", "@laststance/tailwind-suggest-canonical-classes", "prettier-plugin-tailwindcss-canonical-classes", "--no-audit", "--no-fund")
					cmd.Run()

					pkgPath := filepath.Join(m.targetDir, "package.json")
					if data, err := os.ReadFile(pkgPath); err == nil {
						strData := string(data)
						strData = strings.Replace(strData, `"scripts": {`, `"scripts": {`+"\n    \"lint:tailwind\": \"tailwind-suggest-canonical-classes \\\"src/**/*.{ts,tsx,js,jsx,html}\\\"\",\n    \"format:fix\": \"prettier --write \\\"src/**/*.{ts,tsx,js,jsx,html}\\\"\",", 1)
						os.WriteFile(pkgPath, []byte(strData), 0644)
					}

					prettierrc := "{\n  \"plugins\": [\"prettier-plugin-tailwindcss-canonical-classes\"]\n}\n"
					os.WriteFile(filepath.Join(m.targetDir, ".prettierrc"), []byte(prettierrc), 0644)

				case strings.HasPrefix(s, "Lenis"):
					cmd = exec.CommandContext(ctx, "npm", "install", "lenis", "--no-audit", "--no-fund")
					cmd.Run()

				case strings.HasPrefix(s, "Lucide"):
					cmd = exec.CommandContext(ctx, "npm", "install", "lucide-react", "--no-audit", "--no-fund")
					cmd.Run()

				case strings.HasPrefix(s, "Framer Motion"):
					cmd = exec.CommandContext(ctx, "npm", "install", "framer-motion", "--no-audit", "--no-fund")
					cmd.Run()

				case strings.HasPrefix(s, "Zustand"):
					cmd = exec.CommandContext(ctx, "npm", "install", "zustand", "--no-audit", "--no-fund")
					cmd.Run()

				default:
					rawCmd := s
					if mappedCmd, ok := m.customNpmCmdMap[s]; ok {
						rawCmd = mappedCmd
					}
					
					fields := strings.Fields(rawCmd)
					if len(fields) == 1 {
						cmd = exec.CommandContext(ctx, "npm", "install", fields[0], "--no-audit", "--no-fund")
					} else {
						cmd = exec.CommandContext(ctx, fields[0], fields[1:]...)
					}
					cmd.Run()
				}
				cancel()
			}
		}

		return stepDoneMsg{step: execInstallDeps, err: nil}
	}
}

func (m model) runStepRunCustomCmds() tea.Cmd {
	return func() tea.Msg {
		os.Chdir(m.targetDir)

		for i, skill := range m.availableAgentSkills {
			if m.selectedAgentSkills[i] {
				if rawCmd, isCustom := m.customCmdMap[skill]; isCustom {
					skillSlug := sanitizeSlug(skill)
					cachedSkillDir := filepath.Join(m.cachePath, skillSlug)

					if _, err := os.Stat(cachedSkillDir); os.IsNotExist(err) {
						os.MkdirAll(cachedSkillDir, 0755)
						parts := strings.Fields(rawCmd)
						if len(parts) > 0 {
							ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
							cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
							cmd.Dir = cachedSkillDir
							cmd.Stdin = strings.NewReader("\n")
							cmd.Env = append(os.Environ(), "CI=1", "NONINTERACTIVE=1")
							cmd.Run()
							cancel()
						}
					}

					targetSkillDir := filepath.Join(m.targetDir, ".agents", "skills", skillSlug)
					if _, err := os.Stat(cachedSkillDir); err == nil {
						copyDir(cachedSkillDir, targetSkillDir)
					}
				}
			}
		}

		return stepDoneMsg{step: execRunCustomCmds, err: nil}
	}
}

func sanitizeString(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(s)
}

func extractSkillName(raw string) string {
	raw = sanitizeString(raw)
	if raw == "" {
		return "skill"
	}

	if idx := strings.Index(raw, "--skill"); idx != -1 {
		part := raw[idx+7:]
		part = strings.TrimPrefix(part, "=")
		part = strings.TrimSpace(part)
		fields := strings.Fields(part)
		if len(fields) > 0 && !strings.HasPrefix(fields[0], "-") {
			return fields[0]
		}
	}

	fields := strings.Fields(raw)
	for _, f := range fields {
		if strings.HasPrefix(f, "http://") || strings.HasPrefix(f, "https://") {
			u := strings.TrimRight(f, "/")
			base := filepath.Base(u)
			if base != "" && base != "." && base != "/" {
				return base
			}
		}
	}

	if strings.Contains(raw, "\\") || strings.Contains(raw, "/") {
		base := filepath.Base(raw)
		if base != "" && base != "." && base != "/" {
			return base
		}
	}

	raw = strings.TrimPrefix(raw, "npx skills add")
	raw = strings.TrimSpace(raw)

	if len(raw) > 30 {
		return raw[:30]
	}
	return raw
}

func extractNpmDisplayName(raw string) string {
	raw = sanitizeString(raw)
	if raw == "" {
		return ""
	}
	raw = strings.TrimPrefix(raw, "npm install")
	raw = strings.TrimPrefix(raw, "npm i")
	raw = strings.TrimPrefix(raw, "yarn add")
	raw = strings.TrimPrefix(raw, "pnpm add")
	raw = strings.TrimPrefix(raw, "pnpm install")
	raw = strings.TrimPrefix(raw, "bun add")
	raw = strings.TrimPrefix(raw, "bun install")
	raw = strings.TrimSpace(raw)
	
	fields := strings.Fields(raw)
	if len(fields) > 0 {
		return fields[0]
	}
	return raw
}

func sanitizeSlug(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	return strings.Trim(s, "-")
}

func truncateStr(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	if len(s) > maxLen {
		if maxLen > 3 {
			return s[:maxLen-3] + "..."
		}
		return s[:maxLen]
	}
	return s
}

func (m model) renderStepBar() string {
	s1 := stepInactiveStyle.Render("1. Project Setup")
	s2 := stepInactiveStyle.Render("2. Dependencies")
	s3 := stepInactiveStyle.Render("3. Agent Skills")
	s4 := stepInactiveStyle.Render("4. Markdown Docs")

	switch m.state {
	case stateAppName, stateAddCustom:
		s1 = stepActiveStyle.Render("1. Project Setup")
	case stateSkills:
		s2 = stepActiveStyle.Render("2. Dependencies")
	case stateAgentSkills:
		s3 = stepActiveStyle.Render("3. Agent Skills")
	case stateDocs:
		s4 = stepActiveStyle.Render("4. Markdown Docs")
	}

	return fmt.Sprintf("%s ► %s ► %s ► %s", s1, s2, s3, s4)
}

func updateCustomViewportContent(m model) model {
	var tabContent strings.Builder
	tabContent.WriteString(indent + instructionStyle.Render("[Tab] switch field • ↑/↓ scroll • [Enter] add • empty [Enter] continue • [ESC] back") + "\n\n")

	if m.customActiveField == 0 {
		tabContent.WriteString(indent + activeItemStyle.Render("[+] Skill Install Command") + "\n")
		tabContent.WriteString(indent + instructionStyle.Render("    e.g. npx skills add https://github.com/org/skills --skill name --agy") + "\n")
		tabContent.WriteString(indent + "    " + m.spinner.View() + " " + activeItemStyle.Render(m.customSkillInput) + "█\n")
	} else {
		tabContent.WriteString(indent + uncheckStyle.Render("[ ] Skill Install Command") + "\n")
		tabContent.WriteString(indent + instructionStyle.Render("    e.g. npx skills add https://github.com/org/skills --skill name --agy") + "\n")
		tabContent.WriteString(indent + "    " + "  " + inactiveItemStyle.Render(m.customSkillInput) + "\n")
	}

	if len(m.customSkillCmds) > 0 {
		tabContent.WriteString(indent + infoStyle.Render("    Installed / Added Skill Commands:") + "\n")
		maxLen := m.width - 25
		if maxLen < 20 {
			maxLen = 20
		}

		for _, s := range m.customSkillItems {
			truncatedLine := truncateStr(s, maxLen)
			tabContent.WriteString(indent + "      " + successStyle.Render("✓ ") + activeItemStyle.Render(truncatedLine) + "\n")
		}
	}

	tabContent.WriteString("\n")

	if m.customActiveField == 1 {
		tabContent.WriteString(indent + activeItemStyle.Render("[+] Markdown File Path (.md)") + "\n")
		tabContent.WriteString(indent + instructionStyle.Render("    e.g. C:\\Users\\you\\docs\\MY_GUIDE.md") + "\n")
		tabContent.WriteString(indent + "    " + m.spinner.View() + " " + activeItemStyle.Render(m.customDocInput) + "█\n")
	} else {
		tabContent.WriteString(indent + uncheckStyle.Render("[ ] Markdown File Path (.md)") + "\n")
		tabContent.WriteString(indent + instructionStyle.Render("    e.g. C:\\Users\\you\\docs\\MY_GUIDE.md") + "\n")
		tabContent.WriteString(indent + "    " + "  " + inactiveItemStyle.Render(m.customDocInput) + "\n")
	}
	if len(m.customDocPaths) > 0 {
		for _, d := range m.customDocPaths {
			tabContent.WriteString(indent + "      " + successStyle.Render("✓ ") + inactiveItemStyle.Render(filepath.Base(d)) + "\n")
		}
	}

	tabContent.WriteString("\n")

	if m.customActiveField == 2 {
		tabContent.WriteString(indent + activeItemStyle.Render("[+] Custom NPM Package Name") + "\n")
		tabContent.WriteString(indent + instructionStyle.Render("    e.g. npm install framer-motion") + "\n")
		tabContent.WriteString(indent + "    " + m.spinner.View() + " " + activeItemStyle.Render(m.customNpmInput) + "█\n")
	} else {
		tabContent.WriteString(indent + uncheckStyle.Render("[ ] Custom NPM Package Name") + "\n")
		tabContent.WriteString(indent + instructionStyle.Render("    e.g. npm install framer-motion") + "\n")
		tabContent.WriteString(indent + "    " + "  " + inactiveItemStyle.Render(m.customNpmInput) + "\n")
	}
	if len(m.customNpmPackages) > 0 {
		for _, pkg := range m.customNpmPackages {
			tabContent.WriteString(indent + "      " + successStyle.Render("✓ ") + inactiveItemStyle.Render(pkg) + "\n")
		}
	}

	m.viewport.SetContent(tabContent.String())
	return m
}

func (m model) View() string {
	if m.quitting {
		return "\n" + indent + "Setup cancelled.\n"
	}

	var b strings.Builder

	compact := m.height < 24
	veryCompact := m.height < 18
	
	nl2 := "\n\n"
	if compact {
		nl2 = "\n"
	}

	if veryCompact {
		b.WriteString("\n" + headerStyle.Render(" SEPPY CLI ") + "\n")
	} else if compact {
		b.WriteString("\n" + headerStyle.Render(getHeaderArt(m)) + "\n")
		b.WriteString(indent + titleStyle.Render("All-in-One Next.js Project Generator") + "\n")
	} else {
		b.WriteString("\n" + headerStyle.Render(getHeaderArt(m)) + "\n\n")
		b.WriteString(indent + titleStyle.Render("All-in-One Next.js Project Generator") + "\n")
		b.WriteString(indent + infoStyle.Render("Automates app creation with curated NPM packages, local agent skills (.agents),") + "\n")
		b.WriteString(indent + infoStyle.Render("and architectural markdown helpers.") + "\n\n")
	}

	if m.state != stateBoot && m.state != stateExec && m.state != stateDone && m.state != stateLocations && m.state != stateConfirm {
		b.WriteString(indent + m.renderStepBar() + nl2)
	}

	m.viewport.Width = m.width
	m.viewport.Height = computeViewportHeight(m.height)

	switch m.state {
	case stateBoot:
		b.WriteString(indent + titleStyle.Render("SYSTEM INITIALIZING") + "\n")

		barWidth := 28
		filled := (m.bootProgress * barWidth) / 100
		unfilled := barWidth - filled
		barStr := barFilledStyle.Render(strings.Repeat("█", filled)) + barUnfilledStyle.Render(strings.Repeat("░", unfilled))

		var statusText string
		switch {
		case m.bootProgress < 18:
			statusText = "Scanning template repository..."
		case m.bootProgress < 42:
			statusText = "Loading local agent skills (.agents)..."
		case m.bootProgress < 68:
			statusText = "Parsing markdown helper docs (docs/*.md)..."
		case m.bootProgress < 88:
			statusText = "Initializing TUI engine..."
		case m.bootProgress < 100:
			statusText = "Calibrating sound synthesizer..."
		default:
			statusText = "System Ready."
		}

		line := fmt.Sprintf("[%3d%%] [%s] %s", m.bootProgress, barStr, statusText)
		b.WriteString(indent + line + "\n\n")
		b.WriteString(indent + instructionStyle.Render("(Press [Enter] to skip boot animation)") + "\n")

	case stateLocations:
		b.WriteString(indent + titleStyle.Render("SYSTEM LOCATIONS (READ-ONLY)") + "\n")
		b.WriteString(indent + instructionStyle.Render("All system environment configurations and directory paths:") + "\n\n")

		b.WriteString(indent + activeItemStyle.Render("Configuration File:") + "\n")
		b.WriteString(indent + "  " + inactiveItemStyle.Render(m.configPath) + "\n\n")

		b.WriteString(indent + activeItemStyle.Render("Skills Cache Directory:") + "\n")
		b.WriteString(indent + "  " + inactiveItemStyle.Render(m.cachePath) + "\n\n")

		b.WriteString(indent + activeItemStyle.Render("Custom Markdown Docs Directory:") + "\n")
		b.WriteString(indent + "  " + inactiveItemStyle.Render(m.docsPath) + "\n\n")

		b.WriteString(indent + activeItemStyle.Render("Executable Location:") + "\n")
		b.WriteString(indent + "  " + inactiveItemStyle.Render(m.exePath) + "\n\n")

		b.WriteString(indent + instructionStyle.Render("Press [ESC] or [Enter] to return.") + "\n")

	case stateAppName, stateAddCustom:
		b.WriteString(indent + titleStyle.Render("PROJECT SETUP") + "\n")
		b.WriteString(indent + instructionStyle.Render("Enter project name below and press [Enter]:") + nl2)

		if m.state == stateAddCustom {
			b.WriteString(indent + "  " + inactiveItemStyle.Render(m.appName) + nl2)
			b.WriteString(m.viewport.View())
		} else {
			nameDisplay := m.textInput
			b.WriteString(indent + m.spinner.View() + " " + activeItemStyle.Render(nameDisplay) + "█" + nl2)

			currentAppName := strings.TrimSpace(m.textInput)
			if currentAppName == "" {
				currentAppName = m.appName
			}
		}

		hintLine := indent + instructionStyle.Render("(Default: my-awesome-app) • [Tab] custom sources • [Ctrl+L] locations")
		if m.height >= 32 {
			b.WriteString("\n" + hintLine + "\n")
		} else {
			currentLines := strings.Count(b.String(), "\n")
			// Reserve 1 blank line at bottom so top & bottom padding are symmetrical
			padLines := (m.height - 2) - currentLines
			if padLines > 0 {
				b.WriteString(strings.Repeat("\n", padLines))
			}
			b.WriteString(hintLine + "\n")
		}

	case stateSkills:
		m.renderMenu(&b, "SELECT NPM DEPENDENCIES", m.availableSkills, m.selectedSkills)

	case stateAgentSkills:
		m.renderMenu(&b, "SELECT AGENT SKILLS (.agents/skills)", m.availableAgentSkills, m.selectedAgentSkills)

	case stateDocs:
		m.renderMenu(&b, "SELECT MARKDOWN HELPER DOCS (docs/*.md)", m.availableDocs, m.selectedDocs)

	case stateConfirm:
		b.WriteString(indent + titleStyle.Render("CONFIRM SETUP") + "\n")
		b.WriteString(indent + instructionStyle.Render("Review your project settings before continuing:") + nl2)
		b.WriteString(indent + activeItemStyle.Render("Project Name: ") + inactiveItemStyle.Render(m.appName) + "\n")
		b.WriteString(indent + activeItemStyle.Render("Target Dir:   ") + inactiveItemStyle.Render(m.targetDir) + "\n\n")

		npmCount := countSelected(m.selectedSkills)
		agentCount := countSelected(m.selectedAgentSkills)
		docCount := countSelected(m.selectedDocs)

		b.WriteString(indent + activeItemStyle.Render("NPM Packages:  ") + inactiveItemStyle.Render(fmt.Sprintf("%d selected", npmCount)) + "\n")
		b.WriteString(indent + activeItemStyle.Render("Agent Skills:  ") + inactiveItemStyle.Render(fmt.Sprintf("%d selected", agentCount)) + "\n")
		b.WriteString(indent + activeItemStyle.Render("Markdown Docs: ") + inactiveItemStyle.Render(fmt.Sprintf("%d selected", docCount)) + "\n\n")

		b.WriteString(indent + instructionStyle.Render("Press [Enter] or [Y] to start execution.") + "\n")
		b.WriteString(indent + instructionStyle.Render("Press [ESC] to go back and edit.") + "\n")

	case stateExec, stateDone:
		b.WriteString(indent + titleStyle.Render("EXECUTING SETUP") + "\n")
		b.WriteString(indent + instructionStyle.Render("Target Directory: "+m.targetDir) + "\n\n")

		m.renderStepStatus(&b, execCreateApp, "Creating Next.js Application")
		m.renderStepStatus(&b, execCopyTemplates, "Copying Template Assets, Agent Skills & Markdown Docs")
		m.renderStepStatus(&b, execInstallDeps, "Installing Selected NPM Dependencies")
		m.renderStepStatus(&b, execRunCustomCmds, "Caching & Fast-Copying Custom Agent Skills")

		if m.state == stateDone {
			if m.stepError != nil {
				b.WriteString("\n" + indent + headerStyle.Render("X Setup failed: "+m.stepError.Error()) + "\n")
				b.WriteString("\n" + indent + instructionStyle.Render("Press [Enter] or [Q] to exit.") + "\n")
			} else {
				b.WriteString("\n" + indent + successStyle.Render("✓ Setup Complete for '"+m.appName+"'!") + "\n")
				b.WriteString(indent + activeItemStyle.Render("Output Location: "+m.targetDir) + "\n\n")
				b.WriteString(indent + instructionStyle.Render("Press [Space] to open project folder • Press [Enter] or [Q] to exit") + "\n")
			}
		}
	}

	return b.String()
}

func (m model) renderStepStatus(b *strings.Builder, step execStep, label string) {
	status := m.stepStatus[step]
	var icon string

	switch status {
	case "pending":
		icon = pendingStyle.Render("○ ") + pendingStyle.Render(label)
	case "running":
		icon = m.spinner.View() + " " + activeItemStyle.Render(label+"...")
	case "success":
		icon = successStyle.Render("✓ ") + inactiveItemStyle.Render(label)
	case "failed":
		icon = headerStyle.Render("X ") + headerStyle.Render(label)
	}

	b.WriteString(indent + icon + "\n")
}

func (m model) renderMenu(b *strings.Builder, title string, items []string, selectedMap map[int]bool) {
	compact := m.height < 24
	nl2 := "\n\n"
	if compact {
		nl2 = "\n"
	}

	b.WriteString(indent + titleStyle.Render(title) + "\n")
	b.WriteString(indent + instructionStyle.Render("↑/↓ navigate • [Space] toggle • [A] select all • [N] clear • [ESC] back • [Enter] confirm") + nl2)

	usedLines := 17
	if compact {
		usedLines = 11
	}
	if m.height < 18 {
		usedLines = 8
	}
	maxVisible := m.height - usedLines
	if maxVisible < 3 {
		maxVisible = 3
	}

	start := 0
	end := len(items)

	if len(items) > maxVisible {
		start = m.cursor - (maxVisible / 2)
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(items) {
			end = len(items)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	for i := start; i < end; i++ {
		item := items[i]

		prefix := "  "
		if m.cursor == i {
			prefix = m.rowSweepSpinner.View()
		}

		check := uncheckStyle.Render("[ ]")
		if selectedMap[i] {
			check = checkStyle.Render("[X]")
		}

		line := fmt.Sprintf("%s%s %s", prefix, check, item)
		if m.cursor == i {
			b.WriteString(indent + activeItemStyle.Render(line) + "\n")
		} else {
			b.WriteString(indent + inactiveItemStyle.Render(line) + "\n")
		}
	}
}

func main() {
	startAudioWorker()
	centerWindow()
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// Embedded Template Helpers
func copyEmbeddedDir(fsys embed.FS, srcDir, dstDir string) error {
	return fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dstDir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		return copyEmbeddedFile(fsys, path, target)
	})
}

func copyEmbeddedFile(fsys embed.FS, srcFile, dstFile string) error {
	in, err := fsys.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	os.MkdirAll(filepath.Dir(dstFile), 0755)
	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}

	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}

	return out.Close()
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	os.MkdirAll(filepath.Dir(dst), 0755)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err = io.Copy(out, in); err != nil {
		out.Close()
		return err
	}

	if err := out.Sync(); err != nil {
		out.Close()
		return err
	}

	return out.Close()
}
