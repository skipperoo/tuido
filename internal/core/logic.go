package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// AddEntry creates a new entry and appends it to the list
func AddEntry(entries []Entry, text string, author string, entryType EntryType) []Entry {
	// Parse for @mention in text if it's a todo to override author
	finalAuthor := author
	finalText := text

	if entryType == TypeTodo && strings.Contains(text, "@") {
		parts := strings.Fields(text)
		var newParts []string
		for _, p := range parts {
			if strings.HasPrefix(p, "@") && len(p) > 1 {
				finalAuthor = p[1:] // Remove @
				// Do not add the mention to newParts, effectively removing it from the text
			} else {
				newParts = append(newParts, p)
			}
		}
		finalText = strings.Join(newParts, " ")
	}

	newEntry := Entry{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		Text:      finalText,
		Author:    finalAuthor,
		Type:      entryType,
	}
	return append(entries, newEntry)
}

// MarkDone sets the completed_at timestamp for a specific entry ID
func MarkDone(entries []Entry, id string) []Entry {
	for i, e := range entries {
		if e.ID == id {
			now := time.Now()
			entries[i].CompletedAt = &now
			return entries
		}
	}
	return entries
}

// MarkUndone removes the completed_at timestamp
func MarkUndone(entries []Entry, id string) []Entry {
	for i, e := range entries {
		if e.ID == id {
			entries[i].CompletedAt = nil
			return entries
		}
	}
	return entries
}

// RemoveEntry deletes an entry by ID
func RemoveEntry(entries []Entry, id string) []Entry {
	var newEntries []Entry
	for _, e := range entries {
		if e.ID != id {
			newEntries = append(newEntries, e)
		}
	}
	return newEntries
}

// RemoveEntries deletes multiple entries by their IDs
func RemoveEntries(entries []Entry, ids map[string]struct{}) []Entry {
	var newEntries []Entry
	for _, e := range entries {
		if _, found := ids[e.ID]; !found {
			newEntries = append(newEntries, e)
		}
	}
	return newEntries
}

// EditEntry updates the text of an entry
func EditEntry(entries []Entry, id string, newText string) []Entry {
	for i, e := range entries {
		if e.ID == id {
			entries[i].Text = newText
			return entries
		}
	}
	return entries
}

// FilterEntries returns a subset of entries based on criteria
// This is a helper for view logic, not necessarily changing state
func FilterEntries(entries []Entry, filter string) []Entry {
	if filter == "" {
		return entries
	}
	var filtered []Entry
	for _, e := range entries {
		if strings.Contains(strings.ToLower(e.Text), strings.ToLower(filter)) ||
			strings.Contains(strings.ToLower(e.Author), strings.ToLower(filter)) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// GetActiveTodos returns entries that are TODO and not completed
func GetActiveTodos(entries []Entry) []Entry {
	var active []Entry
	for _, e := range entries {
		if e.Type == TypeTodo && e.CompletedAt == nil {
			active = append(active, e)
		}
	}
	return active
}

// GetCompletedTodos returns entries that are TODO and completed
func GetCompletedTodos(entries []Entry) []Entry {
	var completed []Entry
	for _, e := range entries {
		if e.Type == TypeTodo && e.CompletedAt != nil {
			completed = append(completed, e)
		}
	}
	return completed
}

// GetActiveItems returns all items that are not completed (Notes + Open Todos)
// Or maybe just anything that isn't a "completed todo"?
// Spec says for /rm: "Switches view to Selection Mode showing All Active Items."
// Usually notes don't have a completed state, so they are always active.
func GetActiveItems(entries []Entry) []Entry {
	var active []Entry
	for _, e := range entries {
		if e.Type == TypeNote || (e.Type == TypeTodo && e.CompletedAt == nil) {
			active = append(active, e)
		}
	}
	return active
}

// GenerateExportMarkdown creates the markdown content for export
func GenerateExportMarkdown(entries []Entry) string {
	var sb strings.Builder

	// Sort entries by CreatedAt (assuming they are already appended in order, but let's be safe if we merge lists later)
	// For now, assuming input is time-ordered or we just iterate.
	// If strict sorting is needed, we should copy and sort.
	// Logic: Entries are usually appended.
	// Let's filter first.

	var notes []Entry
	var todos []Entry

	for _, e := range entries {
		if e.Type == TypeNote {
			notes = append(notes, e)
		} else if e.Type == TypeTodo {
			todos = append(todos, e)
		}
	}

	// Helper to format date in the export
	fmtDate := func(t time.Time) string {
		return t.Format("2006-01-02 15:04")
	}

	sb.WriteString("# Context\n\n")
	if len(notes) == 0 {
		sb.WriteString("_No notes._\n")
	} else {
		for _, n := range notes {
			sb.WriteString(fmt.Sprintf("- **%s** (%s): %s\n", n.Author, fmtDate(n.CreatedAt), n.Text))
		}
	}

	sb.WriteString("\n# Tasks\n\n")
	if len(todos) == 0 {
		sb.WriteString("_No tasks._\n")
	} else {
		for _, t := range todos {
			check := " "
			if t.CompletedAt != nil {
				check = "x"
			}
			sb.WriteString(fmt.Sprintf("- [%s] **%s** (%s): %s\n", check, t.Author, fmtDate(t.CreatedAt), t.Text))
		}
	}

	return sb.String()
}
