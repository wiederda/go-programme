package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yuin/goldmark/v2/ast"
	"github.com/yuin/goldmark/v2/parser"
)

const version = "1.0.0"

type section struct {
	Name  string
	Start int
	End   int
}

type result struct {
	File string

	FunctionsFound     int
	FunctionsProcessed int
	FunctionsSkipped   int
	FunctionsFailed    int

	ProblemFunctions []string

	Changed bool

	Skipped    bool
	SkipReason string

	Error error
}

func main() {
	args := os.Args[1:]

	inputs, outputDir, excludes, dryRun, showVersion, showHelp, err :=
		parseArguments(args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fehler: %v\n", err)
		fmt.Println()
		printUsage()
		os.Exit(1)
	}

	if showHelp {
		printUsage()
		return
	}

	if showVersion {
		fmt.Printf("vbx-doc-sort %s\n", version)
		return
	}

	if err := run(
		inputs,
		outputDir,
		excludes,
		dryRun,
	); err != nil {
		fmt.Fprintf(os.Stderr, "\nFehler: %v\n", err)
		os.Exit(1)
	}
}

// ------------------------------------------------------------
// Argumente
// ------------------------------------------------------------

func parseArguments(
	args []string,
) (
	inputs []string,
	outputDir string,
	excludes []string,
	dryRun bool,
	showVersion bool,
	showHelp bool,
	err error,
) {
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch strings.ToLower(arg) {

		case "-h", "--help":
			showHelp = true

		case "-version", "--version":
			showVersion = true

		case "-dry-run", "--dry-run":
			dryRun = true

		case "-exclude", "--exclude":
			if i+1 >= len(args) {
				return nil, "", nil, false, false, false,
					errors.New(
						"-exclude benötigt einen Dateinamen",
					)
			}

			i++

			exclude := strings.TrimSpace(args[i])

			if exclude == "" {
				return nil, "", nil, false, false, false,
					errors.New(
						"-exclude benötigt einen Dateinamen",
					)
			}

			excludes = append(
				excludes,
				filepath.Base(exclude),
			)

		default:
			if strings.HasPrefix(arg, "-") {
				return nil, "", nil, false, false, false,
					fmt.Errorf(
						"unbekannte Option %q",
						arg,
					)
			}

			positional = append(
				positional,
				arg,
			)
		}
	}

	if showHelp || showVersion {
		return
	}

	if len(positional) < 2 {
		return nil, "", nil, false, false, false,
			errors.New(
				"mindestens ein Input und ein Ausgabeverzeichnis erforderlich",
			)
	}

	/*
		Das letzte Positionsargument ist immer
		das Ausgabeverzeichnis.
	*/
	outputDir = positional[len(positional)-1]

	/*
		Alle vorherigen Argumente sind Input-Dateien
		oder Input-Verzeichnisse.
	*/
	inputs = positional[:len(positional)-1]

	return
}

// ------------------------------------------------------------
// Hilfe
// ------------------------------------------------------------

func printUsage() {
	fmt.Println("vbx-doc-sort - Sortiert VBX Markdown-Dokumentation")
	fmt.Println()
	fmt.Println("Verwendung:")
	fmt.Println("  vbx-doc-sort <input>... <output> [optionen]")
	fmt.Println()
	fmt.Println("Input:")
	fmt.Println("  Verzeichnis       Alle .md-Dateien des Verzeichnisses")
	fmt.Println("  Datei             Nur diese .md-Datei")
	fmt.Println("  Mehrere Dateien   Mehrere .md-Dateien gleichzeitig")
	fmt.Println()
	fmt.Println("Das letzte Positionsargument ist immer das")
	fmt.Println("Ausgabeverzeichnis.")
	fmt.Println()
	fmt.Println("Optionen:")
	fmt.Println("  -h, --help")
	fmt.Println("      Diese Hilfe anzeigen")
	fmt.Println()
	fmt.Println("  -version, --version")
	fmt.Println("      Version anzeigen")
	fmt.Println()
	fmt.Println("  -dry-run, --dry-run")
	fmt.Println("      Nur prüfen, keine Dateien schreiben")
	fmt.Println()
	fmt.Println("  -exclude, --exclude <datei>")
	fmt.Println("      Datei ausschließen; kann mehrfach angegeben werden")
	fmt.Println()
	fmt.Println("Beispiele:")
	fmt.Println()
	fmt.Println(`  vbx-doc-sort.exe "C:\test\md" "C:\test\sortiert"`)
	fmt.Println()
	fmt.Println(`  vbx-doc-sort.exe "C:\test\md\cert.md" "C:\test\sortiert"`)
	fmt.Println()
	fmt.Println(`  vbx-doc-sort.exe "C:\test\md\cert.md" "C:\test\md\7z.md" "C:\test\sortiert"`)
	fmt.Println()
	fmt.Println(`  vbx-doc-sort.exe "C:\test\md" "C:\test\sortiert" -exclude Allgemein.md`)
	fmt.Println()
	fmt.Println(`  vbx-doc-sort.exe "C:\test\md" "C:\test\sortiert" -exclude Allgemein.md -exclude README.md`)
	fmt.Println()
	fmt.Println(`  vbx-doc-sort.exe "C:\test\md" "C:\test\sortiert" -dry-run`)
	fmt.Println()
	fmt.Println(`  vbx-doc-sort.exe -dry-run -exclude Allgemein.md "C:\test\md" "C:\test\sortiert"`)
	fmt.Println()
}

