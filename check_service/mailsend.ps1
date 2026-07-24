<#
.SYNOPSIS
Sendet die Logdatei eines Services per E-Mail.

.PARAMETER LogFile
Pfad zur Logdatei

.PARAMETER To
Empfänger-Mailadresse

.PARAMETER From
Absender-Mailadresse

.PARAMETER SmtpServer
SMTP-Server (z.B. mail.firma.local)

.PARAMETER Subject
Betreff der Mail. Optional: Platzhalter %SERVICE% und %HOSTNAME%
#>

param(
    [string]$LogFile = "C:\Logs\service.log",
    [string]$ServiceName = "Service",
    [string]$To = "admin@firma.local",
    [string]$From = "monitor@firma.local",
    [string]$SmtpServer = "mail.firma.local",
    [string]$Subject = "Service-Check %SERVICE% – %HOSTNAME%"
)

# Hostname ersetzen
$hostname = $env:COMPUTERNAME
$subjectFinal = $Subject.Replace("%SERVICE%", $ServiceName).Replace("%HOSTNAME%", $hostname)

# Logdatei einlesen
if (!(Test-Path $LogFile)) {
    Write-Host "Logdatei nicht gefunden: $LogFile"
    exit 1
}
$body = Get-Content -Path $LogFile -Raw

# Mail versenden
try {
    Send-MailMessage -From $From -To $To -Subject $subjectFinal -Body $body -SmtpServer $SmtpServer
    Write-Host "Mail erfolgreich gesendet an $To"
} catch {
    Write-Host "Fehler beim Mailversand: $_"
}
