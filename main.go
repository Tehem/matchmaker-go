package main

import (
	"fmt"
	"github.com/spf13/viper"
	"matchmaker/commands"
)

func main() {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	// Default sessions config
	viper.SetDefault("sessions.sessionDurationMinutes", 60)
	viper.SetDefault("sessions.minSessionSpacingHours", 8)
	viper.SetDefault("sessions.maxPerPersonPerWeek", 2)
	viper.SetDefault("sessions.sessionPrefix", "Pairing ")
	// Default working hours
	viper.SetDefault("workingHours.timezone", "Europe/Paris")
	viper.SetDefault("workingHours.morning.start.hour", 10)
	viper.SetDefault("workingHours.morning.start.minute", 0)
	viper.SetDefault("workingHours.morning.end.hour", 12)
	viper.SetDefault("workingHours.morning.end.minute", 0)
	viper.SetDefault("workingHours.afternoon.start.hour", 14)
	viper.SetDefault("workingHours.afternoon.start.minute", 0)
	viper.SetDefault("workingHours.afternoon.end.hour", 18)
	viper.SetDefault("workingHours.afternoon.end.minute", 0)

	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("config file not found!\n Please initialiaze it " +
				"with `cp configs/config.json.example configs/config.json` and adjust values accordingly.\n")
		} else {
			panic(fmt.Errorf("fatal error reading config file: %w", err))
		}
	}

	commands.Execute()
}
