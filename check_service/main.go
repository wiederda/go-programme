package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const StateStopPending = "StopPending"

type CRLFWriter struct {
	wrapped *os.File
}

func (cw *CRLFWriter) Write(p []byte) (int, error) {
	data := bytes.ReplaceAll(p, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
	return cw.wrapped.Write(data)
}

func normalizePath(p string) string {
	return filepath.FromSlash(filepath.Clean(p))
}

func logLine(logger *log.Logger, msg string) {
	logger.Printf("%s %s", time.Now().Format("2006-01-02 15:04:05"), msg)
}

func getServiceInfo(serviceName string, m *mgr.Mgr) (startType, state string, err error) {
	s, err := m.OpenService(serviceName)
	if err != nil {
		return "", "", err
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		return "", "", err
	}

	cfg, err := s.Config()
	if err != nil {
		return "", "", err
	}

	switch cfg.StartType {
	case mgr.StartAutomatic:
		startType = "Auto"
	case mgr.StartManual:
		startType = "Manuell"
	case mgr.StartDisabled:
		startType = "Deaktiviert"
	default:
		startType = "Unbekannt"
	}

	stateMap := map[svc.State]string{
		svc.Stopped:      "Stopped",
		svc.StopPending:  "StopPending",
		svc.StartPending: "StartPending",
		svc.Running:      "Running",
		svc.Paused:       "Paused",
	}
	state, ok := stateMap[status.State]
	if !ok {
		state = "Unknown"
	}
	return
}

func runTask(taskName string) error {
	cmd := exec.Command("schtasks", "/run", "/TN", taskName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("task '%s' failed: %v (%s)", taskName, err, string(out))
	}
	return nil
}

func showHelp() {
	fmt.Println(`
Usage:
  check_service.exe Service=<Dienstname> Log=<LogPfad> [Task=<Task>] [Mail=<MailTask>] [MailStatus=true|false|error]

Parameter:
  Service=      Name des Windows-Dienstes (Pflicht)
  Log=          Pfad zur Logdatei (UTF-8, CRLF, Pflicht)
  Task=         Optionaler Task, der nur bei StopPending ausgeführt wird (nach Mail)
  Mail=         Optionaler Task zum Versenden von Mail
  MailStatus=   Steuerung für Mail:
                 true  - immer (Standard)
                 false - nie
                 error - nur bei StopPending
Beispiel:
  check_service.exe Service=Wildfly Log=C:\Logs\Wildfly.log Task=RestartWildfly Mail=MailTask MailStatus=error
`)
}

func main() {
	if len(os.Args) == 1 {
		showHelp()
		return
	}

	for _, a := range os.Args[1:] {
		if a == "--help" || a == "-h" {
			showHelp()
			return
		}
	}

	var serviceName, logPath, taskName, mailTask, mailStatus string
	mailStatus = "true"

	for _, a := range os.Args[1:] {
		parts := strings.SplitN(a, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(parts[0])
		val := parts[1]

		switch key {
		case "service":
			serviceName = val
		case "log":
			logPath = val
		case "task":
			taskName = val
		case "mail":
			mailTask = val
		case "mailstatus":
			mailStatus = strings.ToLower(val)
		}
	}

	if serviceName == "" || logPath == "" {
		fmt.Println("Fehler: Service und Log müssen angegeben werden.")
		showHelp()
		return
	}

	logPath = normalizePath(logPath)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Fehler beim Öffnen der Logdatei:", err)
		return
	}
	defer f.Close()

	if info, err := f.Stat(); err == nil && info.Size() == 0 {
		f.Write([]byte{0xEF, 0xBB, 0xBF})
	}

	logger := log.New(&CRLFWriter{wrapped: f}, "", 0)

	m, err := mgr.Connect()
	if err != nil {
		logLine(logger, fmt.Sprintf("%s Unknown Unknown (%v)", serviceName, err))

		if mailTask != "" && (mailStatus == "true" || mailStatus == "error") {
			if err := runTask(mailTask); err == nil {
				logLine(logger, "Mail gesendet")
			} else {
				logLine(logger, err.Error())
			}
		}
		return
	}
	defer m.Disconnect()

	startType, state, err := getServiceInfo(serviceName, m)
	if err != nil {
		logLine(logger, fmt.Sprintf("%s Unknown Unknown (%v)", serviceName, err))

		if mailTask != "" && (mailStatus == "true" || mailStatus == "error") {
			if err := runTask(mailTask); err == nil {
				logLine(logger, "Mail gesendet")
			} else {
				logLine(logger, err.Error())
			}
		}
		return
	}
	logLine(logger, fmt.Sprintf("%s %s %s", serviceName, startType, state))

	// Mail nur gesteuert durch MailStatus
	shouldSend := false

	switch mailStatus {
	case "true":
		shouldSend = true
	case "error":
		shouldSend = (state == StateStopPending)
	}

	if mailTask != "" && shouldSend {
		if err := runTask(mailTask); err == nil {
			logLine(logger, "Mail gesendet")
		} else {
			logLine(logger, err.Error())
		}
	}

	// Task nur bei StopPending und erst nach Mail
	if state == StateStopPending && taskName != "" {
		if err := runTask(taskName); err != nil {
			logLine(logger, err.Error())
		}
	}
}
