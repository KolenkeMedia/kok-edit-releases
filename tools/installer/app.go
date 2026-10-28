package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ---------- состояние ----------
type job struct {
	mu        sync.Mutex
	phase     string // check | ready | download | install | done | error
	errKind   string // net | sig | hash | install
	errText   string
	label     string
	got, tot  int64
	cancelled bool
}

var (
	lang      = "en"
	st        = &job{}
	rel       release
	instVer   string // какая версия стоит ("" — не стоит)
	reinstall bool
)

const (
	bPrimary = 101
	bLang    = 102
	bRetry   = 103
	bDone    = 104
	bClose   = 105
)

func L() texts { return T[lang] }
func title() string {
	if channel == "test" {
		return L().Title + " — " + L().TestChannel
	}
	return L().Title
}

// ---------- версия, что уже стоит ----------
var verRe = regexp.MustCompile(`ExtensionBundleVersion="([^"]+)"`)

func installedVersion() string {
	b, err := os.ReadFile(filepath.Join(target(), "CSXS", "manifest.xml"))
	if err != nil {
		return ""
	}
	if m := verRe.FindSubmatch(b); m != nil {
		return string(m[1])
	}
	return ""
}

// ключ части — как в build-release.js: sha256 по путям и содержимому файлов
func partKey(files []string) string {
	sort.Strings(files)
	h := sha256.New()
	for _, f := range files {
		b, err := os.ReadFile(filepath.Join(target(), filepath.FromSlash(f)))
		if err != nil {
			return ""
		}
		h.Write([]byte(f + "\x00"))
		h.Write(b)
	}
	return hex.EncodeToString(h.Sum(nil))
}
func ffmpegFiles() []string {
	var out []string
	for _, f := range []string{"bin/FFMPEG_LICENSE.txt", "bin/ffmpeg.exe"} {
		if _, err := os.Stat(filepath.Join(target(), filepath.FromSlash(f))); err == nil {
			out = append(out, f)
		}
	}
	return out
}
func whisperFiles() []string {
	var out []string
	root := filepath.Join(target(), "bin", "whisper")
	filepath.Walk(root, func(p string, i os.FileInfo, err error) error {
		if err == nil && !i.IsDir() {
			r, _ := filepath.Rel(target(), p)
			out = append(out, filepath.ToSlash(r))
		}
		return nil
	})
	return out
}

// ---------- проверка версии ----------
func fetchRelease() {
	ptrRaw, err := get(fmt.Sprintf("https://raw.githubusercontent.com/%s/main/channels/%s.json?t=%d", repo, channel, time.Now().Unix()), nil)
	fail := func(kind string, e error) {
		st.mu.Lock()
		st.phase, st.errKind = "error", kind
		if e != nil {
			st.errText = e.Error()
		}
		st.mu.Unlock()
	}
	if err != nil {
		fail("net", err)
		return
	}
	var ptr struct{ Version, Release, Sig string }
	if json.Unmarshal(ptrRaw, &ptr) != nil {
		fail("net", nil)
		return
	}
	data, err := get(ptr.Release, nil)
	if err != nil {
		fail("net", err)
		return
	}
	sig, err := get(ptr.Sig, nil)
	if err != nil {
		fail("net", err)
		return
	}
	if !verify(data, sig) {
		fail("sig", nil)
		return
	}
	var r release
	if json.Unmarshal(data, &r) != nil || r.Components["core"].URL == "" {
		fail("sig", nil)
		return
	}
	st.mu.Lock()
	rel = r
	st.phase = "ready"
	st.mu.Unlock()
}

// ---------- установка ----------
func needFFmpeg() bool {
	c, ok := rel.Components["ffmpeg"]
	if !ok {
		return false
	}
	return reinstall || partKey(ffmpegFiles()) != c.Key
}
func downloadSize() int64 {
	n := rel.Components["core"].Size
	if needFFmpeg() {
		n += rel.Components["ffmpeg"].Size
	}
	return n
}

