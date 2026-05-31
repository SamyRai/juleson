package theme

import (
	"github.com/charmbracelet/huh"
)

// Confirm prompts the user for a yes/no response
func Confirm(title string, defaultVal bool) (bool, error) {
	var result bool
	err := huh.NewConfirm().
		Title(title).
		Value(&result).
		Run()
	return result, err
}

// InputString prompts the user for a string input
func InputString(title string, placeholder string) (string, error) {
	var result string
	err := huh.NewInput().
		Title(title).
		Placeholder(placeholder).
		Value(&result).
		Run()
	return result, err
}

// InputSecret prompts the user for a sensitive string
func InputSecret(title string) (string, error) {
	var result string
	err := huh.NewInput().
		Title(title).
		EchoMode(huh.EchoModePassword).
		Value(&result).
		Run()
	return result, err
}
