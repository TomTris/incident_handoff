package main

import "os"

// Enums using this way
// best practice: create a type like this before using iota as below
type LogLevel int

const (
	INFO LogLevel = iota // iota means, first 0, then other 1, 2 ,,,
	DEBUG
	WARN
	ERROR
)

// This is new. With this one, it's also 0,1,2,3
var logLevelToStrings = []string{"INFO", "DEBUG", "WARN", "ERROR"}

func RunApp(l Logger) {
	l.Log(INFO, "The computer is starting")
	l.Log(DEBUG, "There is a bug")
	l.Log(WARN, "There is race possibility. Please have a look!")
	l.Log(ERROR, "The computer is shutdown unexpectedly!")

}
func main() {

	RunApp((&ConsoleLogger{os.Stdout}))
	RunApp((&FileLogger{"log.log"}))
	RunApp((&JsonLogger{fileName: "log.json"}))
}
