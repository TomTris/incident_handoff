package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Logger interface {
	Log(level LogLevel, message string)
}

type LogFormat struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
}
type ConsoleLogger struct {
	output io.Writer
}

func (c *ConsoleLogger) Log(level LogLevel, message string) {
	logMessage := LogFormat{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Level:   logLevelToStrings[level],
		Message: message,
	}

	fmt.Fprintf(c.output, "%s| %s| %s\n", logMessage.Time, logMessage.Level, logMessage.Message)
}

type FileLogger struct {
	fileName string
}

func (c *FileLogger) Log(level LogLevel, message string) {
	file, err := os.OpenFile(c.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logMessage := LogFormat{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Level:   logLevelToStrings[level],
		Message: message,
	}
	fmt.Fprintf(file, "%s| %s| %s\n", logMessage.Time, logMessage.Level, logMessage.Message)
}

type JsonLogger struct {
	fileName string
}

func (c *JsonLogger) Log(level LogLevel, message string) {
	file, err := os.OpenFile(c.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	logMessage := LogFormat{
		Time:    time.Now().Format("2006-01-02 15:04:05"),
		Level:   logLevelToStrings[level],
		Message: message,
	}

	data, err := json.Marshal(logMessage)
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(file, "\n")
}