// ------------------------------------------------------------
// Hauptverarbeitung
// ------------------------------------------------------------

func run(
	inputs []string,
	outputDir string,
	excludes []string,
	dryRun bool,
) error {

	files, err := collectFiles(
		inputs,
		excludes,
	)

	if err != nil {
		return err
	}

	if len(files) == 0 {
		fmt.Println("Keine Markdown-Dateien gefunden.")
		return nil
	}

	if !dryRun {
		if err := os.MkdirAll(
			outputDir,
			0755,
		); err != nil {
			return fmt.Errorf(
				"Ausgabeverzeichnis %q erstellen: %w",
				outputDir,
				err,
			)
		}
	}

	fmt.Println("vbx-doc-sort")
	fmt.Println()

	fmt.Printf(
		"Eingaben:     %d\n",
		len(inputs),
	)

	fmt.Printf(
		"Dateien:      %d\n",
		len(files),
	)

	fmt.Printf(
		"Ausgabe:      %s\n",
		outputDir,
	)

	if len(excludes) > 0 {
		fmt.Printf(
			"Ausnahmen:    %d\n",
			len(excludes),
		)
	}

	if dryRun {
		fmt.Println(
			"Modus:        DRY-RUN",
		)
	}

	fmt.Println()

	var results []result

	for _, file := range files {

		outputPath := filepath.Join(
			outputDir,
			filepath.Base(file),
		)

		r := processFile(
			file,
			outputPath,
			dryRun,
		)

		results = append(
			results,
			r,
		)

		printResult(r)
	}

	printSummary(results)

	errorCount := countErrors(results)

	if errorCount > 0 {
		return fmt.Errorf(
			"%d Datei(en) konnten nicht verarbeitet werden",
			errorCount,
		)
	}

	return nil
}

// ------------------------------------------------------------
// Dateien sammeln
// ------------------------------------------------------------

func collectFiles(
	inputs []string,
	excludes []string,
) ([]string, error) {

	var files []string

	seen := make(
		map[string]bool,
	)

	for _, input := range inputs {

		info, err := os.Stat(input)

		if err != nil {
			return nil, fmt.Errorf(
				"Input %q: %w",
				input,
				err,
			)
		}

		/*
			--------------------------------------------
			Input ist ein Verzeichnis
			--------------------------------------------
		*/
		if info.IsDir() {

			entries, err := os.ReadDir(input)

			if err != nil {
				return nil, fmt.Errorf(
					"Verzeichnis %q lesen: %w",
					input,
					err,
				)
			}

			for _, entry := range entries {

				if entry.IsDir() {
					continue
				}

				filename := entry.Name()

				if !strings.EqualFold(
					filepath.Ext(filename),
					".md",
				) {
					continue
				}

				if isExcluded(
					filename,
					excludes,
				) {
					continue
				}

				path := filepath.Join(
					input,
					filename,
				)

				key := normalizePath(path)

				if seen[key] {
					continue
				}

				seen[key] = true

				files = append(
					files,
					path,
				)
			}

			continue
		}

		/*
			--------------------------------------------
			Input ist eine einzelne Datei
			--------------------------------------------
		*/

		if !strings.EqualFold(
			filepath.Ext(input),
			".md",
		) {
			return nil, fmt.Errorf(
				"%q ist keine Markdown-Datei",
				input,
			)
		}

		filename := filepath.Base(input)

		if isExcluded(
			filename,
			excludes,
		) {
			continue
		}

		key := normalizePath(input)

		if seen[key] {
			continue
		}

		seen[key] = true

		files = append(
			files,
			input,
		)
	}

	/*
		Determinierte Reihenfolge der Dateien.
	*/
	sort.Slice(
		files,
		func(i, j int) bool {

			return strings.ToLower(
				filepath.Base(files[i]),
			) < strings.ToLower(
				filepath.Base(files[j]),
			)
		},
	)

	return files, nil
}

