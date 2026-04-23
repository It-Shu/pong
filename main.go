package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	minFieldWidth  = 24
	minFieldHeight = 10
	paddleHeight   = 4
	targetScore    = 7
	tickRate       = 16 * time.Millisecond
	loaderTickRate = 45 * time.Millisecond
	paddleStep     = 0.40
	subPixelX      = 2
	subPixelY      = 4
	baseBallSpeedX = 0.30
	baseBallSpeedY = 0.10
	maxBallSpeedX  = 0.42
	hitBoostMax    = 0.05
)

type gameState int

const (
	stateLoading gameState = iota
	stateMenu
	stateWaiting
	stateRunning
	statePaused
	stateGameOver
)

type tickMsg time.Time
type loaderTickMsg struct{}

type aiDifficulty struct {
	name             string
	label            string
	playerWindow     string
	moveStep         float64
	reactionFrames   int
	trackFromX       float64
	returnBlend      float64
	aimError         float64
	deadZone         float64
	repositionWindow float64
}

var aiDifficulties = []aiDifficulty{
	{
		name:             "easy",
		label:            "Easy",
		playerWindow:     "Player edge: about 70%",
		moveStep:         0.16,
		reactionFrames:   3,
		trackFromX:       0.72,
		returnBlend:      0.0,
		aimError:         1.15,
		deadZone:         0.35,
		repositionWindow: 0.0,
	},
	{
		name:             "medium",
		label:            "Medium",
		playerWindow:     "Player edge: about 50%",
		moveStep:         0.20,
		reactionFrames:   2,
		trackFromX:       0.82,
		returnBlend:      0.45,
		aimError:         0.65,
		deadZone:         0.24,
		repositionWindow: 0.18,
	},
	{
		name:             "hard",
		label:            "Hard",
		playerWindow:     "Player edge: about 20%",
		moveStep:         0.21,
		reactionFrames:   1,
		trackFromX:       0.90,
		returnBlend:      0.82,
		aimError:         0.40,
		deadZone:         0.18,
		repositionWindow: 0.26,
	},
}

type keyMap struct {
	RightUp   key.Binding
	RightDown key.Binding
	Start     key.Binding
	Pause     key.Binding
	Restart   key.Binding
	Menu      key.Binding
	Help      key.Binding
	Quit      key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.RightUp, k.RightDown, k.Start, k.Pause, k.Restart, k.Menu, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.RightUp, k.RightDown, k.Start, k.Pause},
		{k.Restart, k.Menu, k.Help, k.Quit},
	}
}

type model struct {
	width  int
	height int

	fieldWidth  int
	fieldHeight int

	leftY    float64
	rightY   float64
	ballX    float64
	ballY    float64
	velX     float64
	velY     float64
	leftVel  float64
	rightVel float64

	leftScore  int
	rightScore int

	state      gameState
	loadingPct float64
	menuIndex  int
	aiAnchorY  float64
	aiTick     int
	animTick   int
	flashText  []string
	flashTimer int

	keys        keyMap
	help        help.Model
	showAllHelp bool
	spinner     spinner.Model
	progress    progress.Model

	frameStyle  lipgloss.Style
	titleStyle  lipgloss.Style
	scoreStyle  lipgloss.Style
	statusStyle lipgloss.Style
	helpStyle   lipgloss.Style
	bannerStyle lipgloss.Style
	flashStyle  lipgloss.Style
	ballStyle   lipgloss.Style
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "pong terminal error: %v\n", err)
		os.Exit(1)
	}
}

