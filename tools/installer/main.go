// Лёгкий установщик kök Edit для Premiere Pro (Windows). Сам скачивает свежую версию с GitHub,
// проверяет подпись kök и ставит без прав администратора:
// плагин → %APPDATA%\Adobe\CEP\extensions\kok-edit, разрешение расширений → HKCU\Software\Adobe\CSXS.*
package main

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var channel = "main" // при сборке тестового установщика: -X main.channel=test

const (
	title     = "kök Edit"
	extDir    = "kok-edit"
	repo      = "KolenkeMedia/kok-edit-releases"
	publicKey = "MCowBQYDK2VwAyEAjhdCyWn6yr5lCkRTLjODTJYlSp1+iVa+4Oxg9QipUeg="
	uninKey   = `HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\kokEdit`
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	msgBox   = user32.NewProc("MessageBoxW")
	setCP    = kernel32.NewProc("SetConsoleOutputCP")
	setTitle = kernel32.NewProc("SetConsoleTitleW")
)

const (
	mbOK, mbOKCancel, mbRetryCancel             = 0x0, 0x1, 0x5
	mbIconInfo, mbIconWarn, mbIconError, mbIconQ = 0x40, 0x30, 0x10, 0x20
	idOK, idRetry                                = 1, 4
)

func box(text string, flags uintptr) int {
	t, _ := syscall.UTF16PtrFromString(text)
	c, _ := syscall.UTF16PtrFromString(title)
	r, _, _ := msgBox.Call(0, uintptr(unsafe.Pointer(t)), uintptr(unsafe.Pointer(c)), flags)
	return int(r)
}
func fail(text string) {
	box(text, mbOK|mbIconError)
	os.Exit(1)
}

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
func waitPremiereClosed() bool {
	for premiereRunning() {
		if box("Закройте Adobe Premiere Pro и нажмите «Повтор».\n\nПроекты не пострадают — просто сохраните их перед закрытием.", mbRetryCancel|mbIconWarn) != idRetry {
			return false
		}
	}
	return true
}
func target() string { return filepath.Join(os.Getenv("APPDATA"), "Adobe", "CEP", "extensions", extDir) }
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
	for v := 9; v <= 12; v++ {
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
	set("DisplayName", "REG_SZ", "kök Edit для Premiere Pro")
	set("DisplayVersion", "REG_SZ", version)
	set("Publisher", "REG_SZ", "kök")
	set("URLInfoAbout", "REG_SZ", "https://kok.blue/edit")
	set("InstallLocation", "REG_SZ", dst)
	set("UninstallString", "REG_SZ", `"`+un+`" /uninstall`)
	set("NoModify", "REG_DWORD", "1")
	set("NoRepair", "REG_DWORD", "1")
}

func uninstall() {
	if box("Удалить kök Edit из Premiere Pro?\n\nНастройки, расшифровки и логи в папке «kokEdit» вашей папки пользователя останутся.", mbOKCancel|mbIconQ) != idOK {
		return
	}
	if !waitPremiereClosed() {
		return
	}
	dst := target()
	run("reg", "delete", uninKey, "/f")
	entries, _ := os.ReadDir(dst)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "Удалить") {
			continue
		}
		os.RemoveAll(filepath.Join(dst, e.Name()))
	}
	box("kök Edit удалён.", mbOK|mbIconInfo)
	cmd := exec.Command("cmd", "/C", "ping 127.0.0.1 -n 3 > nul & rmdir /S /Q \""+dst+"\"")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}

