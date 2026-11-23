package space_shooter

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Game
	gameFPS   = 30
	gameDelay = time.Second / gameFPS

	// Player
	playerSymbol = '▲'
	//playerSpeed        = 35.0
	playerStartYOffset = 2
	playerMoveStep     = 2.0

	// Shooting
	bulletSymbol      = '|'
	bulletSpeed       = 55.0
	shootCooldown     = 250 * time.Millisecond
	bulletSpawnOffset = 1.0

	// Aliens
	//alienCountLimit = 30

	// Alien appearance
	alienSymbols = "WMX"

	alienCollisionRadius = 1.0
	// Background
	backgroundStarCount  = 50
	backgroundStarSymbol = '.'

	backgroundStarSpeedMin = 3.0
	backgroundStarSpeedMax = 12.0

	// Progression
	killsPerLevel = 10

	initialSpawnInterval  = 1500 * time.Millisecond
	spawnIntervalDecrease = 100 * time.Millisecond
	minimumSpawnInterval  = 350 * time.Millisecond

	initialAlienSpeed          = 2.0
	alienSpeedIncreasePerLevel = 0.5
	maximumAlienSpeed          = 10.0

	initialWaveSize              = 1
	additionalAlienEveryLevels   = 1
	initialAlienLimit            = 5
	additionalAlienLimitPerLevel = 1

	// Rendering
	hudHeight = 1

	// Storage
	recordFileName = "records.json"
	gameRecordKey  = "space_shooter"

	// Colors
	playerColor = "#5DE4C7"
	bulletColor = "#F9E27D"

	alienColor1 = "#FF6B6B"
	alienColor2 = "#C792EA"
	alienColor3 = "#FFB86C"

	backgroundColor = "#526580"
	hudColor        = "#FFFFFF"
	gameOverColor   = "#FF5555"

	// Game over
	gameOverTitle = " GAME OVER "
	restartText   = "Press R to restart"
	exitText      = "Press Q or ESC to return to menu"
)

type player struct {
	x float64
	y float64
}

type bullet struct {
	x float64
	y float64
}

type alien struct {
	x float64
	y float64

	vx float64
	vy float64

	symbol rune
	color  string
}

type backgroundStar struct {
	x     int
	y     float64
	speed float64
}

type model struct {
	width  int
	height int

	player     player
	bullets    []bullet
	aliens     []alien
	background []backgroundStar

	score  int
	level  int
	record int

	gameOver bool

	lastFrame time.Time
	lastShot  time.Time

	spawnTimer time.Duration
	nextSpawn  time.Duration

	recordPath string
}

type tickMsg time.Time

func initialModel() model {
	path, err := getRecordPath()
	if err != nil {
		path = ""
	}

	return model{
		recordPath: path,
		record:     loadRecord(path),
	}
}

func (m model) Init() tea.Cmd {
	return nextTick()
}

func nextTick() tea.Cmd {
	return tea.Tick(gameDelay, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.width > 0 && m.height > hudHeight+playerStartYOffset {
			m.reset()
		}

		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "r":
			if m.gameOver {
				m.reset()
			}

		case "left", "a":
			if !m.gameOver {
				m.player.x -= playerMoveStep
				m.clampPlayer()
			}

		case "right", "d":
			if !m.gameOver {
				m.player.x += playerMoveStep
				m.clampPlayer()
			}

		case " ", "space":
			if !m.gameOver {
				m.shoot()
			}
		}

		return m, nil

	case tickMsg:
		now := time.Now()

		if !m.gameOver {
			m.update(now)
		}

		m.lastFrame = now

		return m, nextTick()
	}

	return m, nil
}

func (m *model) reset() {
	m.player = player{
		x: float64(m.width) / 2,
		y: float64(m.height - playerStartYOffset),
	}

	m.bullets = nil
	m.aliens = nil

	m.score = 0
	m.level = 1
	m.gameOver = false

	m.lastFrame = time.Now()
	m.lastShot = time.Time{}

	m.spawnTimer = 0
	m.nextSpawn = initialSpawnInterval

	m.generateBackground()
}

func (m *model) generateBackground() {
	m.background = make([]backgroundStar, 0, backgroundStarCount)

	for range backgroundStarCount {
		m.background = append(m.background, backgroundStar{
			x: rand.Intn(m.width),
			y: rand.Float64() * float64(m.height),
			speed: backgroundStarSpeedMin +
				rand.Float64()*(backgroundStarSpeedMax-backgroundStarSpeedMin),
		})
	}
}