func doInstall() {
	set := func(f func()) { st.mu.Lock(); f(); st.mu.Unlock() }
	fail := func(kind, text string) { set(func() { st.phase, st.errKind, st.errText = "error", kind, text }) }
	parts := map[string][]byte{}
	names := []string{"core"}
	if needFFmpeg() {
		names = append(names, "ffmpeg")
	}
	for _, name := range names {
		c := rel.Components[name]
		lbl := L().Plugin
		if name == "ffmpeg" {
			lbl = L().FFmpeg
		}
		set(func() { st.label, st.got, st.tot = lbl, 0, c.Size })
		b, err := get(c.URL, func(got, total int64) {
			set(func() {
				st.got = got
				if total > 0 {
					st.tot = total
				}
			})
		})
		st.mu.Lock()
		cancelled := st.cancelled
		st.mu.Unlock()
		if cancelled {
			return
		}
		if err != nil {
			fail("net", err.Error())
			return
		}
		h := sha256.Sum256(b)
		if hex.EncodeToString(h[:]) != c.Sha256 {
			fail("hash", "")
			return
		}
		parts[name] = b
	}
	set(func() { st.phase = "install" })
	dst := target()
	// убираем старый код плагина; ffmpeg и whisper не трогаем, если они не обновляются
	var old struct {
		Core struct{ Files []string } `json:"core"`
	}
	if b, err := os.ReadFile(filepath.Join(dst, "components.json")); err == nil {
		json.Unmarshal(b, &old)
	}
	if len(old.Core.Files) > 0 {
		for _, f := range old.Core.Files {
			os.Remove(filepath.Join(dst, filepath.FromSlash(f)))
		}
	} else if entries, err := os.ReadDir(dst); err == nil {
		for _, e := range entries {
			if e.Name() == "bin" || strings.HasPrefix(e.Name(), "Удалить") || strings.HasPrefix(e.Name(), "Uninstall") {
				continue
			}
			if err := os.RemoveAll(filepath.Join(dst, e.Name())); err != nil {
				fail("install", err.Error())
				return
			}
		}
	}
	comps := map[string]interface{}{}
	coreFiles, err := unpack(parts["core"], dst)
	if err != nil {
		fail("install", err.Error())
		return
	}
	comps["core"] = map[string]interface{}{"key": rel.Components["core"].Key, "files": coreFiles}
	if b, ok := parts["ffmpeg"]; ok {
		if _, err := unpack(b, dst); err != nil {
			fail("install", err.Error())
			return
		}
	}
	if k := partKey(ffmpegFiles()); k != "" {
		comps["ffmpeg"] = map[string]interface{}{"key": k}
	}
	if w := whisperFiles(); len(w) > 0 {
		comps["whisper"] = map[string]interface{}{"key": partKey(w)}
	}
	cj, _ := json.MarshalIndent(comps, "", "  ")
	os.WriteFile(filepath.Join(dst, "components.json"), cj, 0o644)
	os.MkdirAll(workDir(), 0o755)
	os.WriteFile(filepath.Join(workDir(), "channel.txt"), []byte(channel), 0o644)
	debugMode()
	registerUninstall(dst, rel.Version)
	set(func() { st.phase = "done" })
}

// ---------- страницы ----------
func mb(n int64) string {
	return strings.Replace(fmt.Sprintf("%.1f", float64(n)/1048576), ".", map[bool]string{true: ",", false: "."}[lang != "en"], 1)
}

func pCheck() page {
	return page{main: L().Checking, content: L().CheckingSub, progress: true, marquee: true, timer: true, common: tdcbfCancel}
}
func pWelcome() page {
	l := L()
	p := page{common: tdcbfCancel, defButton: bPrimary}
	switch {
	case instVer == "":
		p.main = fmt.Sprintf(l.WelcomeMain, rel.Version)
		p.content = l.WelcomeText + "\n\n" + fmt.Sprintf(l.Size, mb(downloadSize())) + " " + l.NoAdmin
		p.buttons = []btn{{bPrimary, l.Install}}
	case instVer == rel.Version && !reinstall:
		p.main = fmt.Sprintf(l.SameMain, rel.Version)
		p.content = l.SameText
		p.buttons = []btn{{bClose, l.Close}, {bPrimary, l.Reinstall}}
		p.defButton = bClose
		p.common = 0
	default:
		p.main = fmt.Sprintf(l.UpdateMain, rel.Version)
		p.content = fmt.Sprintf(l.UpdateText, instVer) + "\n\n" + fmt.Sprintf(l.Size, mb(downloadSize()))
		p.buttons = []btn{{bPrimary, l.Update}}
	}
	p.buttons = append([]btn{{bLang, l.Lang}}, p.buttons...)
	p.footer = l.Requires
	return p
}
func pClosePremiere() page {
	return page{main: L().ClosePremiere, content: L().ClosePremiereText, buttons: []btn{{bRetry, L().Retry}}, common: tdcbfCancel, defButton: bRetry}
}
func pProgress() page {
	return page{main: L().Working, content: " ", progress: true, timer: true, common: tdcbfCancel}
}
func pDone() page {
	l := L()
	return page{main: l.DoneMain, content: l.DoneText + "\n\n" + l.DoneUpdates, verify: l.OpenGuide, verifyChecked: true, buttons: []btn{{bDone, l.Done}}, defButton: bDone}
}
func pError() page {
	l := L()
	msg := map[string]string{"net": l.ErrNet, "sig": l.ErrSig, "hash": l.ErrHash}[st.errKind]
	if st.errKind == "install" {
		msg = fmt.Sprintf(l.ErrInstall, st.errText)
	}
	p := page{main: l.ErrMain, content: msg, buttons: []btn{{bRetry, l.Retry}}, common: tdcbfClose, defButton: bRetry}
	if st.errKind == "sig" {
		p.buttons = nil
	}
	return p
}

// ---------- окно ----------
var current string
var lastShown string

