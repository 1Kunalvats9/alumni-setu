package main

import "testing"

func chainOf(actions ...string) []LedgerEntry {
	entries := []LedgerEntry{}
	previous := genesisHash

	for i, action := range actions {
		entry := LedgerEntry{
			ID:          int64(i + 1),
			Action:      action,
			SubjectID:   int64(i + 10),
			SubjectName: "Subject",
			ActorID:     1,
			ActorName:   "Admin",
			Details:     "details for " + action,
			CreatedAt:   "2026-10-25T10:0" + string(rune('0'+i)) + ":00Z",
			PrevHash:    previous,
		}
		entry.Hash = entryHash(entry)
		previous = entry.Hash
		entries = append(entries, entry)
	}
	return entries
}

func TestLedgerIntactChainPasses(t *testing.T) {
	intact, brokenAt, reason := auditLedger(chainOf("verified", "unverified", "verified"))

	if !intact {
		t.Fatalf("expected intact chain, broke at %d: %s", brokenAt, reason)
	}
}

func TestLedgerEmptyChainPasses(t *testing.T) {
	if intact, _, _ := auditLedger([]LedgerEntry{}); !intact {
		t.Error("an empty ledger should be considered intact")
	}
}

func TestLedgerDetectsEditedContent(t *testing.T) {
	entries := chainOf("verified", "unverified", "verified")
	entries[1].Details = "quietly changed after the fact"

	intact, brokenAt, _ := auditLedger(entries)
	if intact {
		t.Fatal("expected the edit to be detected")
	}
	if brokenAt != 2 {
		t.Errorf("expected the break at record 2, got %d", brokenAt)
	}
}

func TestLedgerDetectsRecomputedHash(t *testing.T) {
	entries := chainOf("verified", "unverified", "verified")
	entries[1].Details = "quietly changed after the fact"
	entries[1].Hash = entryHash(entries[1])

	intact, brokenAt, _ := auditLedger(entries)
	if intact {
		t.Fatal("recomputing one hash must still break the following link")
	}
	if brokenAt != 3 {
		t.Errorf("expected the break to surface at record 3, got %d", brokenAt)
	}
}

func TestLedgerDetectsDeletedRecord(t *testing.T) {
	entries := chainOf("verified", "unverified", "verified")
	shortened := []LedgerEntry{entries[0], entries[2]}

	if intact, _, _ := auditLedger(shortened); intact {
		t.Fatal("expected a removed record to be detected")
	}
}