// ------------------------------------------------------------
// Pfad normalisieren
// ------------------------------------------------------------

func normalizePath(path string) string {

	abs, err := filepath.Abs(path)

	if err != nil {
		return filepath.Clean(path)
	}

	return filepath.Clean(abs)
}

// ------------------------------------------------------------
// Datei ausgeschlossen?
// ------------------------------------------------------------

func isExcluded(
	filename string,
	excludes []string,
) bool {

	for _, excluded := range excludes {

		if strings.EqualFold(
			filename,
			filepath.Base(excluded),
		) {
			return true
		}
	}

	return false
}

// ------------------------------------------------------------
// Einzelne Datei verarbeiten
// ------------------------------------------------------------

func processFile(
	inputPath string,
	outputPath string,
	dryRun bool,
) result {

	filename := filepath.Base(inputPath)

	source, err := os.ReadFile(
		inputPath,
	)

	if err != nil {
		return result{
			File:  filename,
			Error: err,
		}
	}

	sections, headerEnd, err := parseMarkdown(
		source,
	)

	if err != nil {
		return result{
			File:  filename,
			Error: err,
		}
	}

	/*
		Keine Funktionsabschnitte.

		Die Datei wurde untersucht, enthält aber
		keine erkennbaren VBX-Funktionsabschnitte.
	*/
	if len(sections) == 0 {
		return result{
			File:       filename,
			Skipped:    true,
			SkipReason: "keine Funktionsabschnitte gefunden",
		}
	}

	functionsFound := len(sections)

	/*
		Sortierte Kopie erstellen.
	*/
	sorted := make(
		[]section,
		len(sections),
	)

	copy(
		sorted,
		sections,
	)

	sort.SliceStable(
		sorted,
		func(i, j int) bool {

			left := strings.ToLower(
				sorted[i].Name,
			)

			right := strings.ToLower(
				sorted[j].Name,
			)

			return left < right
		},
	)

	/*
		Prüfen, ob die Datei bereits korrekt sortiert ist.
	*/
	changed := false

	for i := range sections {

		if sections[i].Start != sorted[i].Start {
			changed = true
			break
		}
	}

	functionsProcessed := functionsFound
	functionsSkipped := 0
	functionsFailed := 0

	/*
		Die Datei ist bereits korrekt sortiert.
	*/
	if !changed {
		return result{
			File:               filename,
			FunctionsFound:     functionsFound,
			FunctionsProcessed: functionsProcessed,
			FunctionsSkipped:   functionsSkipped,
			FunctionsFailed:    functionsFailed,
			Changed:            false,
		}
	}

	/*
		Dry-Run:
		Nur analysieren, nichts schreiben.
	*/
	if dryRun {
		return result{
			File:               filename,
			FunctionsFound:     functionsFound,
			FunctionsProcessed: functionsProcessed,
			FunctionsSkipped:   functionsSkipped,
			FunctionsFailed:    functionsFailed,
			Changed:            true,
		}
	}

	/*
		Neue Datei aufbauen.

		Der Header bleibt exakt unverändert.
	*/
	var output strings.Builder

	output.Write(
		source[:headerEnd],
	)

	/*
		Die sortierten Funktionsblöcke werden direkt
		aus der Originalquelle übernommen.

		Dadurch bleiben Inhalt und Formatierung
		innerhalb der Blöcke unverändert.
	*/
	for _, item := range sorted {

		output.Write(
			source[item.Start:item.End],
		)
	}

	if err := os.WriteFile(
		outputPath,
		[]byte(output.String()),
		0644,
	); err != nil {

		return result{
			File:               filename,
			FunctionsFound:     functionsFound,
			FunctionsProcessed: 0,
			FunctionsSkipped:   functionsSkipped,
			FunctionsFailed:    functionsFound,
			ProblemFunctions:   sectionNames(sections),
			Error:              err,
		}
	}

	return result{
		File:               filename,
		FunctionsFound:     functionsFound,
		FunctionsProcessed: functionsProcessed,
		FunctionsSkipped:   functionsSkipped,
		FunctionsFailed:    functionsFailed,
		Changed:            true,
	}
}

// ------------------------------------------------------------
// Markdown analysieren
// ------------------------------------------------------------

