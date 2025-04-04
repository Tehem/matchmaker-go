package main

import (
	"matchmaker/commands"
	"matchmaker/libs/config"
	"matchmaker/libs/util"
)

func main() {
	// Initialize logging
	util.ConfigureLogging()

	// Initialize configuration
	err := config.Initialize()
	if err != nil {
		util.LogError(err, "Failed to initialize configuration")
		return
	}

	commands.Execute()
}
