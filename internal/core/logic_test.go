package core

import (
	"strings"
	"testing"
)

func TestAddEntry(t *testing.T) {
	entries := []Entry{}
	entries = AddEntry(entries, "Test Note", "User", TypeNote)

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Type != TypeNote {
		t.Errorf("Expected type note, got %s", entries[0].Type)
	}

	// Test @mention
	entries = AddEntry(entries, "Task for @Bob", "User", TypeTodo)
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
	if entries[1].Author != "Bob" {
		t.Errorf("Expected author Bob, got %s", entries[1].Author)
	}
	if strings.Contains(entries[1].Text, "@Bob") {
		t.Error("Entry text should not contain the @mention")
	}
	if entries[1].Text != "Task for" {
		t.Errorf("Expected 'Task for', got '%s'", entries[1].Text)
	}
}

func TestMarkDone(t *testing.T) {
	entries := []Entry{}
	entries = AddEntry(entries, "Task 1", "User", TypeTodo)
	id := entries[0].ID

	entries = MarkDone(entries, id)
	if entries[0].CompletedAt == nil {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestMarkUndone(t *testing.T) {
	entries := []Entry{}
	entries = AddEntry(entries, "Task 1", "User", TypeTodo)
	id := entries[0].ID
	entries = MarkDone(entries, id)
	entries = MarkUndone(entries, id)

	if entries[0].CompletedAt != nil {
		t.Error("Expected CompletedAt to be nil after undone")
	}
}

func TestRemoveEntry(t *testing.T) {
	entries := []Entry{}
	entries = AddEntry(entries, "Task 1", "User", TypeTodo)
	entries = AddEntry(entries, "Task 2", "User", TypeTodo)
	id := entries[0].ID

	entries = RemoveEntry(entries, id)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Text != "Task 2" {
		t.Error("Wrong entry removed")
	}
}

func TestRemoveEntries(t *testing.T) {
	entries := []Entry{}
	entries = AddEntry(entries, "Task 1", "User", TypeTodo)
	entries = AddEntry(entries, "Task 2", "User", TypeTodo)
	entries = AddEntry(entries, "Task 3", "User", TypeTodo)

	ids := map[string]struct{}{
		entries[0].ID: {},
		entries[2].ID: {},
	}

	entries = RemoveEntries(entries, ids)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Text != "Task 2" {
		t.Errorf("Expected 'Task 2', got '%s'", entries[0].Text)
	}
}

func TestEditEntry(t *testing.T) {
	entries := []Entry{}
	entries = AddEntry(entries, "Old Text", "User", TypeTodo)
	id := entries[0].ID

	entries = EditEntry(entries, id, "New Text")
	if entries[0].Text != "New Text" {
		t.Errorf("Expected New Text, got %s", entries[0].Text)
	}
}

func TestFilterEntries(t *testing.T) {
	entries := []Entry{}
	entries = AddEntry(entries, "Apple", "User", TypeNote)
	entries = AddEntry(entries, "Banana", "User", TypeNote)

	filtered := FilterEntries(entries, "app")
	if len(filtered) != 1 {
		t.Errorf("Expected 1 match, got %d", len(filtered))
	}
	if filtered[0].Text != "Apple" {
		t.Error("Wrong filter result")
	}
}

func TestGenerateExportMarkdown(t *testing.T) {
	entries := []Entry{}
	entries = AddEntry(entries, "Note 1", "User", TypeNote)
	entries = AddEntry(entries, "Task 1", "User", TypeTodo)

	md := GenerateExportMarkdown(entries)

	if !strings.Contains(md, "# Context") {
		t.Error("Markdown missing Context section")
	}
	if !strings.Contains(md, "# Tasks") {
		t.Error("Markdown missing Tasks section")
	}
	if !strings.Contains(md, "Note 1") {
		t.Error("Markdown missing Note 1")
	}
	if !strings.Contains(md, "- [ ]") { // Unchecked task
		t.Error("Markdown missing unchecked task indicator")
	}
	if !strings.Contains(md, "Task 1") {
		t.Error("Markdown missing Task 1")
	}

	// Test completed task
	entries = MarkDone(entries, entries[1].ID)
	md = GenerateExportMarkdown(entries)
	if !strings.Contains(md, "- [x]") { // Checked task
		t.Error("Markdown missing checked task indicator")
	}
}