func parseMarkdown(
	source []byte,
) ([]section, int, error) {

	md := parser.New()

	document := md.Parse(
		source,
	)

	if document == nil {
		return nil, 0, errors.New(
			"Markdown konnte nicht geparst werden",
		)
	}

	var headings []*ast.Heading

	/*
		Wir suchen ausschließlich echte Heading-Nodes
		im Markdown-AST.

		Wichtig:
		Ein ## innerhalb eines fenced Codeblocks wird
		von Goldmark nicht als ast.Heading erkannt.

		Dadurch können Beispiele wie:

		    ```vbx
		    ## kein echter Heading
		    Print Format(file.AccessTime(path), "YYYY-MM-DD HH:mm:ss")
		    ```

		die Sortierung nicht beeinflussen.
	*/
	for node := document.FirstChild(); node != nil; node = node.NextSibling() {

		heading, ok := node.(*ast.Heading)

		if !ok {
			continue
		}

		if heading.Level != 2 &&
			heading.Level != 3 {
			continue
		}

		name := headingLine(
			heading.Pos(),
			source,
		)

		name = functionName(name)

		if name == "" {
			continue
		}

		headings = append(headings, heading)
	}

	if len(headings) == 0 {
		return nil, 0, nil
	}

	/*
		Alles vor der ersten Funktionsüberschrift
		bleibt unverändert.
	*/
	headerEnd := headings[0].Pos()

	sections := make(
		[]section,
		0,
		len(headings),
	)

	for i, heading := range headings {

		start := heading.Pos()

		end := len(source)

		if i+1 < len(headings) {
			end = headings[i+1].Pos()
		}

		name := headingLine(
			heading.Pos(),
			source,
		)

		name = functionName(name)

		if name == "" {
			continue
		}

		sections = append(
			sections,
			section{
				Name:  name,
				Start: start,
				End:   end,
			},
		)
	}

	return sections, headerEnd, nil
}

// ------------------------------------------------------------
// Überschrift aus Originalquelle lesen
// ------------------------------------------------------------

func headingLine(
	pos int,
	source []byte,
) string {

	if pos < 0 || pos >= len(source) {
		return ""
	}

	end := pos

	for end < len(source) &&
		source[end] != '\n' &&
		source[end] != '\r' {

		end++
	}

	line := strings.TrimSpace(
		string(source[pos:end]),
	)

	/*
		Markdown-Überschrift entfernen.

		Beispiele:

		    ## app.StartupPath()

		    ### `cert.GenerateKey(...)`
	*/
	line = strings.TrimLeft(
		line,
		"#",
	)

	line = strings.TrimSpace(line)

	/*
		Inline-Code-Markierung entfernen.

		Beispiel:

		    `cert.GenerateKey(...)`

		wird:

		    cert.GenerateKey(...)
	*/
	line = strings.TrimSpace(
		strings.Trim(line, "`"),
	)

	/*
		Optionalen abschließenden Heading-Marker
		entfernen.

		Markdown erlaubt beispielsweise:

		    ## cert.GenerateKey(...) ##

		Der Marker gehört nicht zum Funktionsnamen.
	*/
	line = strings.TrimSpace(
		strings.TrimRight(line, "#"),
	)

	return strings.TrimSpace(line)
}

// ------------------------------------------------------------
// Funktionsname extrahieren
// ------------------------------------------------------------

func functionName(
	name string,
) string {

	name = strings.TrimSpace(name)

	if name == "" {
		return ""
	}

	/*
		Alles ab der ersten öffnenden Klammer
		abschneiden.

		Beispiel:

		    cert.GenerateKey(outFile, algo, bits)

		wird:

		    cert.GenerateKey
	*/
	pos := strings.Index(
		name,
		"(",
	)

	if pos < 0 {
		return ""
	}

	name = strings.TrimSpace(
		name[:pos],
	)

	/*
		Ein VBX-Funktionsname darf sowohl einen
		Namespace enthalten als auch global sein.

		Beispiele:

		    cert.GenerateKey
		    app.StartupPath
		    array.Sort
		    Left
		    Right
		    Format
	*/

	if name == "" {
		return ""
	}

	/*
		Keine Leerzeichen im Funktionsnamen.

		Damit können normale Überschriften oder
		versehentliche Treffer nicht als Funktion
		interpretiert werden.
	*/
	if strings.ContainsAny(
		name,
		" \t\r\n",
	) {
		return ""
	}

	return name
}

// ------------------------------------------------------------
// Abschnittsnamen
// ------------------------------------------------------------

func sectionNames(
	sections []section,
) []string {

	names := make(
		[]string,
		0,
		len(sections),
	)

	for _, section := range sections {
		names = append(names, section.Name)
	}

	return names
}

// ------------------------------------------------------------
// Ergebnis ausgeben
// ------------------------------------------------------------

