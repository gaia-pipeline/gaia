package scheduler

import "testing"

func TestFindRubyGemName(t *testing.T) {
	// Adjust gubyGemName (might be not available in test container)
	rubyGemName = "echo"

	// Overwrite gem commands
	findRubyGemCommands = []string{"name: testruby"}

	// Run command and compare output
	gemName, err := findRubyGemName("")
	if err != nil {
		t.Errorf("error thrown during findRubyGemName: %s", err.Error())
	}
	if gemName != "testruby" {
		t.Errorf("Gem name should be 'testruby' but was %s", gemName)
	}
}
