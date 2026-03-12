package render

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/njern/fz/internal/api"
	"golang.org/x/term"
)

// TerminalWidth returns the width of the terminal, defaulting to 80.
func TerminalWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		return w
	}

	return 80
}

var columnColors = map[string]lipgloss.Color{
	"Blue":   lipgloss.Color("12"),
	"Gray":   lipgloss.Color("245"),
	"Tan":    lipgloss.Color("180"),
	"Yellow": lipgloss.Color("220"),
	"Lime":   lipgloss.Color("118"),
	"Aqua":   lipgloss.Color("44"),
	"Violet": lipgloss.Color("141"),
	"Purple": lipgloss.Color("135"),
	"Pink":   lipgloss.Color("205"),
}

var (
	dimStyle    = lipgloss.NewStyle().Faint(true)
	goldenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	boardStyle  = lipgloss.NewStyle().Bold(true)
)

const (
	BoardLaneColorMaybe    = "Blue"
	BoardLaneColorComplete = "Gray"
)

type boardColumnView struct {
	name      string
	colorName string
	cards     []api.Card
}

// BoardView renders a kanban board with columns and cards.
func BoardView(boardName string, columns []api.Column, cards []api.Card, width int) string {
	var maybeCards, notNowCards, doneCards []api.Card

	byColumn := map[string][]api.Card{}

	for _, c := range cards {
		if c.Column != nil {
			byColumn[c.Column.ID] = append(byColumn[c.Column.ID], c)
		} else if c.Closed {
			doneCards = append(doneCards, c)
		} else if c.Postponed {
			notNowCards = append(notNowCards, c)
		} else {
			maybeCards = append(maybeCards, c)
		}
	}

	cols := boardLeadingColumns(maybeCards, notNowCards)

	for _, col := range columns {
		cols = append(cols, boardColumnView{
			name:      col.Name,
			colorName: col.Color.Name,
			cards:     byColumn[col.ID],
		})
	}

	cols = append(cols, boardTrailingColumns(doneCards)...)

	numCols := len(cols)
	colWidth := max((width-numCols-1)/numCols, 12)

	maxCards := 0
	for _, col := range cols {
		if len(col.cards) > maxCards {
			maxCards = len(col.cards)
		}
	}

	if maxCards == 0 {
		maxCards = 1
	}

	var sb strings.Builder

	sb.WriteString(boardStyle.Render(boardName))
	sb.WriteString("\n")

	// ┌───┬───┐
	sb.WriteString("┌")

	for i := range numCols {
		sb.WriteString(strings.Repeat("─", colWidth))

		if i < numCols-1 {
			sb.WriteString("┬")
		}
	}

	sb.WriteString("┐\n")

	// │ Header │
	sb.WriteString("│")

	for i, col := range cols {
		header := styleHeader(fmt.Sprintf(" %s (%d)", col.name, len(col.cards)), col.colorName)
		sb.WriteString(padRight(header, colWidth))

		if i < numCols-1 {
			sb.WriteString("│")
		}
	}

	sb.WriteString("│\n")

	// ├───┼───┤
	sb.WriteString("├")

	for i := range numCols {
		sb.WriteString(strings.Repeat("─", colWidth))

		if i < numCols-1 {
			sb.WriteString("┼")
		}
	}

	sb.WriteString("┤\n")

	// Card rows
	for row := range maxCards {
		sb.WriteString("│")

		for i, col := range cols {
			if row < len(col.cards) {
				sb.WriteString(formatCard(col.cards[row], colWidth))
			} else {
				sb.WriteString(strings.Repeat(" ", colWidth))
			}

			if i < numCols-1 {
				sb.WriteString("│")
			}
		}

		sb.WriteString("│\n")
	}

	// └───┴───┘
	sb.WriteString("└")

	for i := range numCols {
		sb.WriteString(strings.Repeat("─", colWidth))

		if i < numCols-1 {
			sb.WriteString("┴")
		}
	}

	sb.WriteString("┘\n")

	return sb.String()
}

func boardLeadingColumns(maybeCards, notNowCards []api.Card) []boardColumnView {
	return []boardColumnView{
		{name: "Not Now", colorName: BoardLaneColorComplete, cards: notNowCards},
		{name: "Maybe?", colorName: BoardLaneColorMaybe, cards: maybeCards},
	}
}

func boardTrailingColumns(doneCards []api.Card) []boardColumnView {
	return []boardColumnView{
		{name: "Done", colorName: BoardLaneColorComplete, cards: doneCards},
	}
}

func styleHeader(text, colorName string) string {
	if c, ok := columnColors[colorName]; ok {
		return lipgloss.NewStyle().Foreground(c).Bold(true).Render(text)
	}

	return lipgloss.NewStyle().Bold(true).Render(text)
}

func formatCard(card api.Card, width int) string {
	num := dimStyle.Render(fmt.Sprintf(" #%d", card.Number))

	suffix := ""
	if card.Golden {
		suffix = goldenStyle.Render(" ★")
	}

	numWidth := lipgloss.Width(num)
	suffixWidth := lipgloss.Width(suffix)
	titleWidth := max(width-numWidth-suffixWidth-1, 0)

	title := card.Title
	if titleRunes := []rune(title); len(titleRunes) > titleWidth {
		if titleWidth > 1 {
			title = string(titleRunes[:titleWidth-1]) + "…"
		} else {
			title = ""
		}
	}

	cell := num + " " + title + suffix

	return padRight(cell, width)
}

func padRight(s string, width int) string {
	vis := lipgloss.Width(s)
	if vis >= width {
		return s
	}

	return s + strings.Repeat(" ", width-vis)
}
