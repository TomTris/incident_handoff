package main

// import (
// 	"os"
// 	"testing"
// )

// func TestLoadConfig(t *testing.T) {
// 	t.Run("check default", func(t *testing.T) {
// 		c := loadConfig()
// 		if c.Port != "8080" {
// 			t.Errorf("default Port not correct")
// 		}
// 		if c.LogLevel != "info" {
// 			t.Errorf("default LOG LEVEL not correct")
// 		}
// 		if c.Environment != "development" {
// 			t.Errorf("default Environment not correct")
// 		}
// 	})
// 	t.Run("check default", func(t *testing.T) {
// 		os.Setenv("HANDOFF_PORT", "7998")
// 		os.Setenv("HANDOFF_LOG_LEVEL", "debug")
// 		os.Setenv("HANDOFF_ENV", "production")
// 		c := loadConfig()
// 		if c.Port != "7998" {
// 			t.Errorf("default Port not correct")
// 		}
// 		if c.LogLevel != "debug" {
// 			t.Errorf("default LOG LEVEL not correct")
// 		}
// 		if c.Environment != "production" {
// 			t.Errorf("default Environment not correct")
// 		}
// 	})
// }
