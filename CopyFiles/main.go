package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// cleanPath entfernt den Laufwerksbuchstaben (C:) oder den UNC-Server/Share-Teil (\\Server\Share)
// vom absoluten Pfad, um den relativen Pfad für das Zielverzeichnis zu erhalten.
func cleanPath(srcPath string, volume string, isUNC bool) string {
	if isUNC {
		// UNC-Pfad: \\Server\Share\Path\To\File
		// Ziel soll sein: Path\To\File

		// Entferne '\\' am Anfang
		pathWithoutPrefix := srcPath[2:]

		// Teile den Pfad auf: [Server Share Path To File]
		parts := strings.Split(pathWithoutPrefix, `\`)

		if len(parts) >= 3 {
			// Füge alle Teile ab dem dritten (Index 2) wieder zusammen
			return filepath.Join(parts[2:]...)
		}

		// Falls der Pfad nur \\Server\Share ist, ist der relative Pfad leer.
		return ""
	}

	// Normaler Pfad: C:\Path\To\File
	// Ziel soll sein: Path\To\File
	relPath := strings.TrimPrefix(srcPath, volume)
	// Entferne den führenden Backslash, falls vorhanden
	return strings.TrimLeft(relPath, `\`)
}

func printHelp() {
	fmt.Println("Verwendung:")
	fmt.Println("  go run main.go -searchlist=<suchliste.txt> -dest=<zielordner> [-zip=<zipdatei>]")
	fmt.Println()
	fmt.Println("Beschreibung:")
	fmt.Println("  Kopiert Dateien oder Ordner aus der <suchliste.txt> in <zielordner>,")
	fmt.Println("  behält dabei die Original-Ordnerstruktur bei (Laufwerksbuchstabe oder UNC-Pfad entfällt).")
	fmt.Println("  Optional kann ein ZIP-Archiv erstellt werden.")
	fmt.Println()
	fmt.Println("Optionen:")
	fmt.Println("  -searchlist  Pfad zur Textdatei mit absoluten Windows- oder UNC-Pfaden (eine Datei oder ein Ordner pro Zeile)")
	fmt.Println("  -dest        Zielordner für die Kopie")
	fmt.Println("  -zip         Optional: Pfad für das ZIP-Archiv")
	fmt.Println("  -h           Zeigt diese Hilfe an")
	fmt.Println()
	fmt.Println("Hinweise:")
	fmt.Println("  - Die Ordnerstruktur wird beibehalten (nur Laufwerksbuchstabe oder UNC-Server wird entfernt).")
	fmt.Println()
	fmt.Println("🔍 Unschärfe bei Dateinamen (Optimiert):")
	fmt.Println("  - Es wird nach dem **Basisnamen** gesucht, unabhängig von der in der Liste angegebenen Erweiterung.")
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDirRecursive(srcDir, destRoot string, volume string, isUNC bool) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Verwendung der zentralen cleanPath Funktion
		relPath := cleanPath(path, volume, isUNC)
		dstPath := filepath.Join(destRoot, relPath)

		if info.IsDir() {
			os.MkdirAll(dstPath, os.ModePerm)
			return nil
		}

		fmt.Printf("Kopiere %s → %s\n", path, dstPath)
		return copyFile(path, dstPath)
	})
}

func zipFolder(sourceDir, zipFilePath string) error {
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		if path == zipFilePath {
			return nil // Schließe die ZIP-Datei selbst aus
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		w, err := zw.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, f)
		return err
	})
}

func main() {
	help := flag.Bool("h", false, "Zeigt die Hilfe an")
	searchList := flag.String("searchlist", "", "Pfad zur Suchliste")
	destDir := flag.String("dest", "", "Zielordner für die Kopie")
	zipPath := flag.String("zip", "", "Optional: Pfad für das ZIP-Archiv")

	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	if *searchList == "" || *destDir == "" {
		fmt.Println("Fehler: -searchlist und -dest müssen angegeben werden. -h für Hilfe.")
		os.Exit(1)
	}

	// NEU: Zielordner erstellen, falls er nicht existiert
	if err := os.MkdirAll(*destDir, os.ModePerm); err != nil {
		fmt.Printf("Fehler beim Erstellen des Zielordners %s: %v\n", *destDir, err)
		os.Exit(1)
	}

	file, err := os.Open(*searchList)
	if err != nil {
		fmt.Printf("Fehler beim Öffnen der Suchliste: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		srcPath := strings.TrimSpace(scanner.Text())
		if srcPath == "" {
			continue
		}

		info, err := os.Stat(srcPath)

		// --- Optimierte Fuzzy-Logik ---
		if os.IsNotExist(err) {
			dir := filepath.Dir(srcPath)
			base := filepath.Base(srcPath)

			// Erweiterung entfernen, um nach "filename.*" statt "filename.ext.*" zu suchen
			baseWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))

			matches, _ := filepath.Glob(filepath.Join(dir, baseWithoutExt+".*"))

			if len(matches) == 1 {
				fmt.Printf("⚠️  Datei fehlt, verwende ähnliche Datei: %s\n", matches[0])
				srcPath = matches[0]
				info, _ = os.Stat(srcPath)
			} else if len(matches) > 1 {
				fmt.Printf("⚠️  Mehrere mögliche Treffer für %s → überspringe.\n", srcPath)
				continue
			} else {
				fmt.Printf("❌  Datei nicht gefunden: %s\n", srcPath)
				continue
			}
		} else if err != nil {
			fmt.Printf("Fehler beim Prüfen von %s: %v\n", srcPath, err)
			continue
		}
		// --- Ende der Fuzzy-Logik ---

		volume := filepath.VolumeName(srcPath)
		isUNC := strings.HasPrefix(srcPath, `\\`)

		if info.IsDir() {
			fmt.Printf("📁 Kopiere Ordner: %s\n", srcPath)
			if err := copyDirRecursive(srcPath, *destDir, volume, isUNC); err != nil {
				fmt.Printf("Fehler beim Kopieren des Ordners: %v\n", err)
			}
		} else {
			// Verwendung der zentralen cleanPath Funktion
			relPath := cleanPath(srcPath, volume, isUNC)
			dstPath := filepath.Join(*destDir, relPath)

			fmt.Printf("📄 Kopiere Datei: %s → %s\n", srcPath, dstPath)
			if err := copyFile(srcPath, dstPath); err != nil {
				fmt.Printf("Fehler beim Kopieren: %v\n", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Fehler beim Lesen der Suchliste: %v\n", err)
	}

	if *zipPath != "" {
		fmt.Printf("📦 Erstelle ZIP-Archiv: %s\n", *zipPath)
		if err := zipFolder(*destDir, *zipPath); err != nil {
			fmt.Printf("Fehler beim ZIPpen: %v\n", err)
		} else {
			fmt.Println("✅ ZIP erfolgreich erstellt.")
		}
	}
}
