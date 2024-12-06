package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	xstrings "github.com/charmbracelet/x/exp/strings"
)

// DevSetup holds all configuration for the development environment
type DevSetup struct {
	Language   Language
	CreateRepo bool
}

// Language contains all language-specific settings and state
type Language struct {
	Type       string   // Selected programming language
	Version    string   // Version to install
	Editors    []string // Selected development editors
	CurrentVer string   // Currently installed version
	Path       string   // Installation path
}

var (
	highlight = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212")).
			Bold(true)

	subtle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("161")).
			Bold(true)
)

// getLanguageVersion checks if a language is installed and returns version info
func getLanguageVersion(language string) (version, path string, err error) {
	var cmd *exec.Cmd
	var out bytes.Buffer
	var stderr bytes.Buffer

	language = strings.Split(language, " ")[0] // Remove emoji

	switch language {
	case "Python":
		cmd = exec.Command("python3", "--version")
	case "JavaScript":
		cmd = exec.Command("node", "--version")
	case "Go":
		cmd = exec.Command("go", "version")
	case "Rust":
		cmd = exec.Command("rustc", "--version")
	case "Java":
		cmd = exec.Command("java", "--version")
	default:
		return "", "", fmt.Errorf("unsupported language: %s", language)
	}

	path, err = exec.LookPath(cmd.Path)
	if err != nil {
		return "", "", fmt.Errorf("not found in PATH")
	}

	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", "", err
	}

	// Process version string based on language format
	version = strings.TrimSpace(out.String())
	switch language {
	case "Python":
		parts := strings.Split(version, " ")
		if len(parts) >= 2 {
			version = parts[1]
		}
	case "JavaScript":
		version = strings.TrimPrefix(version, "v")
	}

	return version, path, nil
}

// getAvailableVersions returns available versions for a given language
func getAvailableVersions(language string) []string {
	versions := map[string][]string{
		"Python": {
			"3.12.1 (Latest)",
			"3.11.7 (LTS)",
			"3.10.13",
			"3.9.18",
		},
		"JavaScript": {
			"20.11.0 (Latest)",
			"18.19.0 (LTS)",
			"16.20.2",
			"14.21.3",
		},
		"Go": {
			"1.22.0 (Latest)",
			"1.21.6 (LTS)",
			"1.20.12",
			"1.19.13",
		},
		"Rust": {
			"1.75.0 (Latest)",
			"1.74.1",
			"1.73.0",
		},
		"Java": {
			"21.0.2 (Latest)",
			"17.0.10 (LTS)",
			"11.0.22",
		},
	}

	language = strings.Split(language, " ")[0]
	return versions[language]
}

// getLanguageEditors returns recommended editors for a given language
func getLanguageEditors(language string) []huh.Option[string] {
	// Base editors available for all languages
	baseEditors := []huh.Option[string]{
		huh.NewOption("VS Code 💻", "VS Code").Selected(true),
		huh.NewOption("Neovim 🔮", "Neovim"),
		huh.NewOption("Sublime Text ✨", "Sublime Text"),
	}

	// Add language-specific editors
	language = strings.Split(language, " ")[0]
	switch language {
	case "Python":
		return append(baseEditors,
			huh.NewOption("PyCharm 🐍", "PyCharm"),
		)
	case "Go":
		return append(baseEditors,
			huh.NewOption("GoLand 🎯", "GoLand"),
		)
	case "JavaScript":
		return append(baseEditors,
			huh.NewOption("WebStorm 🌐", "WebStorm"),
		)
	case "Java":
		return append(baseEditors,
			huh.NewOption("IntelliJ IDEA ☕", "IntelliJ IDEA"),
		)
	default:
		return baseEditors
	}
}

func main() {
	var setup DevSetup
	accessible, _ := strconv.ParseBool(os.Getenv("ACCESSIBLE"))

	// First form: Language selection and version checking
	languageForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Options(huh.NewOptions(
					"Python 🐍",
					"JavaScript 💫",
					"Go 🚀",
					"Rust 🦀",
					"Java ☕",
				)...).
				Title("Choose a programming language").
				Value(&setup.Language.Type),

			// Show current version if installed
			huh.NewNote().
				TitleFunc(func() string {
					ver, path, err := getLanguageVersion(setup.Language.Type)
					if err == nil {
						setup.Language.CurrentVer = ver
						setup.Language.Path = path
						return highlight.Render(fmt.Sprintf("Found %s installation:", setup.Language.Type))
					}
					return subtle.Render("No existing installation found")
				}, &setup.Language.Type).
				DescriptionFunc(func() string {
					if setup.Language.CurrentVer != "" {
						return fmt.Sprintf("Version: %s\nPath: %s",
							highlight.Render(setup.Language.CurrentVer),
							subtle.Render(setup.Language.Path))
					}
					return subtle.Render("You can proceed with a fresh installation")
				}, &setup.Language.Type),

			// Version selection
			huh.NewSelect[string]().
				Title("Choose version to install").
				OptionsFunc(func() []huh.Option[string] {
					versions := getAvailableVersions(setup.Language.Type)
					options := make([]huh.Option[string], len(versions))
					for i, version := range versions {
						options[i] = huh.NewOption(version, version)
					}
					return options
				}, &setup.Language.Type).
				Value(&setup.Language.Version),
		),
	).WithAccessible(accessible)

	err := languageForm.Run()
	if err != nil {
		fmt.Println(errorStyle.Render("Error: " + err.Error()))
		os.Exit(1)
	}

	// Second form: Editor selection
	if setup.Language.Type != "" && setup.Language.Version != "" {
		editorForm := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Development Editors").
					OptionsFunc(func() []huh.Option[string] {
						return getLanguageEditors(setup.Language.Type)
					}, &setup.Language.Type).
					Value(&setup.Language.Editors).
					Filterable(true),
			),
		).WithAccessible(accessible)

		err = editorForm.Run()
		if err != nil {
			fmt.Println(errorStyle.Render("Error: " + err.Error()))
			os.Exit(1)
		}
	}

	// Show installation progress
	setupEnvironment := func() {
		time.Sleep(2 * time.Second)
	}

	_ = spinner.New().
		Title("Setting up your development environment...").
		Accessible(accessible).
		Action(setupEnvironment).
		Run()

	// Print final summary
	{
		var sb strings.Builder
		_, err := fmt.Fprintf(&sb,
			"%s\n\nLanguage: %s\nVersion: %s\nEditors: %s",
			headerStyle.Render("DEV ENVIRONMENT SETUP COMPLETE"),
			highlight.Render(setup.Language.Type),
			highlight.Render(setup.Language.Version),
			highlight.Render(xstrings.EnglishJoin(setup.Language.Editors, true)),
		)
		if err != nil {
			return
		}

		if setup.Language.CurrentVer != "" {
			_, err := fmt.Fprintf(&sb, "\n\nPrevious version: %s",
				subtle.Render(setup.Language.CurrentVer))
			if err != nil {
				return
			}
		}

		fmt.Println(
			lipgloss.NewStyle().
				Width(60).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Padding(1, 2).
				Render(sb.String()),
		)
	}
}