func newModel() model {
	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	pr := progress.New(
		progress.WithDefaultGradient(),
		progress.WithScaledGradient("#5E81AC", "#A3BE8C"),
	)
	pr.Width = 18

	keys := keyMap{
		RightUp: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp("↑", "up"),
		),
		RightDown: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp("↓", "down"),
		),
		Start: key.NewBinding(
			key.WithKeys("space", "enter"),
			key.WithHelp("space", "start/pause"),
		),
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart"),
		),
		Menu: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "menu"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}

	m := model{
		width:       100,
		height:      34,
		fieldWidth:  58,
		fieldHeight: 18,
		keys:        keys,
		help:        help.New(),
		spinner:     s,
		progress:    pr,
		frameStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("63")).
			Padding(0, 1),
		scoreStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("229")),
		statusStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")),
		helpStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")),
		bannerStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("62")).
			Padding(0, 1),
		flashStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("24")),
		ballStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")),
	}

	m.help.ShowAll = false
	m.resetGame()
	m.state = stateLoading

	return m
}

func tick() tea.Cmd {
	return tea.Tick(tickRate, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func loaderTick() tea.Cmd {
	return tea.Tick(loaderTickRate, func(time.Time) tea.Msg {
		return loaderTickMsg{}
	})
}

func (m model) Init() tea.Cmd {
	cmd := m.progress.SetPercent(0)
	return tea.Batch(tick(), loaderTick(), m.spinner.Tick, cmd)
}

func (m *model) resetGame() {
	m.leftScore = 0
	m.rightScore = 0
	m.state = stateWaiting
	m.aiTick = 0
	m.animTick = 0
	m.flashText = []string{"READY?", "PRESS SPACE"}
	m.flashTimer = 0
	m.resetRound(-1)
}

func (m *model) resetRound(direction float64) {
	m.leftY = float64(max(1, (m.fieldHeight-paddleHeight)/2))
	m.rightY = m.leftY
	m.ballX = float64(m.fieldWidth / 2)
	m.ballY = float64(m.fieldHeight / 2)
	m.aiAnchorY = m.leftY
	m.leftVel = 0
	m.rightVel = 0

	if direction == 0 {
		direction = -1
	}
	if direction < 0 {
		m.velX = -baseBallSpeedX
	} else {
		m.velX = baseBallSpeedX
	}

	if (m.leftScore+m.rightScore)%2 == 0 {
		m.velY = baseBallSpeedY
	} else {
		m.velY = -baseBallSpeedY
	}
}

func (m *model) resize() {
	usableWidth := max(minFieldWidth, m.width-8)
	usableHeight := max(minFieldHeight, m.height-9)

	m.fieldWidth = min(usableWidth, 72)
	m.fieldHeight = min(usableHeight, 24)

	maxPaddleY := float64(max(1, m.fieldHeight-paddleHeight+1))
	m.leftY = minFloat(m.leftY, maxPaddleY)
	m.rightY = minFloat(m.rightY, maxPaddleY)
	m.ballX = minFloat(m.ballX, float64(m.fieldWidth))
	m.ballY = minFloat(m.ballY, float64(m.fieldHeight))

	m.progress.Width = max(10, min(22, m.fieldWidth/3))
}

func (m *model) stepBall() {
	nextX := m.ballX + m.velX
	nextY := m.ballY + m.velY

	if nextY <= 1 {
		nextY = 1
		m.velY *= -1
	} else if nextY >= float64(m.fieldHeight) {
		nextY = float64(m.fieldHeight)
		m.velY *= -1
	}

	leftPaddleX := 2.0
	rightPaddleX := float64(m.fieldWidth - 1)

	if m.velX < 0 && nextX <= leftPaddleX {
		if paddleHit(nextY, m.leftY) {
			nextX = leftPaddleX
			m.velX = nextBounceSpeed(math.Abs(m.velX), m.leftVel, 0.35)
			m.velY = spinFromPaddle(nextY, m.leftY) + m.leftVel*0.08
			m.updateAIAnchor()
		}
	}

	if m.velX > 0 && nextX >= rightPaddleX {
		if paddleHit(nextY, m.rightY) {
			nextX = rightPaddleX
			m.velX = -nextBounceSpeed(math.Abs(m.velX), m.rightVel, 1.0)
			m.velY = spinFromPaddle(nextY, m.rightY) + m.rightVel*0.10
		}
	}

	m.velY = clampFloat(m.velY, -0.34, 0.34)
	m.ballX = nextX
	m.ballY = nextY

	if m.ballX < 1 {
		m.rightScore++
		if m.rightScore >= targetScore {
			m.state = stateGameOver
			m.flashText = []string{"YOU WIN THE MATCH", "crowd goes wild"}
			m.flashTimer = 9999
			return
		}
		m.state = stateWaiting
		m.flashText = []string{"GOAL FOR YOU", "nice return"}
		m.flashTimer = 96
		m.resetRound(1)
		return
	}

	if m.ballX > float64(m.fieldWidth) {
		m.leftScore++
		if m.leftScore >= targetScore {
			m.state = stateGameOver
			m.flashText = []string{"AI TAKES THE MATCH", "terminal integrity critical"}
			m.flashTimer = 9999
			return
		}
		m.state = stateWaiting
		m.flashText = []string{"AI SCORES", "shake it off"}
		m.flashTimer = 96
		m.resetRound(-1)
	}
}

func (m *model) stepAI() {
	cfg := m.currentDifficulty()
	if cfg.reactionFrames > 1 && m.aiTick%cfg.reactionFrames != 0 {
		m.leftVel = 0
		return
	}

	centerY := float64(max(1, (m.fieldHeight-paddleHeight)/2))
	targetY := m.aiAnchorY
	if m.velX < 0 && m.ballX <= float64(m.fieldWidth)*cfg.trackFromX {
		targetY = m.ballY - float64(paddleHeight)/2 + m.aiErrorOffset()
	} else if m.velX > 0 && cfg.returnBlend > 0 {
		targetY = m.aiAnchorY
	} else if m.velX > 0 {
		m.leftVel = 0
		return
	} else if cfg.repositionWindow > 0 {
		targetY = m.leftY + (centerY-m.leftY)*cfg.repositionWindow
	}

	targetY = clampFloat(targetY, 1, float64(m.fieldHeight-paddleHeight+1))
	diff := targetY - m.leftY
	if math.Abs(diff) < cfg.deadZone {
		m.leftVel = 0
		return
	}

	step := minFloat(cfg.moveStep, math.Abs(diff))
	if diff < 0 {
		m.leftY = maxFloat(1, m.leftY-step)
		m.leftVel = -step
		return
	}

	m.leftY = minFloat(float64(m.fieldHeight-paddleHeight+1), m.leftY+step)
	m.leftVel = step
}

func paddleHit(ballY, paddleY float64) bool {
	return ballY >= paddleY && ballY <= paddleY+float64(paddleHeight-1)
}

func spinFromPaddle(ballY, paddleY float64) float64 {
	center := paddleY + float64(paddleHeight-1)/2
	offset := (ballY - center) / (float64(paddleHeight) / 2)
	offset = clampFloat(offset, -1, 1)
	return offset * 0.22
}

func nextBounceSpeed(current, paddleVelocity, multiplier float64) float64 {
	boost := clampFloat(math.Abs(paddleVelocity)/paddleStep, 0, 1) * hitBoostMax * multiplier
	return clampFloat(current+0.015+boost, baseBallSpeedX, maxBallSpeedX)
}

func (m *model) updateAIAnchor() {
	cfg := m.currentDifficulty()
	centerY := float64(max(1, (m.fieldHeight-paddleHeight)/2))
	m.aiAnchorY = m.leftY + (centerY-m.leftY)*cfg.returnBlend
}

func (m model) aiErrorOffset() float64 {
	cfg := m.currentDifficulty()
	phase := m.ballX*0.37 + m.ballY*0.61 + float64(m.leftScore+m.rightScore)*0.83
	return math.Sin(phase) * cfg.aimError
}

func (m model) currentDifficulty() aiDifficulty {
	return aiDifficulties[m.menuIndex]
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resize()
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		var cmd tea.Cmd
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case loaderTickMsg:
		if m.state != stateLoading {
			return m, nil
		}

		m.loadingPct += 0.08
		if m.loadingPct >= 1 {
			m.loadingPct = 1
			m.state = stateMenu
		}

		cmd := m.progress.SetPercent(m.loadingPct)
		if m.state == stateLoading {
			return m, tea.Batch(loaderTick(), cmd)
		}
		return m, cmd

	case tickMsg:
		m.animTick++
		if m.flashTimer > 0 {
			m.flashTimer--
		}
		switch m.state {
		case stateRunning:
			m.aiTick++
			m.stepAI()
			m.stepBall()
			m.leftVel *= 0.55
			m.rightVel *= 0.55
			if math.Abs(m.leftVel) < 0.02 {
				m.leftVel = 0
			}
			if math.Abs(m.rightVel) < 0.02 {
				m.rightVel = 0
			}
		}
		return m, tick()

	case tea.KeyMsg:
		switch {
		case m.state == stateMenu && isUpKey(msg):
			m.menuIndex = (m.menuIndex + len(aiDifficulties) - 1) % len(aiDifficulties)
			return m, nil

		case m.state == stateMenu && isDownKey(msg):
			m.menuIndex = (m.menuIndex + 1) % len(aiDifficulties)
			return m, nil

		case m.state == stateMenu && isSelectKey(msg):
			m.resetGame()
			return m, nil

		case isSpaceKey(msg):
			switch m.state {
			case stateWaiting:
				m.state = stateRunning
			case stateRunning:
				m.state = statePaused
			case statePaused:
				m.state = stateRunning
			}
			return m, nil

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showAllHelp = !m.showAllHelp
			m.help.ShowAll = m.showAllHelp
			return m, nil

		case key.Matches(msg, m.keys.Restart):
			m.resetGame()
			return m, nil

		case key.Matches(msg, m.keys.Menu):
			m.state = stateMenu
			m.leftScore = 0
			m.rightScore = 0
			m.flashText = nil
			m.flashTimer = 0
			m.resetRound(-1)
			return m, nil

		case key.Matches(msg, m.keys.Start):
			if m.state == stateWaiting {
				m.state = stateRunning
			}
			return m, nil

		case key.Matches(msg, m.keys.Pause):
			switch m.state {
			case stateRunning:
				m.state = statePaused
			case statePaused:
				m.state = stateRunning
			}
			return m, nil

		case key.Matches(msg, m.keys.RightUp):
			m.rightY = maxFloat(1, m.rightY-paddleStep)
			m.rightVel = -paddleStep
			return m, nil

		case key.Matches(msg, m.keys.RightDown):
			m.rightY = minFloat(float64(m.fieldHeight-paddleHeight+1), m.rightY+paddleStep)
			m.rightVel = paddleStep
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.width < minFieldWidth || m.height < minFieldHeight {
		return "Window too small for Pong terminal"
	}

	if m.state == stateLoading {
		return m.renderLoadingView()
	}

	if m.state == stateMenu {
		return m.renderMenuView()
	}

	header := m.renderHeader()
	field := m.renderField()
	helpView := m.helpStyle.Render(m.help.View(m.keys))

	body := lipgloss.JoinVertical(lipgloss.Left, header, field, "", helpView)

	frame := m.frameStyle
	if m.width < 72 {
		frame = frame.Padding(1, 1)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, frame.Render(body))
}

func (m model) renderLoadingView() string {
	title := m.titleStyle.Render("PONG TERMINAL")
	status := m.statusStyle.Render("loading arena...")
	bar := m.progress.View()
	pct := m.helpStyle.Render(fmt.Sprintf("%.0f%%", m.loadingPct*100))

	body := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		status,
		bar,
		pct,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.frameStyle.Render(body))
}

func (m model) renderMenuView() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Render(strings.Join([]string{
			"██████╗  ██████╗ ███╗   ██╗ ██████╗ ",
			"██╔══██╗██╔═══██╗████╗  ██║██╔════╝ ",
			"██████╔╝██║   ██║██╔██╗ ██║██║  ███╗",
			"██╔═══╝ ██║   ██║██║╚██╗██║██║   ██║",
			"██║     ╚██████╔╝██║ ╚████║╚██████╔╝",
			"╚═╝      ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ",
		}, "\n"))
	subtitle := m.statusStyle.Render("choose AI difficulty")
	creator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Render(terminalLink("https://github.com/It-Shu", "github.com/It-Shu"))

	options := make([]string, 0, len(aiDifficulties))
	menuWidth := 12
	for i, difficulty := range aiDifficulties {
		line := lipgloss.NewStyle().Width(menuWidth).Align(lipgloss.Left).Render(difficulty.label)
		if i == m.menuIndex {
			line = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("230")).
				Background(lipgloss.Color("62")).
				Padding(0, 1).
				Width(menuWidth + 4).
				Render("▶ " + line)
		} else {
			line = m.helpStyle.Width(menuWidth + 4).Render("  " + line)
		}
		options = append(options, line)
	}

	hint := m.helpStyle.Render("↑/↓ choose  •  space/enter play  •  q quit")
	body := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		creator,
		"",
		subtitle,
		"",
		lipgloss.JoinVertical(lipgloss.Left, options...),
		"",
		hint,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.frameStyle.Render(body))
}

