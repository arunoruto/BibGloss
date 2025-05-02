// bibmanager/main.go
package main

import (
	"fmt"
	"io/fs" // Use io/fs for reading directories (more modern than ioutil)
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Model ---
// The model holds the application's state.
type model struct {
	pdfFiles []string // List of PDF filenames found
	cursor   int      // Which file the user's cursor is pointing at
	err      error    // To store any errors that occur
	loading  bool     // Flag to indicate if we are still loading files
	styles   Styles   // Lip Gloss styles
}

// --- Styles ---
// Using a struct for styles is a nice way to organize them.
type Styles struct {
	Selected lipgloss.Style
	Normal   lipgloss.Style
	Help     lipgloss.Style
	Loading  lipgloss.Style
	Error    lipgloss.Style
}

// Default styles
func DefaultStyles() Styles {
	return Styles{
		Selected: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")). // A nice purple
			Background(lipgloss.Color("236")). // A subtle background
			PaddingLeft(1).
			PaddingRight(1),
		Normal: lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1),
		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")), // Dim gray
		Loading: lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")), // Magenta-ish
		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")), // Red
	}
}

// --- Messages ---
// Bubble Tea uses messages to communicate events (like key presses,
// window size changes, or results from commands).

// A message indicating that the PDF file scan is complete.
type pdfFilesLoadedMsg struct {
	files []string
	err   error
}

// --- Commands ---
// Commands are functions that perform I/O or other operations that
// might block or take time. They return a tea.Msg when done.

func listPDFFilesCmd() tea.Msg {
	dir := "." // Current directory
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		// Handle potential errors during walk
		if err != nil {
			// We could potentially log this or decide to stop the walk
			// For now, let's just report the first error encountered.
			// Returning the error stops the walk.
			return err
		}

		// Skip directories and check for PDF extension (case-insensitive)
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".pdf") {
			// Only add the base name, not the full path (unless desired)
			// If you ran this in a subdirectory, path would include the subdir name.
			// Let's just use the base name for simplicity here.
			files = append(files, d.Name())
		}
		return nil // Continue walking
	})

	// If WalkDir itself returned an error (e.g., permission denied on root dir),
	// pass it along in the message.
	if err != nil {
		return pdfFilesLoadedMsg{err: fmt.Errorf("error walking directory '%s': %w", dir, err)}
	}

	// Return the message with the found files (or nil error if successful)
	return pdfFilesLoadedMsg{files: files, err: nil}
}

// --- Initial Model ---
// initialModel creates the starting state of our application.
func initialModel() model {
	return model{
		pdfFiles: []string{},
		cursor:   0,
		err:      nil,
		loading:  true, // Start in loading state
		styles:   DefaultStyles(),
	}
}

// --- Init ---
// Init is the first function that Bubble Tea calls. It can return a command.
// We'll use it to trigger the initial scan for PDF files.
func (m model) Init() tea.Cmd {
	// Return the command that lists PDF files.
	// Bubble Tea will run this command and send the resulting message
	// back to our Update function.
	return listPDFFilesCmd
}

// --- Update ---
// Update handles incoming messages and updates the model accordingly.
// It can also return commands.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Handle the result of our file scanning command
	case pdfFilesLoadedMsg:
		m.loading = false // Loading is finished
		if msg.err != nil {
			m.err = msg.err // Store the error
			return m, nil   // No further command
		}
		m.pdfFiles = msg.files
		// Reset cursor if list is empty or shorter than current cursor
		if len(m.pdfFiles) == 0 || m.cursor >= len(m.pdfFiles) {
			m.cursor = 0
		}
		return m, nil

	// Handle keyboard input
	case tea.KeyMsg:
		switch msg.String() {
		// Keys to quit the application
		case "ctrl+c", "q":
			return m, tea.Quit // Special command to exit Bubble Tea

		// Navigation keys
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.pdfFiles)-1 {
				m.cursor++
			}
			// TODO: Add 'enter' key later to do something with the selected file
		}

	// Handle errors (e.g., from Bubble Tea itself)
	case error:
		m.err = msg
		return m, nil
	}

	// Return the updated model and potentially a command (nil in this case)
	return m, nil
}

// --- View ---
// View renders the UI based on the current model state.
func (m model) View() string {
	var s strings.Builder // Efficiently build the output string

	if m.loading {
		s.WriteString(m.styles.Loading.Render("Scanning for PDF files..."))
		return s.String()
	}

	if m.err != nil {
		s.WriteString(m.styles.Error.Render(fmt.Sprintf("Error: %v\n\n", m.err)))
		s.WriteString(m.styles.Help.Render("Press q to quit."))
		return s.String()
	}

	s.WriteString("Found PDF files:\n\n")

	if len(m.pdfFiles) == 0 {
		s.WriteString("No PDF files found in the current directory.\n")
	} else {
		for i, file := range m.pdfFiles {
			// Check if the cursor is on this item
			if m.cursor == i {
				s.WriteString(m.styles.Selected.Render("> " + file))
			} else {
				s.WriteString(m.styles.Normal.Render("  " + file))
			}
			s.WriteString("\n") // Add newline after each file
		}
	}

	s.WriteString("\n") // Add some space before help text
	s.WriteString(m.styles.Help.Render("Use up/down keys (or k/j) to navigate. Press q to quit."))

	return s.String()
}

// --- Main Function ---
func main() {
	// Create the Bubble Tea program
	p := tea.NewProgram(initialModel())

	// Start the program
	// Use the alternate screen buffer (like vim or less) by default
	// FullScreen() enables this. If you omit it, it runs inline.
	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
