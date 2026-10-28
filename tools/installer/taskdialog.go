package main

// Окна установщика — стандартный TaskDialog Windows (comctl32 v6): прогресс, кнопки, переход между страницами.
// Структуры в commctrl.h упакованы по 1 байту, поэтому собираем их вручную в байтовый буфер.

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

var (
	comctl32          = syscall.NewLazyDLL("comctl32.dll")
	procTaskDialogInd = comctl32.NewProc("TaskDialogIndirect")
	procLoadIconScale = comctl32.NewProc("LoadIconWithScaleDown")
	procSendMessage   = user32.NewProc("SendMessageW")
	kGetModule        = kernel32.NewProc("GetModuleHandleW")
	kUILang           = kernel32.NewProc("GetUserDefaultUILanguage")
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecute  = shell32.NewProc("ShellExecuteW")
)

const (
	tdfEnableHyperlinks    = 0x0001
	tdfUseHiconMain        = 0x0002
	tdfAllowCancel         = 0x0008
	tdfShowProgressBar     = 0x0200
	tdfShowMarquee         = 0x0400
	tdfCallbackTimer       = 0x0800
	tdfVerificationChecked = 0x0100
	tdfSizeToContent       = 0x01000000

	tdcbfCancel = 0x0008
	tdcbfClose  = 0x0020

	tdnCreated       = 0
	tdnNavigated     = 1
	tdnButtonClicked = 2
	tdnHyperlink     = 3
	tdnTimer         = 4
	tdnVerification  = 8

	wmUser                = 0x0400
	tdmNavigatePage       = wmUser + 101
	tdmClickButton        = wmUser + 102
	tdmSetMarquee         = wmUser + 103
	tdmSetProgressPos     = wmUser + 106
	tdmSetElementText     = wmUser + 108
	tdmEnableButton       = wmUser + 111
	tdmSetProgressMarquee = wmUser + 107
	tdeContent            = 0
	tdeMainInstruction    = 3

	idCancel = 2
	idClose  = 8
)

type btn struct {
	id   int
	text string
}

type page struct {
	main, content, footer, verify string
	buttons                       []btn
	common                        uint32
	progress, marquee, timer      bool
	verifyChecked                 bool
	defButton                     int
}

// всё, что передаём в Windows, держим живым, пока окно открыто
var keep [][]uint16
var keepBuf [][]byte

func wstr(s string) uintptr {
	if s == "" {
		return 0
	}
	p, _ := syscall.UTF16FromString(s)
	keep = append(keep, p)
	return uintptr(unsafe.Pointer(&p[0]))
}

var (
	hInst    uintptr
	hIcon    uintptr
	callback = syscall.NewCallback(tdCallback)
	onEvent  func(hwnd uintptr, msg uint32, wp, lp uintptr) uintptr
)

func tdCallback(hwnd, msg, wp, lp, ref uintptr) uintptr {
	if onEvent != nil {
		return onEvent(hwnd, uint32(msg), wp, lp)
	}
	return 0
}

func initUI() {
	hInst, _, _ = kGetModule.Call(0)
	// иконка из ресурсов exe (группа 2 — так её кладёт rsrc)
	procLoadIconScale.Call(hInst, 2, 32, 32, uintptr(unsafe.Pointer(&hIcon)))
}

// собрать TASKDIALOGCONFIG (160 байт на x64)
func config(title string, p page) uintptr {
	b := make([]byte, 160)
	put32 := func(off int, v uint32) { binary.LittleEndian.PutUint32(b[off:], v) }
	put64 := func(off int, v uintptr) { binary.LittleEndian.PutUint64(b[off:], uint64(v)) }
	flags := uint32(tdfAllowCancel | tdfSizeToContent | tdfEnableHyperlinks)
	if hIcon != 0 {
		flags |= tdfUseHiconMain
	}
	if p.progress {
		flags |= tdfShowProgressBar
	}
	if p.marquee {
		flags |= tdfShowMarquee
	}
	if p.timer {
		flags |= tdfCallbackTimer
	}
	if p.verifyChecked {
		flags |= tdfVerificationChecked
	}
	var btns uintptr
	if len(p.buttons) > 0 {
		bb := make([]byte, 12*len(p.buttons))
		for i, x := range p.buttons {
			binary.LittleEndian.PutUint32(bb[i*12:], uint32(x.id))
			binary.LittleEndian.PutUint64(bb[i*12+4:], uint64(wstr(x.text)))
		}
		keepBuf = append(keepBuf, bb)
		btns = uintptr(unsafe.Pointer(&bb[0]))
	}
	put32(0, 160)           // cbSize
	put64(4, 0)             // hwndParent
	put64(12, hInst)        // hInstance
	put32(20, flags)        // dwFlags
	put32(24, p.common)     // dwCommonButtons
	put64(28, wstr(title))  // pszWindowTitle
	put64(36, hIcon)        // hMainIcon
	put64(44, wstr(p.main)) // pszMainInstruction
	put64(52, wstr(p.content))
	put32(60, uint32(len(p.buttons)))
	put64(64, btns)
	put32(72, uint32(p.defButton))
	put32(76, 0) // cRadioButtons
	put64(80, 0) // pRadioButtons
	put32(88, 0) // nDefaultRadioButton
	put64(92, wstr(p.verify))
	put64(100, 0) // expanded info
	put64(108, 0)
	put64(116, 0)
	put64(124, 0) // footer icon
	put64(132, wstr(p.footer))
	put64(140, callback)
	put64(148, 0)
	put32(156, 0) // cxWidth — по содержимому
	keepBuf = append(keepBuf, b)
	return uintptr(unsafe.Pointer(&b[0]))
}

// показать окно; возвращает нажатую кнопку и галочку
func show(title string, p page) (int, bool) {
	var button, radio int32
	var checked int32
	procTaskDialogInd.Call(config(title, p), uintptr(unsafe.Pointer(&button)), uintptr(unsafe.Pointer(&radio)), uintptr(unsafe.Pointer(&checked)))
	return int(button), checked != 0
}

func navigate(hwnd uintptr, title string, p page) {
	procSendMessage.Call(hwnd, tdmNavigatePage, 0, config(title, p))
}
func setText(hwnd uintptr, el int, s string) {
	procSendMessage.Call(hwnd, tdmSetElementText, uintptr(el), wstr(s))
}
func setProgress(hwnd uintptr, pct int) {
	procSendMessage.Call(hwnd, tdmSetProgressPos, uintptr(pct), 0)
}
func enableButton(hwnd uintptr, id int, on bool) {
	v := uintptr(0)
	if on {
		v = 1
	}
	procSendMessage.Call(hwnd, tdmEnableButton, uintptr(id), v)
}
func marquee(hwnd uintptr) {
	procSendMessage.Call(hwnd, tdmSetProgressMarquee, 1, 0)
}
func openURL(u string) {
	procShellExecute.Call(0, wstr("open"), wstr(u), 0, 0, 1)
}
