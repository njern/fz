package cmd

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/njern/fz/internal/api"
	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import <board-name> <csv-file>",
	Short: "Import cards from a CSV file into a board",
	Long: `Import cards from a CSV file into a board.

The CSV file must contain a "name" column (card title) and may contain a
"notes" column (card description). If the board does not exist you will be
prompted to create it.`,
	Args: cobra.ExactArgs(2),
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(importCmd)
}

type csvCard struct {
	name  string
	notes string
}

func runImport(cmd *cobra.Command, args []string) error {
	boardName := args[0]
	csvPath := args[1]

	if err := lintImportCSV(csvPath); err != nil {
		return err
	}

	cards, err := parseImportCSV(csvPath)
	if err != nil {
		return err
	}

	if len(cards) == 0 {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "No cards found in CSV file.")
		return nil
	}

	client, slug, err := accountClient()
	if err != nil {
		return err
	}

	board, err := findBoardByName(cmd, client, slug, boardName)
	if err != nil {
		return err
	}

	if board == nil {
		confirmed, err := confirmAction(cmd, fmt.Sprintf("Board %q not found. Create it?", boardName))
		if err != nil {
			return err
		}

		if !confirmed {
			return nil
		}

		board, err = createBoard(cmd, client, slug, boardName)
		if err != nil {
			return err
		}
	}

	confirmed, err := confirmAction(cmd, fmt.Sprintf("Import %d card(s) into board %q?", len(cards), board.Name))
	if err != nil {
		return err
	}

	if !confirmed {
		return nil
	}

	var created int

	for _, c := range cards {
		if err := createImportCard(cmd, client, slug, board.ID, c); err != nil {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error creating card %q: %s\n", c.name, err)
			continue
		}

		created++
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Imported %d/%d card(s) into board %q.\n", created, len(cards), board.Name)

	return nil
}

func lintImportCSV(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("opening CSV file: %w", err)
	}

	raw := strings.ReplaceAll(string(data), "\r\n", "\n")

	var nonEmpty []string

	for _, line := range strings.Split(raw, "\n") {
		if strings.TrimSpace(line) != "" {
			nonEmpty = append(nonEmpty, line)
		}
	}

	if len(nonEmpty) == 0 {
		return nil
	}

	// Semicolon delimiter check must be done first; if true the field-level
	// checks below would not be meaningful.
	if csvLineHasSemicolonDelimiter(nonEmpty[0]) {
		return fmt.Errorf("CSV lint: line 1: semicolon used as delimiter; use a comma instead")
	}

	var errs []string

	for lineNum, line := range nonEmpty {
		fields := csvRawFields(line)

		for i, field := range fields {
			if !csvFieldIsQuoted(field) {
				if lineNum == 0 {
					errs = append(errs, fmt.Sprintf("line 1: header %d (%s) is not quoted", i+1, field))
				} else {
					errs = append(errs, fmt.Sprintf("line %d: column %d value is not quoted", lineNum+1, i+1))
				}
			}

			if lineNum == 0 {
				unquoted := strings.Trim(field, `"`)
				if unquoted != strings.ToLower(unquoted) {
					errs = append(errs, fmt.Sprintf("line 1: header %d (%s) must be lowercase", i+1, unquoted))
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("CSV lint errors:\n  %s", strings.Join(errs, "\n  "))
	}

	return nil
}

// csvLineHasSemicolonDelimiter reports whether line contains a semicolon
// outside of a quoted field.
func csvLineHasSemicolonDelimiter(line string) bool {
	inQuote := false

	for _, ch := range line {
		switch ch {
		case '"':
			inQuote = !inQuote
		case ';':
			if !inQuote {
				return true
			}
		}
	}

	return false
}

// csvRawFields splits a CSV line into raw field tokens, preserving surrounding
// quote characters so callers can check whether each field was quoted.
func csvRawFields(line string) []string {
	var fields []string
	var cur strings.Builder

	inQuote := false

	for i := 0; i < len(line); i++ {
		ch := line[i]

		switch {
		case ch == '"' && !inQuote:
			inQuote = true
			cur.WriteByte(ch)
		case ch == '"' && inQuote:
			cur.WriteByte(ch)

			if i+1 < len(line) && line[i+1] == '"' {
				// Escaped quote inside a quoted field — consume both.
				i++
				cur.WriteByte(line[i])
			} else {
				inQuote = false
			}
		case ch == ',' && !inQuote:
			fields = append(fields, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(ch)
		}
	}

	fields = append(fields, cur.String())

	return fields
}

// csvFieldIsQuoted reports whether a raw CSV field token is wrapped in double
// quotes.
func csvFieldIsQuoted(field string) bool {
	return len(field) >= 2 && field[0] == '"' && field[len(field)-1] == '"'
}

func parseImportCSV(path string) ([]csvCard, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening CSV file: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV file: %w", err)
	}

	if len(records) == 0 {
		return nil, nil
	}

	header := records[0]

	nameIdx := -1
	notesIdx := -1

	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "name":
			nameIdx = i
		case "notes":
			notesIdx = i
		}
	}

	if nameIdx == -1 {
		return nil, fmt.Errorf("CSV file must contain a \"name\" column")
	}

	var cards []csvCard

	for _, row := range records[1:] {
		if nameIdx >= len(row) {
			continue
		}

		name := strings.TrimSpace(row[nameIdx])
		if name == "" {
			continue
		}

		var notes string
		if notesIdx >= 0 && notesIdx < len(row) {
			notes = strings.TrimSpace(row[notesIdx])
		}

		cards = append(cards, csvCard{name: name, notes: notes})
	}

	return cards, nil
}

func findBoardByName(cmd *cobra.Command, client *api.Client, slug, name string) (*api.Board, error) {
	var boards []api.Board
	if err := client.GetAll(cmd.Context(), fmt.Sprintf("/%s/boards", slug), &boards); err != nil {
		return nil, err
	}

	lower := strings.ToLower(name)

	for i := range boards {
		if strings.ToLower(boards[i].Name) == lower {
			return &boards[i], nil
		}
	}

	return nil, nil
}

func createBoard(cmd *cobra.Command, client *api.Client, slug, name string) (*api.Board, error) {
	body, err := json.Marshal(boardRequest{Board: boardPayload{Name: name}})
	if err != nil {
		return nil, err
	}

	var board api.Board
	if err := client.Post(cmd.Context(), fmt.Sprintf("/%s/boards", slug), bytes.NewReader(body), &board); err != nil {
		return nil, err
	}

	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Board %q created.\n", name)

	return &board, nil
}

func createImportCard(cmd *cobra.Command, client *api.Client, slug, boardID string, c csvCard) error {
	payload := map[string]any{
		"title": c.name,
	}

	if c.notes != "" {
		payload["description"] = c.notes
	}

	body, err := json.Marshal(map[string]any{"card": payload})
	if err != nil {
		return err
	}

	return client.Post(cmd.Context(), fmt.Sprintf("/%s/boards/%s/cards", slug, boardID), bytes.NewReader(body), nil)
}
