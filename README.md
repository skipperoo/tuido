# TUIDO (TUI Do)

A minimal TUI application for tracking project todos and developer notes. Designed for both humans and AI Agents to maintain project context.

## Features

- **Categorized Entries:** Track Notes and Todos separately.
- **Mentions:** Assign tasks using `@name`.
- **Interactive UI:** Scrollable viewport with sticky header and input area.
- **Management:** Interactive selection modes for marking tasks as Done/Undone, Editing, or batch Removal.
- **Export:** Export your context and tasks to a clean Markdown file with `/export`.
- **Responsive:** Adapts to terminal resizing.

## Installation

### From Source

```bash
make install
```

### Using Go

```bash
go install github.com/skipperoo/tuido@latest
```

## Usage

Run the application by typing `tuido` in any project directory. It creates a local `.tuido` data file.

### Commands

- `Text`: Add a new note.
- `/todo <text>`: Add a new task.
- `/todo @name <text>`: Assign a task to someone else.
- `/done` or `/d`: Mark tasks as completed.
- `/undone`: Revert completed tasks to active.
- `/edit` or `/e`: Modify existing entries.
- `/rm`: Remove entries (supports multiselect with Space).
- `/dhist`: View history of completed tasks.
- `/author <name>`: Change your display name.
- `/export`: Generate a Markdown summary.
- `/exit`: Quit the app.

## Development

### Prerequisites

- Go 1.24+

### Workflow

- **Build:** `make`
- **Test:** `make test`

## License

MIT
