# CLAUDE.md - Guidelines for ohmyaime Go Project

## Build Commands

- Build: `go build`
- Run: `go run main.go wm.go`
- Run with custom path: `go run main.go wm.go [custom_aime_folder_path]`
- Test: `go test ./...`
- Test single file: `go test -v [filename]_test.go`
- Run with debug: `go run -race main.go wm.go`

## Code Style Guidelines

- **Imports**: Standard library first, third-party separated by blank line
- **Error Handling**: Check errors immediately, use panic only for initialization failures
- **Naming**: CamelCase for exported functions, lowerCamelCase for internal
- **Comments**: Use // for line comments, add comments for exported functions
- **Constants**: Define related constants in const blocks
- **Types**: Define structured types with fields aligned for readability
- **Indentation**: Tabs for indentation, spaces for alignment
- **Line Length**: Keep lines under 100 characters
- **Function Length**: Keep functions concise and focused on a single task
- **File Organization**: Related functions should be grouped together