func (m model) renderBanner() string {
	switch m.state {
	case stateWaiting:
		return m.bannerStyle.Render("PRESS SPACE")
	default:
		return ""
	}
}

func (m model) renderHeader() string {
	title := m.titleStyle.Render("PONG TERMINAL")
	score := m.scoreStyle.Render(fmt.Sprintf("%s  AI %d : %d YOU", strings.ToUpper(m.currentDifficulty().name), m.leftScore, m.rightScore))
	status := m.statusStyle.Render(m.statusLine())
	banner := m.renderBanner()

	headerWidth := m.fieldWidth + 2
	gap := max(1, headerWidth-lipgloss.Width(title)-lipgloss.Width(score))
	topRow := lipgloss.JoinHorizontal(lipgloss.Top, title, strings.Repeat(" ", gap), score)
	lines := []string{topRow, status, banner}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m model) statusLine() string {
	switch m.state {
	case stateRunning:
		return "match running"
	case statePaused:
		return ""
	case stateWaiting:
		return ""
	case stateGameOver:
		if m.leftScore > m.rightScore {
			return "you lose"
		}
		return "you win"
	default:
		return ""
	}
}

func (m model) renderField() string {
	subW := m.fieldWidth * subPixelX
	subH := m.fieldHeight * subPixelY
	subGrid := make([][]bool, subH)
	for y := range subGrid {
		subGrid[y] = make([]bool, subW)
	}

	centerX := (m.fieldWidth / 2) * subPixelX
	for y := 0; y < subH; y++ {
		if y%2 == 0 {
			setSubPixel(subGrid, centerX, y)
		}
	}

	rows := make([]string, 0, m.fieldHeight)
	ballCellX := clampInt(int(math.Round(m.ballX))-1, 0, m.fieldWidth-1)
	for cellY := 0; cellY < m.fieldHeight; cellY++ {
		var row strings.Builder
		for cellX := 0; cellX < m.fieldWidth; cellX++ {
			switch cellX {
			case 1:
				row.WriteRune(paddleRuneForRow(cellY, m.leftY))
			case m.fieldWidth - 2:
				row.WriteRune(paddleRuneForRow(cellY, m.rightY))
			default:
				if cellX == ballCellX {
					if ballGlyph, ok := m.ballGlyphForRow(cellY, m.ballY); ok {
						row.WriteString(ballGlyph)
						continue
					}
				}
				row.WriteRune(brailleAt(subGrid, cellX*subPixelX, cellY*subPixelY))
			}
		}
		rows = append(rows, row.String())
	}

	switch m.state {
	case stateGameOver:
		if m.rightScore > m.leftScore {
			rows = applyConfetti(rows, m.animTick, m.fieldWidth)
		} else {
			rows = applyCracks(rows, m.animTick, m.fieldWidth)
		}
	}
	rows = overlayRows(rows, m.overlayLines(), m.fieldWidth, m.overlayStyle())

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		Render(strings.Join(rows, "\n"))
}

