package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"matchmaker/internal/commands"

	"github.com/spf13/viper"
)

func init() {
	// Initialize configuration
	if err := initConfig(); err != nil {
		slog.Error("Failed to initialize configuration", "error", err)
		os.Exit(1)
	}
}

func initConfig() error {
	viper.AddConfigPath("./configs")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

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

	// Read environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			slog.Warn("Config file not found", "error", err)
			slog.Info("Please initialize it with `cp configs/config.yml.example configs/config.yml` and adjust values accordingly")
		} else {
			return fmt.Errorf("fatal error reading config file: %w", err)
		}
	}

	return nil
}

func main() {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	// Execute the root command
	if err := commands.RootCmd.ExecuteContext(ctx); err != nil {
		slog.Error("Command execution failed", "error", err)
		os.Exit(1)
	}
}
