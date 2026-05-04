package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

const (
	outputDir = "logs"
)

type FileResult struct {
	Filename   string
	InfoCount  int
	WarnCount  int
	ErrorCount int
}

type ErrorLog struct {
	Filename string
	Line     string
}

var (
	allErrors []ErrorLog
	mu        sync.Mutex
)

func scanFile(filename string, ch chan<- FileResult, wg *sync.WaitGroup) {
	defer wg.Done()
	fileResult := FileResult{Filename: filename, InfoCount: 0, WarnCount: 0, ErrorCount: 0}
	path := filepath.Join(outputDir, filename)
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.Contains(line, "[INFO ]"):
			fileResult.InfoCount++
		case strings.Contains(line, "[WARN ]"):
			fileResult.WarnCount++
		case strings.Contains(line, "[ERROR]"):
			fileResult.ErrorCount++
			mu.Lock()
			allErrors = append(allErrors, ErrorLog{Filename: filename, Line: line})
			mu.Unlock()
		default:
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	ch <- fileResult
}

func generateGeneralReport(fileResults []FileResult, nameLengthMax int) {
	fmt.Println()
	for _, result := range fileResults {
		fmt.Printf("[worker] %-*s → INFO:%d  WARN:%d  ERROR:%d\n",
			nameLengthMax, result.Filename, result.InfoCount, result.WarnCount, result.ErrorCount)
	}
}

func generateErrorReport(fileResults []FileResult, nameLengthMax int) {
	sort.Slice(fileResults, func(i, j int) bool {
		a := fileResults[i].Filename
		b := fileResults[j].Filename
		return a[len(a)-5:] < b[len(b)-5:]
	})

	totalInfo := 0
	totalWarn := 0
	totalError := 0
	fmt.Println()
	fmt.Println("══════════════════════════════════════════════════════")
	fmt.Println("                    ERROR REPORT")
	fmt.Println("══════════════════════════════════════════════════════")
	fmt.Printf("File %-*s INFO    WARN    ERROR\n", nameLengthMax, "")
	for _, r := range fileResults {
		fmt.Printf("%-*s    %d     %d       %d\n", nameLengthMax+3,
			r.Filename, r.InfoCount, r.WarnCount, r.ErrorCount)
		totalInfo += r.InfoCount
		totalWarn += r.WarnCount
		totalError += r.ErrorCount
	}
	fmt.Println("══════════════════════════════════════════════════════")
	fmt.Printf("%-*s    %d     %d       %d\n", nameLengthMax+3, "TOTAL", totalInfo, totalWarn, totalError)
	fmt.Println("══════════════════════════════════════════════════════")
}

func generateWorstFile(fileResults []FileResult) {
	sort.Slice(fileResults, func(i, j int) bool {
		a := fileResults[i].ErrorCount
		b := fileResults[j].ErrorCount
		return a > b
	})

	fmt.Println()
	fmt.Printf("Worst file: %s (%d errors)\n", fileResults[0].Filename, fileResults[0].ErrorCount)
}

func generateErrorLines() {
	fmt.Println("First 5 ERROR lines:")
	for i := 0; i < 5 && i < len(allErrors); i++ {
		fmt.Printf("  [%s] %s\n", allErrors[i].Filename, allErrors[i].Line)
	}
}

func main() {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		log.Fatal(err)
	}
	nameLengthMax := 0
	fmt.Println("Scanning 7 files concurrently...")
	ch := make(chan FileResult, len(entries))
	var wg sync.WaitGroup
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		nameLengthMax = max(nameLengthMax, len(entry.Name()))
		wg.Add(1)
		go scanFile(entry.Name(), ch, &wg)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	fileResults := []FileResult{}
	for result := range ch {
		fileResults = append(fileResults, result)
	}
	generateGeneralReport(fileResults, nameLengthMax)
	generateErrorReport(fileResults, nameLengthMax)
	generateWorstFile(fileResults)
	generateErrorLines()
}
