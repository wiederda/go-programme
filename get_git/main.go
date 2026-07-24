package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

type RepoType string

const (
	RepoAuto    RepoType = "auto"
	RepoGitLab  RepoType = "gitlab"
	RepoGitea   RepoType = "gitea"
	RepoForgejo RepoType = "forgejo"
)

func showHelp() {
	fmt.Println("Verwendung:")
	fmt.Println("  <Programm> <LISTE_DATEI_URL> <LISTE_DATEI> [TOKEN_DATEI] [gitlab|gitea|forgejo]")
	fmt.Println("")
	fmt.Println("REPO_TYP (optional):")
	fmt.Println("  gitlab | gitea | forgejo")
	fmt.Println("")
	fmt.Println("Ohne Angabe erfolgt die Erkennung strikt anhand der URL:")
	fmt.Println("  1. Struktur:")
	fmt.Println("     - GitLab:   enthält '/-/raw/'")
	fmt.Println("     - Gitea:    enthält '/raw/' (ohne '/-/raw/')")
	fmt.Println("  2. Hostname:")
	fmt.Println("     - beginnt mit 'gitlab.' oder 'gitlab-'")
	fmt.Println("")
	fmt.Println("Ist die URL danach nicht eindeutig, bricht das Programm ab.")
}

func readTokenFromFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

func detectRepoType(rawURL string, forced RepoType) RepoType {
	if forced != RepoAuto {
		return forced
	}

	lower := strings.ToLower(rawURL)

	// 1. Struktur
	if strings.Contains(lower, "/-/raw/") {
		return RepoGitLab
	}
	if strings.Contains(lower, "/raw/") && !strings.Contains(lower, "/-/raw/") {
		return RepoGitea
	}

	// 2. Hostname
	if u, err := url.Parse(rawURL); err == nil {
		host := strings.ToLower(u.Host)
		if strings.HasPrefix(host, "gitlab.") || strings.HasPrefix(host, "gitlab-") {
			return RepoGitLab
		}
	}

	return RepoAuto
}

func generateGitLabAPIURL(rawURL string) (string, error) {
	parts := strings.SplitN(rawURL, "/-/raw/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("keine gültige GitLab raw URL")
	}

	u, err := url.Parse(parts[0])
	if err != nil {
		return "", err
	}

	projectPath := strings.TrimPrefix(u.Path, "/")
	fileParts := strings.SplitN(parts[1], "/", 2)
	if len(fileParts) != 2 {
		return "", fmt.Errorf("ungültiger Dateipfad")
	}

	branch := fileParts[0]
	filePath := url.PathEscape(fileParts[1])

	return fmt.Sprintf(
		"%s://%s/api/v4/projects/%s/repository/files/%s/raw?ref=%s",
		u.Scheme,
		u.Host,
		url.PathEscape(projectPath),
		filePath,
		branch,
	), nil
}

func hardFail(msg string) {
	fmt.Fprintln(os.Stderr, "FEHLER:", msg)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 || os.Args[1] == "--help" {
		showHelp()
		return
	}

	listURL := os.Args[1]
	listFile := os.Args[2]

	var token string
	if len(os.Args) >= 4 {
		token = readTokenFromFile(os.Args[3])
	}

	repoType := RepoAuto
	if len(os.Args) >= 5 {
		switch strings.ToLower(os.Args[4]) {
		case "gitlab":
			repoType = RepoGitLab
		case "gitea":
			repoType = RepoGitea
		case "forgejo":
			repoType = RepoForgejo
		default:
			hardFail("unbekannter REPO_TYP")
		}
	}

	resolved := detectRepoType(listURL, repoType)
	if resolved == RepoAuto {
		hardFail("URL nicht eindeutig – REPO_TYP angeben")
	}

	downloadURL := listURL
	var headerName, headerValue string

	switch resolved {
	case RepoGitLab:
		apiURL, err := generateGitLabAPIURL(listURL)
		if err != nil {
			hardFail(err.Error())
		}
		downloadURL = apiURL
		headerName = "PRIVATE-TOKEN"
		headerValue = token

	case RepoGitea, RepoForgejo:
		headerName = "Authorization"
		headerValue = "token " + token
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		hardFail(err.Error())
	}

	if token != "" {
		req.Header.Set(headerName, headerValue)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		hardFail(err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		hardFail(fmt.Sprintf("HTTP %d beim Laden der Liste", resp.StatusCode))
	}

	out, err := os.Create(listFile)
	if err != nil {
		hardFail(err.Error())
	}
	defer out.Close()

	_, _ = out.ReadFrom(resp.Body)

	// --- Verarbeitung der Liste ---
	f, err := os.Open(listFile)
	if err != nil {
		hardFail(err.Error())
	}
	defer f.Close()

	hostname, _ := os.Hostname()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		target, fileURL, filename := fields[0], fields[1], fields[2]
		if target != "ALL" && target != hostname {
			continue
		}

		args := []string{"-O", filename}
		if token != "" {
			switch resolved {
			case RepoGitLab:
				args = append([]string{"--header", "PRIVATE-TOKEN: " + token}, args...)
			case RepoGitea, RepoForgejo:
				args = append([]string{"--header", "Authorization: token " + token}, args...)
			}
		}
		args = append(args, fileURL)

		_ = exec.Command("wget", args...).Run()
		_ = os.Chmod(filename, 0755)
	}
}