func main() {
	setCP.Call(65001)
	t, _ := syscall.UTF16PtrFromString("Установка kök Edit")
	setTitle.Call(uintptr(unsafe.Pointer(t)))
	for _, a := range os.Args[1:] {
		if strings.EqualFold(a, "/uninstall") {
			uninstall()
			return
		}
	}
	fmt.Println("kök Edit для Premiere Pro — установка" + map[bool]string{true: " (тестовая среда)", false: ""}[channel == "test"])
	fmt.Println("Узнаю последнюю версию…")
	ptrRaw, err := get(fmt.Sprintf("https://raw.githubusercontent.com/%s/main/channels/%s.json?t=%d", repo, channel, time.Now().Unix()), nil)
	if err != nil {
		fail("Не удалось узнать версию: " + err.Error())
	}
	var ptr struct{ Version, Release, Sig string }
	if json.Unmarshal(ptrRaw, &ptr) != nil {
		fail("Ответ сервера не понятен — попробуйте позже.")
	}
	data, err := get(ptr.Release, nil)
	if err != nil {
		fail("Не удалось скачать описание версии: " + err.Error())
	}
	sig, err := get(ptr.Sig, nil)
	if err != nil || !verify(data, sig) {
		fail("Подпись версии не сходится — установка остановлена. Скачайте установщик с kok.blue/edit.")
	}
	var rel release
	if json.Unmarshal(data, &rel) != nil || rel.Components["core"].URL == "" {
		fail("Описание версии повреждено.")
	}
	dst := target()
	_, statErr := os.Stat(filepath.Join(dst, "CSXS", "manifest.xml"))
	update := statErr == nil
	q := "Установить kök Edit " + rel.Version + " для Adobe Premiere Pro?\n\nСкачается около 55 МБ."
	if update {
		q = "kök Edit уже установлен. Обновить до версии " + rel.Version + "?\n\nНастройки и вход в аккаунт сохранятся."
	}
	if box(q, mbOKCancel|mbIconQ) != idOK {
		return
	}
	if !waitPremiereClosed() {
		return
	}
	// сначала всё скачать и проверить, потом ставить
	parts := map[string][]byte{}
	for _, name := range []string{"core", "ffmpeg"} {
		c, ok := rel.Components[name]
		if !ok {
			continue
		}
		label := map[string]string{"core": "плагин", "ffmpeg": "ffmpeg (обработка звука и видео)"}[name]
		last := time.Now()
		b, err := get(c.URL, func(got, total int64) {
			if time.Since(last) > 300*time.Millisecond || got == total {
				last = time.Now()
				if total <= 0 {
					total = c.Size
				}
				fmt.Printf("\rСкачиваю %s: %.1f из %.1f МБ   ", label, float64(got)/1048576, float64(total)/1048576)
			}
		})
		fmt.Println()
		if err != nil {
			fail("Не удалось скачать " + label + ": " + err.Error())
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != c.Sha256 {
			fail("Файл " + label + " скачался с ошибкой — запустите установку ещё раз.")
		}
		parts[name] = b
	}
	fmt.Println("Устанавливаю…")
	if update {
		if err := os.RemoveAll(dst); err != nil {
			fail("Не удалось убрать старую версию: " + err.Error() + "\n\nЗакройте Premiere и папку плагина в Проводнике и запустите установку снова.")
		}
	}
	comps := map[string]interface{}{}
	for name, b := range parts {
		files, err := unpack(b, dst)
		if err != nil {
			fail("Ошибка установки: " + err.Error())
		}
		if name == "core" {
			comps[name] = map[string]interface{}{"key": rel.Components[name].Key, "files": files}
		} else {
			comps[name] = map[string]interface{}{"key": rel.Components[name].Key}
		}
	}
	cj, _ := json.MarshalIndent(comps, "", "  ")
	os.WriteFile(filepath.Join(dst, "components.json"), cj, 0o644)
	os.MkdirAll(workDir(), 0o755)
	os.WriteFile(filepath.Join(workDir(), "channel.txt"), []byte(channel), 0o644)
	debugMode()
	registerUninstall(dst, rel.Version)
	box("Готово! kök Edit "+rel.Version+" установлен.\n\nОткройте Premiere Pro → меню «Окно» → «Расширения» → kök Edit.\nДальше обновления приходят прямо в плагине.", mbOK|mbIconInfo)
}
