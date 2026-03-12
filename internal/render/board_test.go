package render

import "testing"

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