func printResult(
	r result,
) {

	if r.Error != nil {

		fmt.Printf(
			"ERROR %-25s %d Funktionen gefunden, %d verarbeitet, %d fehlerhaft\n",
			r.File,
			r.FunctionsFound,
			r.FunctionsProcessed,
			r.FunctionsFailed,
		)

		if len(r.ProblemFunctions) > 0 {

			fmt.Println()
			fmt.Println("  Fehlerhafte Funktionen:")

			for _, name := range r.ProblemFunctions {

				fmt.Printf(
					"    - %s\n",
					name,
				)
			}
		}

		return
	}

	if r.Skipped {

		fmt.Printf(
			"SKIP  %-25s %s\n",
			r.File,
			r.SkipReason,
		)

		return
	}

	if !r.Changed {

		fmt.Printf(
			"OK    %-25s %d Funktionen gefunden, %d verarbeitet - bereits sortiert\n",
			r.File,
			r.FunctionsFound,
			r.FunctionsProcessed,
		)

		return
	}

	fmt.Printf(
		"OK    %-25s %d Funktionen gefunden, %d verarbeitet - Reihenfolge geändert\n",
		r.File,
		r.FunctionsFound,
		r.FunctionsProcessed,
	)
}

// ------------------------------------------------------------
// Zusammenfassung
// ------------------------------------------------------------

func printSummary(
	results []result,
) {

	var files int

	var functionsFound int
	var functionsProcessed int
	var functionsSkipped int
	var functionsFailed int

	var filesChanged int
	var filesUnchanged int
	var filesSkipped int
	var filesFailed int

	for _, r := range results {

		if r.Error != nil {

			filesFailed++

			functionsFound += r.FunctionsFound
			functionsProcessed += r.FunctionsProcessed
			functionsSkipped += r.FunctionsSkipped
			functionsFailed += r.FunctionsFailed

			continue
		}

		if r.Skipped {

			filesSkipped++

			continue
		}

		files++

		functionsFound += r.FunctionsFound
		functionsProcessed += r.FunctionsProcessed
		functionsSkipped += r.FunctionsSkipped
		functionsFailed += r.FunctionsFailed

		if r.Changed {
			filesChanged++
		} else {
			filesUnchanged++
		}
	}

	fmt.Println()
	fmt.Println("Zusammenfassung:")

	fmt.Printf(
		"  Dateien:                 %d\n",
		files,
	)

	fmt.Printf(
		"  Funktionen gefunden:    %d\n",
		functionsFound,
	)

	fmt.Printf(
		"  Funktionen verarbeitet: %d\n",
		functionsProcessed,
	)

	fmt.Printf(
		"  Funktionen übersprungen: %d\n",
		functionsSkipped,
	)

	fmt.Printf(
		"  Funktionen fehlerhaft:   %d\n",
		functionsFailed,
	)

	fmt.Println()

	fmt.Printf(
		"  Dateien sortiert:         %d\n",
		filesChanged,
	)

	fmt.Printf(
		"  Dateien unverändert:      %d\n",
		filesUnchanged,
	)

	fmt.Printf(
		"  Dateien übersprungen:     %d\n",
		filesSkipped,
	)

	fmt.Printf(
		"  Dateien mit Fehlern:      %d\n",
		filesFailed,
	)

	/*
		Übersprungene Dateien anzeigen.
	*/
	if filesSkipped > 0 {

		fmt.Println()
		fmt.Println("Übersprungen:")

		for _, r := range results {

			if !r.Skipped {
				continue
			}

			fmt.Printf(
				"  - %s: %s\n",
				r.File,
				r.SkipReason,
			)
		}
	}

	/*
		Fehlerhafte Dateien und Funktionen anzeigen.
	*/
	if filesFailed > 0 {

		fmt.Println()
		fmt.Println("Fehler:")

		for _, r := range results {

			if r.Error == nil {
				continue
			}

			fmt.Printf(
				"  %s:\n",
				r.File,
			)

			if len(r.ProblemFunctions) == 0 {

				fmt.Printf(
					"    - %v\n",
					r.Error,
				)

				continue
			}

			for _, name := range r.ProblemFunctions {

				fmt.Printf(
					"    - %s\n",
					name,
				)
			}

			fmt.Printf(
				"    Ursache: %v\n",
				r.Error,
			)
		}
	}
}

// ------------------------------------------------------------
// Fehler zählen
// ------------------------------------------------------------

func countErrors(
	results []result,
) int {

	count := 0

	for _, r := range results {

		if r.Error != nil {
			count++
		}
	}

	return count
}