func (m *model) clampPlayer() {
	minX := 1.0
	maxX := float64(m.width - 2)

	if m.player.x < minX {
		m.player.x = minX
	}

	if m.player.x > maxX {
		m.player.x = maxX
	}
}

func (m *model) shoot() {
	now := time.Now()

	if now.Sub(m.lastShot) < shootCooldown {
		return
	}

	m.lastShot = now

	m.bullets = append(m.bullets, bullet{
		x: m.player.x,
		y: m.player.y - bulletSpawnOffset,
	})
}

func (m *model) update(now time.Time) {
	deltaTime := now.Sub(m.lastFrame).Seconds()

	// Ограничиваем скачок времени после подвисания.
	if deltaTime > 0.1 {
		deltaTime = 0.1
	}

	m.updateBackground(deltaTime)
	m.updateBullets(deltaTime)
	m.updateAliens(deltaTime)
	m.updateSpawning(time.Duration(deltaTime * float64(time.Second)))
	m.checkCollisions()
	m.checkGameOver()
}

func (m *model) updateBackground(deltaTime float64) {
	for i := range m.background {
		s := &m.background[i]

		s.y += s.speed * deltaTime

		if s.y >= float64(m.height) {
			s.y = 0
			s.x = rand.Intn(m.width)
		}
	}
}

func (m *model) updateBullets(deltaTime float64) {
	for i := range m.bullets {
		m.bullets[i].y -= bulletSpeed * deltaTime
	}

	remaining := m.bullets[:0]

	for _, b := range m.bullets {
		if b.y >= hudHeight {
			remaining = append(remaining, b)
		}
	}

	m.bullets = remaining
}

func (m *model) updateAliens(deltaTime float64) {
	for i := range m.aliens {
		a := &m.aliens[i]

		a.x += a.vx * deltaTime
		a.y += a.vy * deltaTime
	}
}

func (m *model) updateSpawning(deltaTime time.Duration) {
	m.spawnTimer += deltaTime

	if m.spawnTimer < m.nextSpawn {
		return
	}

	m.spawnTimer = 0
	m.nextSpawn = m.getSpawnInterval()

	alienLimit := initialAlienLimit +
		(m.level-1)*additionalAlienLimitPerLevel

	if len(m.aliens) >= alienLimit {
		return
	}

	waveSize := initialWaveSize +
		(m.level-1)/additionalAlienEveryLevels

	for range waveSize {
		if len(m.aliens) >= alienLimit {
			break
		}

		m.spawnAlien()
	}
}

func (m model) getSpawnInterval() time.Duration {
	interval := initialSpawnInterval -
		time.Duration(m.level-1)*spawnIntervalDecrease

	if interval < minimumSpawnInterval {
		interval = minimumSpawnInterval
	}

	return interval
}

func (m *model) spawnAlien() {
	if m.width < 5 || m.height < 5 {
		return
	}

	a := alien{
		x: float64(rand.Intn(m.width-2)) + 1,
		y: float64(hudHeight),
	}

	// Пришелец направляется к кораблю.
	dx := m.player.x - a.x
	dy := m.player.y - a.y

	distance := math.Sqrt(dx*dx + dy*dy)

	if distance == 0 {
		distance = 1
	}

	speed := initialAlienSpeed +
		float64(m.level-1)*alienSpeedIncreasePerLevel

	if speed > maximumAlienSpeed {
		speed = maximumAlienSpeed
	}

	a.vx = dx / distance * speed
	a.vy = dy / distance * speed

	alienType := rand.Intn(len(alienSymbols))

	a.symbol = rune(alienSymbols[alienType])

	switch alienType {
	case 0:
		a.color = alienColor1
	case 1:
		a.color = alienColor2
	default:
		a.color = alienColor3
	}

	m.aliens = append(m.aliens, a)
}

func (m *model) checkCollisions() {
	aliveAliens := m.aliens[:0]

	for _, a := range m.aliens {
		hit := false

		for i := range m.bullets {
			b := &m.bullets[i]

			if b.y < hudHeight {
				continue
			}

			if math.Abs(a.x-b.x) <= alienCollisionRadius &&
				math.Abs(a.y-b.y) <= alienCollisionRadius {
				b.y = -1
				hit = true

				m.score++
				m.level = m.score/killsPerLevel + 1

				if m.score > m.record {
					m.record = m.score
				}

				break
			}
		}

		if !hit {
			aliveAliens = append(aliveAliens, a)
		}
	}

	m.aliens = aliveAliens

	aliveBullets := m.bullets[:0]

	for _, b := range m.bullets {
		if b.y >= hudHeight {
			aliveBullets = append(aliveBullets, b)
		}
	}

	m.bullets = aliveBullets
}