func goPage(hwnd uintptr, name string) {
	current = name
	var p page
	switch name {
	case "check":
		p = pCheck()
	case "welcome":
		p = pWelcome()
	case "premiere":
		p = pClosePremiere()
	case "progress":
		p = pProgress()
	case "done":
		p = pDone()
	case "error":
		p = pError()
	}
	navigate(hwnd, title(), p)
}

func startInstall(hwnd uintptr) {
	if premiereRunning() {
		goPage(hwnd, "premiere")
		return
	}
	st.mu.Lock()
	st.phase, st.cancelled, st.got, st.tot, st.label = "download", false, 0, 0, ""
	st.mu.Unlock()
	goPage(hwnd, "progress")
	go doInstall()
}

func installerEvent(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
	switch msg {
	case tdnCreated, tdnNavigated:
		if current == "check" || current == "progress" {
			// бегущая полоса, пока размер не известен
			if current == "check" {
				marquee(hwnd)
			}
		}
	case tdnTimer:
		st.mu.Lock()
		phase, label, got, tot := st.phase, st.label, st.got, st.tot
		st.mu.Unlock()
		switch current {
		case "check":
			if phase == "ready" {
				goPage(hwnd, "welcome")
			} else if phase == "error" {
				goPage(hwnd, "error")
			}
		case "progress":
			switch phase {
			case "download":
				if label != "" {
					pct := 0
					if tot > 0 {
						pct = int(got * 100 / tot)
					}
					setProgress(hwnd, pct)
					txt := fmt.Sprintf(L().Downloading, label) + "\n" + mb(got) + " " + L().Of + " " + mb(tot) + " " + L().MB
					if txt != lastShown {
						setText(hwnd, tdeContent, txt)
						lastShown = txt
					}
				}
			case "install":
				enableButton(hwnd, idCancel, false)
				setProgress(hwnd, 100)
				if lastShown != L().Installing {
					setText(hwnd, tdeContent, L().Installing)
					lastShown = L().Installing
				}
			case "done":
				goPage(hwnd, "done")
			case "error":
				goPage(hwnd, "error")
			}
		}
	case tdnButtonClicked:
		id := int(wp)
		switch id {
		case bLang:
			for i, x := range langs {
				if x == lang {
					lang = langs[(i+1)%len(langs)]
					break
				}
			}
			goPage(hwnd, current)
			return 1
		case bPrimary:
			if instVer == rel.Version {
				reinstall = true
			}
			startInstall(hwnd)
			return 1
		case bRetry:
			if current == "premiere" {
				startInstall(hwnd)
			} else if rel.Version == "" {
				st.mu.Lock()
				st.phase = "check"
				st.mu.Unlock()
				goPage(hwnd, "check")
				go fetchRelease()
			} else {
				startInstall(hwnd)
			}
			return 1
		case idCancel:
			st.mu.Lock()
			if st.phase == "install" {
				st.mu.Unlock()
				return 1 // файлы уже пишутся — не прерываем
			}
			st.cancelled = true
			st.mu.Unlock()
		}
	}
	return 0
}

// ---------- удаление ----------
func uninstallUI() {
	onEvent = func(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr {
		if msg == tdnButtonClicked && int(wp) == bLang {
			for i, x := range langs {
				if x == lang {
					lang = langs[(i+1)%len(langs)]
					break
				}
			}
			navigate(hwnd, L().Name, unPage())
			return 1
		}
		return 0
	}
	b, _ := show(L().Name, unPage())
	if b != bPrimary {
		return
	}
	for premiereRunning() {
		onEvent = nil
		if r, _ := show(L().Name, pClosePremiere()); r != bRetry {
			return
		}
	}
	dst := target()
	run("reg", "delete", uninKey, "/f")
	entries, _ := os.ReadDir(dst)
	for _, e := range entries {
		if strings.HasSuffix(strings.ToLower(e.Name()), ".exe") {
			continue
		}
		os.RemoveAll(filepath.Join(dst, e.Name()))
	}
	onEvent = nil
	show(L().Name, page{main: L().UnDone, buttons: []btn{{bDone, L().Done}}})
	cmd := exec.Command("cmd", "/C", "ping 127.0.0.1 -n 3 > nul & rmdir /S /Q \""+dst+"\"")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Start()
}
func unPage() page {
	return page{main: L().UnMain, content: L().UnText, buttons: []btn{{bLang, L().Lang}, {bPrimary, L().UnBtn}}, common: tdcbfCancel, defButton: idCancel}
}

func main() {
	runtime.LockOSThread()
	initUI()
	id, _, _ := kUILang.Call()
	lang = detectLang(uint16(id))
	for _, a := range os.Args[1:] {
		if strings.EqualFold(a, "/uninstall") {
			uninstallUI()
			return
		}
	}
	instVer = installedVersion()
	st.phase = "check"
	current = "check"
	go fetchRelease()
	onEvent = installerEvent
	_, guide := show(title(), pCheck())
	st.mu.Lock()
	done := st.phase == "done"
	st.mu.Unlock()
	if done && guide {
		openURL("https://kok.blue/edit")
	}
}