func (m model) overlayLines() []string {
	if m.state == stateWaiting && m.flashTimer > 0 && len(m.flashText) > 0 {
		return m.flashText
	}

	switch m.state {
	case statePaused:
		return []string{"PAUSED", "SPACE TO RESUME", "M FOR MENU"}
	case stateGameOver:
		if m.leftScore > m.rightScore {
			return []string{"YOU LOSE", "R TO RESTART", "M FOR MENU"}
		}
		return []string{"YOU WIN", "R TO RESTART", "M FOR MENU"}
	default:
		return nil
	}
}

func (m model) overlayStyle() lipgloss.Style {
	if (m.state == stateWaiting && m.flashTimer > 0 && len(m.flashText) > 0) || m.state == statePaused {
		return m.flashStyle
	}
	return m.bannerStyle
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isSpaceKey(msg tea.KeyMsg) bool {
	return msg.String() == " " || msg.String() == "space"
}

func isSelectKey(msg tea.KeyMsg) bool {
	return isSpaceKey(msg) || msg.String() == "enter"
}

func isUpKey(msg tea.KeyMsg) bool {
	return msg.String() == "up" || msg.String() == "k"
}

func isDownKey(msg tea.KeyMsg) bool {
	return msg.String() == "down" || msg.String() == "j"
}

func terminalLink(url, label string) string {
	return "\x1b]8;;" + url + "\x1b\\" + label + "\x1b]8;;\x1b\\"
}

func toSubPixelX(v float64) int {
	return max(0, int(math.Round((v-1)*subPixelX)))
}

func toSubPixelY(v float64) int {
	return max(0, int(math.Round((v-1)*subPixelY)))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func clampInt(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func setSubPixel(grid [][]bool, x, y int) {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[y]) {
		return
	}
	grid[y][x] = true
}

func brailleAt(grid [][]bool, startX, startY int) rune {
	mask := 0
	if getSubPixel(grid, startX, startY) {
		mask |= 1 << 0
	}
	if getSubPixel(grid, startX, startY+1) {
		mask |= 1 << 1
	}
	if getSubPixel(grid, startX, startY+2) {
		mask |= 1 << 2
	}
	if getSubPixel(grid, startX+1, startY) {
		mask |= 1 << 3
	}
	if getSubPixel(grid, startX+1, startY+1) {
		mask |= 1 << 4
	}
	if getSubPixel(grid, startX+1, startY+2) {
		mask |= 1 << 5
	}
	if getSubPixel(grid, startX, startY+3) {
		mask |= 1 << 6
	}
	if getSubPixel(grid, startX+1, startY+3) {
		mask |= 1 << 7
	}
	if mask == 0 {
		return ' '
	}
	return rune(0x2800 + mask)
}

func getSubPixel(grid [][]bool, x, y int) bool {
	if y < 0 || y >= len(grid) || x < 0 || x >= len(grid[y]) {
		return false
	}
	return grid[y][x]
}

func paddleRuneForRow(cellY int, paddleY float64) rune {
	rowTop := float64(cellY)
	rowBottom := rowTop + 1
	paddleTop := paddleY - 1
	paddleBottom := paddleTop + float64(paddleHeight)
	overlap := math.Min(rowBottom, paddleBottom) - math.Max(rowTop, paddleTop)

	switch {
	case overlap <= 0:
		return ' '
	case overlap >= 0.75:
		return '█'
	default:
		mid := rowTop + 0.5
		if paddleTop < mid && paddleBottom <= rowBottom {
			return '▀'
		}
		if paddleTop >= rowTop && paddleBottom > mid {
			return '▄'
		}
		return '█'
	}
}

func (m model) ballGlyphForRow(cellY int, ballY float64) (string, bool) {
	ballRow := clampInt(int(math.Round(ballY))-1, 0, m.fieldHeight-1)
	if cellY != ballRow {
		return "", false
	}

	return "●", true
}

func clampFloat(v, low, high float64) float64 {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}

func applyConfetti(rows []string, tick, width int) []string {
	if len(rows) == 0 {
		return rows
	}
	if width == 0 {
		return rows
	}

	symbols := []rune{'*', '+', 'o', '.', 'x'}
	for y := 0; y < len(rows); y++ {
		count := 1 + (y+tick)%3
		for i := 0; i < count; i++ {
			x := (y*11 + i*17 + tick*(i+1)) % width
			r := symbols[(y+i+tick)%len(symbols)]
			rows[y] = replaceRuneAt(rows[y], x, r)
		}
	}

	return rows
}

func applyCracks(rows []string, tick, width int) []string {
	strokes := []crackStroke{
		{points: []point{{14, 4}, {15, 5}, {16, 6}, {17, 7}, {18, 8}, {19, 9}}, r: '\\'},
		{points: []point{{14, 4}, {13, 5}, {12, 6}, {11, 7}, {10, 8}}, r: '/'},
		{points: []point{{14, 4}, {15, 4}, {16, 4}, {17, 4}}, r: '_'},
		{points: []point{{17, 7}, {18, 7}, {19, 6}, {20, 5}}, r: '_'},
		{points: []point{{17, 7}, {16, 8}, {15, 9}}, r: '/'},

		{points: []point{{38, 6}, {37, 7}, {36, 8}, {35, 9}, {34, 10}, {33, 11}}, r: '/'},
		{points: []point{{38, 6}, {39, 7}, {40, 8}, {41, 9}, {42, 10}}, r: '\\'},
		{points: []point{{38, 6}, {39, 6}, {40, 6}, {41, 6}}, r: '_'},
		{points: []point{{35, 9}, {36, 10}, {37, 11}}, r: '\\'},
		{points: []point{{35, 9}, {34, 9}, {33, 9}}, r: '_'},

		{points: []point{{25, 11}, {26, 12}, {27, 13}, {28, 14}}, r: '\\'},
		{points: []point{{25, 11}, {24, 12}, {23, 13}}, r: '/'},
		{points: []point{{25, 11}, {26, 11}, {27, 11}}, r: '_'},
	}

	impactPoints := []point{{14, 4}, {38, 6}, {25, 11}}
	visible := min(len(strokes), 2+tick/6)
	for i := 0; i < visible; i++ {
		for _, p := range strokes[i].points {
			if p.y >= 0 && p.y < len(rows) {
				rows[p.y] = replaceRuneAt(rows[p.y], p.x%max(1, width), strokes[i].r)
			}
		}
	}

	impactVisible := min(len(impactPoints), 1+tick/10)
	for i := 0; i < impactVisible; i++ {
		p := impactPoints[i]
		if p.y >= 0 && p.y < len(rows) {
			rows[p.y] = replaceRuneAt(rows[p.y], p.x%max(1, width), 'x')
		}
	}

	return rows
}

type point struct {
	x int
	y int
}

type crackStroke struct {
	points []point
	r      rune
}

func replaceRuneAt(s string, idx int, r rune) string {
	runes := []rune(s)
	if idx < 0 || idx >= len(runes) {
		return s
	}
	runes[idx] = r
	return string(runes)
}

func overlayRows(rows, lines []string, width int, style lipgloss.Style) []string {
	if len(rows) == 0 || len(lines) == 0 {
		return rows
	}

	startY := max(0, len(rows)/2-len(lines)/2)
	for i, line := range lines {
		y := startY + i
		if y >= len(rows) {
			break
		}
		rowRunes := []rune(rows[y])
		textWidth := len([]rune(line))
		if textWidth == 0 || textWidth > len(rowRunes) {
			continue
		}

		startX := max(0, (width-textWidth)/2)
		if startX+textWidth > len(rowRunes) {
			startX = max(0, len(rowRunes)-textWidth)
		}

		left := string(rowRunes[:startX])
		right := string(rowRunes[startX+textWidth:])
		rows[y] = left + style.Render(line) + right
	}
	return rows
}
