package night_sky

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Количество звёзд
	starCount = 150

	// Частота обновления анимации
	animationFPS   = 30
	animationDelay = time.Second / animationFPS

	// Размеры звёзд
	minStarSize = 1
	maxStarSize = 3

	// Символы звёзд: от меньшего к большему
	tinyStarSymbol   = '·'
	smallStarSymbol  = '*'
	mediumStarSymbol = '✦'

	// Интервал между этапами мерцания
	twinkleStepDuration = 180 * time.Millisecond

	// Интервалы между мерцаниями
	minTwinkleInterval = 3 * time.Second
	maxTwinkleInterval = 10 * time.Second

	// Цвета
	starColor       = "#AFCBFF"
	brightStarColor = "#FFFFFF"
)

type star struct {
	x int
	y int

	size int

	twinkling       bool
	twinkleStage    int
	nextTwinkle     time.Time
	nextTwinkleStep time.Time
}

type model struct {
	width  int
	height int

	stars []star
}

type tickMsg time.Time

func newStar(width, height int) star {
	now := time.Now()

	return star{
		x: rand.Intn(width),
		y: rand.Intn(height),

		size: rand.Intn(maxStarSize-minStarSize+1) + minStarSize,

		nextTwinkle: now.Add(randomDuration(
			0,
			minTwinkleInterval,
		)),
	}
}

func newStarfield(width, height int) []star {
	stars := make([]star, 0, starCount)

	for range starCount {
		stars = append(stars, newStar(width, height))
	}

	return stars
}

func randomDuration(min, max time.Duration) time.Duration {
	return min + time.Duration(rand.Int63n(int64(max-min)))
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return tea.Tick(animationDelay, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if m.width > 0 && m.height > 0 {
			m.stars = newStarfield(m.width, m.height)
		}

		return m, nil

	case tickMsg:
		now := time.Now()

		m.updateStars(now)

		return m, tea.Tick(animationDelay, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	return m, nil
}

func (m *model) updateStars(now time.Time) {
	for i := range m.stars {
		s := &m.stars[i]

		if !s.twinkling && now.After(s.nextTwinkle) {
			s.twinkling = true
			s.twinkleStage = 0
			s.nextTwinkleStep = now.Add(twinkleStepDuration)
			continue
		}

		if !s.twinkling || now.Before(s.nextTwinkleStep) {
			continue
		}

		s.twinkleStage++

		maxStage := s.size*2 + 2

		if s.twinkleStage >= maxStage {
			s.twinkling = false
			s.twinkleStage = 0
			s.nextTwinkle = now.Add(randomDuration(
				minTwinkleInterval,
				maxTwinkleInterval,
			))
			continue
		}

		s.nextTwinkleStep = now.Add(twinkleStepDuration)
	}
}

func getStarSymbol(s star) rune {
	if !s.twinkling {
		return starSymbol(s.size)
	}

	stage := s.twinkleStage

	switch {
	case stage <= s.size:
		return starSymbol(s.size - stage)

	case stage == s.size+1:
		return ' '

	default:
		return starSymbol(stage - s.size - 2)
	}
}

func starSymbol(size int) rune {
	switch size {
	case 0:
		return tinyStarSymbol
	case 1:
		return smallStarSymbol
	default:
		return mediumStarSymbol
	}
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

	starStyle := lipgloss.NewStyle().Foreground(
		lipgloss.Color(starColor),
	)

	brightStyle := lipgloss.NewStyle().Foreground(
		lipgloss.Color(brightStarColor),
	)

	for _, s := range m.stars {
		symbol := getStarSymbol(s)

		if symbol == ' ' {
			continue
		}

		style := starStyle

		if s.size == maxStarSize {
			style = brightStyle
		}

		canvas[s.y][s.x] = style.Render(string(symbol))
	}

	var output strings.Builder

	for _, row := range canvas {
		output.WriteString(strings.Join(row, ""))
		output.WriteRune('\n')
	}

	return output.String()
}

func Run() error {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	_, err := p.Run()

	return err
}
