// Лёгкий установщик kök Edit для Premiere Pro (Windows). Сам скачивает свежую версию с GitHub,
// проверяет подпись kök и ставит без прав администратора:
// плагин → %APPDATA%\Adobe\CEP\extensions\kok-edit, разрешение расширений → HKCU\Software\Adobe\CSXS.*
package main

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

var channel = "main" // при сборке тестового установщика: -X main.channel=test

const (
	extDir    = "kok-edit"
	repo      = "KolenkeMedia/kok-edit-releases"
	publicKey = "MCowBQYDK2VwAyEAjhdCyWn6yr5lCkRTLjODTJYlSp1+iVa+4Oxg9QipUeg="
	uninKey   = `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\kokEdit`
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
)

func run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func premiereRunning() bool {
	out, _ := run("tasklist", "/FI", "IMAGENAME eq Adobe Premiere Pro.exe", "/NH")
	return strings.Contains(strings.ToLower(out), "premiere")
}
func target() string {
	return filepath.Join(os.Getenv("APPDATA"), "Adobe", "CEP", "extensions", extDir)
}
func workDir() string {
	h, _ := os.UserHomeDir()
	return filepath.Join(h, "kokEdit")
}

// ---------- скачивание ----------
var client = &http.Client{Timeout: 10 * time.Minute}

func get(url string, progress func(got, total int64)) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "kok-edit-setup")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("нет связи с интернетом (%v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("сервер ответил %d (%s)", resp.StatusCode, url)
	}
	var buf bytes.Buffer
	tmp := make([]byte, 256*1024)
	var got int64
	for {
		n, err := resp.Body.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
			got += int64(n)
			if progress != nil {
				progress(got, resp.ContentLength)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("загрузка прервалась: %v", err)
		}
	}
	return buf.Bytes(), nil
}

type comp struct {
	Key    string `json:"key"`
	URL    string `json:"url"`
	Sha256 string `json:"sha256"`
	Size   int64  `json:"size"`
}
type release struct {
	Version    string          `json:"version"`
	Components map[string]comp `json:"components"`
}

func verify(data, sig []byte) bool {
	der, _ := base64.StdEncoding.DecodeString(publicKey)
	k, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return false
	}
	s, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(sig)))
	if err != nil {
		return false
	}
	return ed25519.Verify(k.(ed25519.PublicKey), data, s)
}

// распаковать zip-часть в папку плагина → список файлов
func unpack(data []byte, dst string) ([]string, error) {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	var files []string
	for _, f := range zr.File {
		rel := strings.ReplaceAll(f.Name, "\\", "/")
		if f.FileInfo().IsDir() || rel == "" || strings.Contains(rel, "..") {
			continue
		}
		p := filepath.Join(dst, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return nil, err
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		w, err := os.Create(p)
		if err != nil {
			rc.Close()
			return nil, err
		}
		_, err = io.Copy(w, rc)
		rc.Close()
		w.Close()
		if err != nil {
			return nil, err
		}
		files = append(files, rel)
	}
	return files, nil
}

func debugMode() {
	for v := 9; v <= 14; v++ { // CEP 9–12 — Premiere 2019–2025, 13–14 — на будущее
		run("reg", "add", fmt.Sprintf(`HKCU\Software\Adobe\CSXS.%d`, v), "/v", "PlayerDebugMode", "/t", "REG_SZ", "/d", "1", "/f")
	}
}

func registerUninstall(dst, version string) {
	self, err := os.Executable()
	if err != nil {
		return
	}
	un := filepath.Join(dst, "Удалить kök Edit.exe")
	if b, err := os.ReadFile(self); err == nil {
		os.WriteFile(un, b, 0o755)
	}
	set := func(name, typ, val string) { run("reg", "add", uninKey, "/v", name, "/t", typ, "/d", val, "/f") }
	set("DisplayName", "REG_SZ", "kök Edit for Premiere Pro")
	set("DisplayVersion", "REG_SZ", version)
	set("Publisher", "REG_SZ", "kök")
	set("URLInfoAbout", "REG_SZ", "https://kok.blue/edit")
	set("InstallLocation", "REG_SZ", dst)
	set("UninstallString", "REG_SZ", `"`+un+`" /uninstall`)
	set("NoModify", "REG_DWORD", "1")
	set("NoRepair", "REG_DWORD", "1")
}