func (m *model) checkGameOver() {
	for _, a := range m.aliens {
		if a.y >= m.player.y-1 {
			m.gameOver = true

			if m.recordPath != "" {
				_ = saveRecord(m.recordPath, m.score)
			}

			return
		}
	}
}

func getRecordPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "termfun", recordFileName), nil
}

func loadRecord(path string) int {
	if path == "" {
		return 0
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}

	records := make(map[string]int)

	if err := json.Unmarshal(data, &records); err != nil {
		return 0
	}

	return records[gameRecordKey]
}

func saveRecord(path string, score int) error {
	if path == "" {
		return nil
	}

	records := make(map[string]int)

	data, err := os.ReadFile(path)
	if err == nil {
		_ = json.Unmarshal(data, &records)
	}

	if score <= records[gameRecordKey] {
		return nil
	}

	records[gameRecordKey] = score

	data, err = json.MarshalIndent(records, "", "    ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	canvas := make([][]string, m.height)

	for y := range canvas {
		canvas[y] = make([]string, m.width)

		for x := range canvas[y] {
			canvas[y][x] = " "
		}
	}

	m.drawBackground(canvas)
	m.drawBullets(canvas)
	m.drawAliens(canvas)
	m.drawPlayer(canvas)

	var output strings.Builder

	output.WriteString(m.drawHUD())
	output.WriteRune('\n')

	for y := hudHeight; y < m.height; y++ {
		output.WriteString(strings.Join(canvas[y], ""))
		output.WriteRune('\n')
	}

	if m.gameOver {
		output.WriteString(m.drawGameOver())
	}

	return output.String()
}

func (m model) drawHUD() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(hudColor)).
		Bold(true)

	return style.Render(fmt.Sprintf(
		" SCORE: %d  |  RECORD: %d  |  LEVEL: %d ",
		m.score,
		m.record,
		m.level,
	))
}

func (m model) drawBackground(canvas [][]string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(backgroundColor))

	for _, s := range m.background {
		x := s.x
		y := int(math.Round(s.y))

		if x >= 0 && x < m.width &&
			y >= hudHeight && y < m.height {
			canvas[y][x] = style.Render(string(backgroundStarSymbol))
		}
	}
}

func (m model) drawBullets(canvas [][]string) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(bulletColor)).
		Bold(true)

	for _, b := range m.bullets {
		x := int(math.Round(b.x))
		y := int(math.Round(b.y))

		if x >= 0 && x < m.width &&
			y >= hudHeight && y < m.height {
			canvas[y][x] = style.Render(string(bulletSymbol))
		}
	}
}

func (m model) drawAliens(canvas [][]string) {
	for _, a := range m.aliens {
		x := int(math.Round(a.x))
		y := int(math.Round(a.y))

		if x >= 0 && x < m.width &&
			y >= hudHeight && y < m.height {
			style := lipgloss.NewStyle().
				Foreground(lipgloss.Color(a.color)).
				Bold(true)

			canvas[y][x] = style.Render(string(a.symbol))
		}
	}
}

func (m model) drawPlayer(canvas [][]string) {
	x := int(math.Round(m.player.x))
	y := int(math.Round(m.player.y))

	if x >= 0 && x < m.width &&
		y >= hudHeight && y < m.height {
		style := lipgloss.NewStyle().
			Foreground(lipgloss.Color(playerColor)).
			Bold(true)

		canvas[y][x] = style.Render(string(playerSymbol))
	}
}

func (m model) drawGameOver() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(gameOverColor)).
		Bold(true)

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(hudColor))

	return "\n" +
		titleStyle.Render(gameOverTitle) + "\n" +
		infoStyle.Render(fmt.Sprintf("Score: %d", m.score)) + "\n" +
		infoStyle.Render(fmt.Sprintf("Record: %d", m.record)) + "\n" +
		infoStyle.Render(restartText) + "\n" +
		infoStyle.Render(exitText)
}

func Run() error {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	_, err := p.Run()

	return err
}
