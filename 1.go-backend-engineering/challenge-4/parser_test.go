package main

import (
	"bufio"
	"os"
	"testing"
)

var logfilePath = "log/logs.txt"

func loadLines(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func BenchmarkParseA(b *testing.B) {
	lines := loadLines(logfilePath)
	b.ResetTimer()
	for b.Loop() {
		for _, line := range lines {
			ParseA(line)
		}
	}
}

func BenchmarkParseB(b *testing.B) {
	lines := loadLines(logfilePath)
	b.ResetTimer()
	for b.Loop() {
		for _, line := range lines {
			ParseB(line)
		}
	}
}

func BenchmarkParseC(b *testing.B) {
	lines := loadLines(logfilePath)
	b.ResetTimer()
	for b.Loop() {
		for _, line := range lines {
			results := ParseC(line)
			ReleaseMap(results)
		}
	}
}
