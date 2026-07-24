package seed

import "testing"

func TestFINForIndexIsStableAndFixedWidth(t *testing.T) {
	seen := make(map[string]struct{})
	for index := 0; index < 100_000; index++ {
		fin := FINForIndex(index)
		if len(fin) != 7 {
			t.Fatalf("FIN %q has length %d", fin, len(fin))
		}
		if _, exists := seen[fin]; exists {
			t.Fatalf("duplicate FIN %q", fin)
		}
		seen[fin] = struct{}{}
	}
}

func TestGeneratedRowIsInternallyConsistent(t *testing.T) {
	row := RowForIndex(42)
	if len(row) != len(ColumnNames()) {
		t.Fatalf("row has %d fields, want %d", len(row), len(ColumnNames()))
	}
	if row[0] != "T000016" {
		t.Fatalf("unexpected FIN %v", row[0])
	}
}
