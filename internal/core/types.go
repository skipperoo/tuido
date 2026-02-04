package core

import "time"

type EntryType string

const (
	TypeNote EntryType = "note"
	TypeTodo EntryType = "todo"
)

type Entry struct {
	ID          string     `yaml:"id"`
	CreatedAt   time.Time  `yaml:"created_at"`
	CompletedAt *time.Time `yaml:"completed_at,omitempty"` // Pointer to allow null
	Text        string     `yaml:"text"`
	Author      string     `yaml:"author"`
	Type        EntryType  `yaml:"type"`
}

type Config struct {
	Author string `yaml:"author"`
}
