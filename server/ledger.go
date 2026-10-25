package main

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
)

const genesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

func entryHash(e LedgerEntry) string {
	parts := []string{
		strconv.FormatInt(e.ID, 10),
		e.Action,
		strconv.FormatInt(e.SubjectID, 10),
		e.SubjectName,
		strconv.FormatInt(e.ActorID, 10),
		e.ActorName,
		e.Details,
		e.CreatedAt,
		e.PrevHash,
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func lastLedgerHash() string {
	var hash string
	err := db.QueryRow(`SELECT hash FROM ledger ORDER BY id DESC LIMIT 1`).Scan(&hash)
	if err != nil {
		return genesisHash
	}
	return hash
}

func appendLedger(action string, subject *User, actor *User, details string) {
	entry := LedgerEntry{
		Action:      action,
		SubjectID:   subject.ID,
		SubjectName: subject.Name,
		ActorID:     actor.ID,
		ActorName:   actor.Name,
		Details:     details,
		CreatedAt:   now(),
		PrevHash:    lastLedgerHash(),
	}

	result, err := db.Exec(`INSERT INTO ledger
		(action, subject_id, subject_name, actor_id, actor_name, details, created_at, prev_hash, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, '')`,
		entry.Action, entry.SubjectID, entry.SubjectName, entry.ActorID,
		entry.ActorName, entry.Details, entry.CreatedAt, entry.PrevHash)
	if err != nil {
		return
	}

	entry.ID, _ = result.LastInsertId()
	db.Exec(`UPDATE ledger SET hash = ? WHERE id = ?`, entryHash(entry), entry.ID)
}

func readLedger() ([]LedgerEntry, error) {
	rows, err := db.Query(`SELECT id, action, subject_id, subject_name, actor_id,
		actor_name, details, created_at, prev_hash, hash FROM ledger ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []LedgerEntry{}
	for rows.Next() {
		var e LedgerEntry
		if err := rows.Scan(&e.ID, &e.Action, &e.SubjectID, &e.SubjectName,
			&e.ActorID, &e.ActorName, &e.Details, &e.CreatedAt,
			&e.PrevHash, &e.Hash); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func auditLedger(entries []LedgerEntry) (bool, int64, string) {
	previous := genesisHash

	for _, entry := range entries {
		if entry.PrevHash != previous {
			return false, entry.ID, "the link to the previous record does not match"
		}
		if entryHash(entry) != entry.Hash {
			return false, entry.ID, "this record's contents do not match its fingerprint"
		}
		previous = entry.Hash
	}

	return true, 0, ""
}

func handleLedger(w http.ResponseWriter, r *http.Request, _ *User) {
	entries, err := readLedger()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not read the ledger")
		return
	}

	intact, brokenAt, reason := auditLedger(entries)

	writeJSON(w, http.StatusOK, map[string]any{
		"count":     len(entries),
		"intact":    intact,
		"broken_at": brokenAt,
		"reason":    reason,
		"results":   entries,
	})
}
