package render

import (
	"strings"
	"testing"

	"github.com/njern/fz/internal/api"
)

func TestBoardLeadingAndTrailingColumns_UseVirtualLaneColors(t *testing.T) {
	leading := boardLeadingColumns(nil, nil)
	trailing := boardTrailingColumns(nil)

	if len(leading) != 2 {
		t.Fatalf("len(leading) = %d, want 2", len(leading))
	}

	if leading[0].name != "Not Now" || leading[0].colorName != BoardLaneColorComplete {
		t.Fatalf("unexpected Not Now column: %#v", leading[0])
	}

	if leading[1].name != "Maybe?" || leading[1].colorName != BoardLaneColorMaybe {
		t.Fatalf("unexpected Maybe column: %#v", leading[1])
	}

	if len(trailing) != 1 {
		t.Fatalf("len(trailing) = %d, want 1", len(trailing))
	}

	if trailing[0].name != "Done" || trailing[0].colorName != BoardLaneColorComplete {
		t.Fatalf("unexpected Done column: %#v", trailing[0])
	}
}

func TestBoardView_ClassifiesBuiltInLaneCardsBeforeCustomColumns(t *testing.T) {
	customColumn := api.Column{ID: "doing", Name: "Doing", Color: api.ColumnColor{Name: "Blue"}}
	cards := []api.Card{
		{Number: 1, Title: "Postponed Card", Column: &customColumn, Postponed: true},
		{Number: 2, Title: "Closed Card", Column: &customColumn, Closed: true},
		{Number: 3, Title: "Active Card", Column: &customColumn},
	}

	output := BoardView("Test Board", []api.Column{customColumn}, cards, 80)

	for _, header := range []string{"Not Now (1)", "Doing (1)", "Done (1)"} {
		if !strings.Contains(output, header) {
			t.Fatalf("output missing %q:\n%s", header, output)
		}
	}
}
