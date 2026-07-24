package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func replaceLineInFile(filename, oldLine, newLine string) error {
	inputFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	var lines []string
	scanner := bufio.NewScanner(inputFile)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == oldLine {
			lines = append(lines, newLine)
		} else {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	outputFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	makeFileExecutable(filename)
	return nil
}

func replaceLinesInFiles(fileList string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	file, err := os.Open(fileList)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), " ", 4)
		if len(parts) != 4 {
			continue
		}
		entryHostname, filename, oldLine, newLine := parts[0], parts[1], parts[2], parts[3]

		if entryHostname != hostname {
			continue
		}

		if strings.HasPrefix(filename, "https://") {
			continue
		}

		_ = replaceLineInFile(filename, oldLine, newLine)
	}

	return scanner.Err()
}

func makeFileExecutable(filePath string) {
	if !strings.HasSuffix(filePath, ".txt") {
		_ = os.Chmod(filePath, 0755)
	}
}

func printHelp() {
	fmt.Println("Verwendung: ")
	fmt.Println("  Einzelne Datei: <datei> <alte_zeile> <neue_zeile>")
	fmt.Println("  Mehrere Dateien aus Datei: file_list.txt")
	fmt.Println("\nFormat der Datei file_list.txt:")
	fmt.Println("  Jede Zeile: <hostname> <datei> <alte_zeile> <neue_zeile>")
	fmt.Println("\nOptionen:")
	fmt.Println("  --help       Zeigt diese Hilfemeldung an")
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--help" {
		printHelp()
		return
	}

	if len(os.Args) == 2 && strings.HasSuffix(os.Args[1], ".txt") {
		_ = replaceLinesInFiles(os.Args[1])
	} else if len(os.Args) == 4 {
		filename := os.Args[1]
		if !strings.HasPrefix(filename, "https://") {
			_ = replaceLineInFile(filename, os.Args[2], os.Args[3])
		}
	}
}
