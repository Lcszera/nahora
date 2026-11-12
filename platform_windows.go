//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

func init() { runtime.LockOSThread() } // laço de mensagens precisa ficar na thread principal

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	pRegisterClassExW         = user32.NewProc("RegisterClassExW")
	pCreateWindowExW          = user32.NewProc("CreateWindowExW")
	pDefWindowProcW           = user32.NewProc("DefWindowProcW")
	pDestroyWindow            = user32.NewProc("DestroyWindow")
	pGetMessageW              = user32.NewProc("GetMessageW")
	pTranslateMessage         = user32.NewProc("TranslateMessage")
	pDispatchMessageW         = user32.NewProc("DispatchMessageW")
	pPostQuitMessage          = user32.NewProc("PostQuitMessage")
	pPostMessageW             = user32.NewProc("PostMessageW")
	pLoadImageW               = user32.NewProc("LoadImageW")
	pGetSystemMetrics         = user32.NewProc("GetSystemMetrics")
	pRegisterWindowMessageW   = user32.NewProc("RegisterWindowMessageW")
	pCreatePopupMenu          = user32.NewProc("CreatePopupMenu")
	pAppendMenuW              = user32.NewProc("AppendMenuW")
	pSetMenuDefaultItem       = user32.NewProc("SetMenuDefaultItem")
	pTrackPopupMenu           = user32.NewProc("TrackPopupMenu")
	pDestroyMenu              = user32.NewProc("DestroyMenu")
	pGetCursorPos             = user32.NewProc("GetCursorPos")
	pSetForegroundWindow      = user32.NewProc("SetForegroundWindow")
	pShowWindow               = user32.NewProc("ShowWindow")
	pIsIconic                 = user32.NewProc("IsIconic")
	pIsWindowVisible          = user32.NewProc("IsWindowVisible")
	pEnumWindows              = user32.NewProc("EnumWindows")
	pGetClassNameW            = user32.NewProc("GetClassNameW")
	pGetWindowTextW           = user32.NewProc("GetWindowTextW")
	pMessageBoxW              = user32.NewProc("MessageBoxW")
	pShellNotifyIconW         = shell32.NewProc("Shell_NotifyIconW")
	pShellExecuteW            = shell32.NewProc("ShellExecuteW")
	pGetModuleHandleW         = kernel32.NewProc("GetModuleHandleW")
	pGetCurrentProcess        = kernel32.NewProc("GetCurrentProcess")
	pSetProcessWorkingSetSize = kernel32.NewProc("SetProcessWorkingSetSize")
	pRegCreateKeyExW          = advapi32.NewProc("RegCreateKeyExW")
	pRegSetValueExW           = advapi32.NewProc("RegSetValueExW")
	pRegDeleteValueW          = advapi32.NewProc("RegDeleteValueW")
	pRegCloseKey              = advapi32.NewProc("RegCloseKey")
)

const (
	wmDestroy     = 0x0002
	wmNull        = 0x0000
	wmClose       = 0x0010
	wmLButtonUp   = 0x0202
	wmRButtonUp   = 0x0205
	wmApp         = 0x8000
	wmTray        = wmApp + 1
	ninBalloonClk = 0x0405

	nimAdd    = 0
	nimModify = 1
	nimDelete = 2

	nifMessage = 0x01
	nifIcon    = 0x02
	nifTip     = 0x04
	nifInfo    = 0x10
	nifShowTip = 0x80

	niifUser      = 0x04
	niifLargeIcon = 0x20
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

type notifyIconData struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	UVersion         uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         guid
	HBalloonIcon     uintptr
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type point struct{ X, Y int32 }

type winMsg struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
	Private uint32
}

var (
	hwndMain       uintptr
	hInst          uintptr
	iconSmall      uintptr
	iconLarge      uintptr
	taskbarCreated uintptr
	trayMu         sync.Mutex
)

func u16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }

func copyU16(dst []uint16, s string) {
	src, _ := syscall.UTF16FromString(s)
	if len(src) > len(dst) {
		src = src[:len(dst)]
		src[len(src)-1] = 0
	}
	copy(dst, src)
}

func baseNID() notifyIconData {
	var n notifyIconData
	n.CbSize = uint32(unsafe.Sizeof(n))
	n.HWnd = hwndMain
	n.UID = 1
	return n
}

func addTrayIcon() {
	n := baseNID()
	n.UFlags = nifMessage | nifIcon | nifTip | nifShowTip
	n.UCallbackMessage = wmTray
	n.HIcon = iconSmall
	copyU16(n.SzTip[:], "naHora — rotina e tarefas")
	trayMu.Lock()
	pShellNotifyIconW.Call(nimAdd, uintptr(unsafe.Pointer(&n)))
	trayMu.Unlock()
}

func removeTrayIcon() {
	n := baseNID()
	trayMu.Lock()
	pShellNotifyIconW.Call(nimDelete, uintptr(unsafe.Pointer(&n)))
	trayMu.Unlock()
}

func notify(title, msg string, silent bool) {
	if hwndMain == 0 {
		return
	}
	n := baseNID()
	n.UFlags = nifInfo
	copyU16(n.SzInfoTitle[:], title)
	copyU16(n.SzInfo[:], msg)
	n.DwInfoFlags = niifUser | niifLargeIcon
	if silent {
		n.DwInfoFlags |= 0x10 // NIIF_NOSOUND: o naHora toca o próprio som
	}
	n.HBalloonIcon = iconLarge
	trayMu.Lock()
	pShellNotifyIconW.Call(nimModify, uintptr(unsafe.Pointer(&n)))
	trayMu.Unlock()
}

func loadIcon(size uintptr) uintptr {
	h, _, _ := pLoadImageW.Call(hInst, 1, 1 /*IMAGE_ICON*/, size, size, 0)
	return h
}

func showMenu(hwnd uintptr) {
	m, _, _ := pCreatePopupMenu.Call()
	pAppendMenuW.Call(m, 0, 1, uintptr(unsafe.Pointer(u16("Abrir naHora"))))
	var flags uintptr
	if autoStartOn() {
		flags = 0x8 // MF_CHECKED
	}
	pAppendMenuW.Call(m, flags, 2, uintptr(unsafe.Pointer(u16("Iniciar com o Windows"))))
	pAppendMenuW.Call(m, 0, 4, uintptr(unsafe.Pointer(u16("Testar notificação"))))
	pAppendMenuW.Call(m, 0, 5, uintptr(unsafe.Pointer(u16("Procurar atualização"))))
	pAppendMenuW.Call(m, 0x800 /*MF_SEPARATOR*/, 0, 0)
	pAppendMenuW.Call(m, 0, 3, uintptr(unsafe.Pointer(u16("Sair"))))
	pSetMenuDefaultItem.Call(m, 1, 0)
	var pt point
	pGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	pSetForegroundWindow.Call(hwnd)
	cmd, _, _ := pTrackPopupMenu.Call(m, 0x100|0x80|0x2, /*RETURNCMD|NONOTIFY|RIGHTBUTTON*/
		uintptr(pt.X), uintptr(pt.Y), 0, hwnd, 0)
	pPostMessageW.Call(hwnd, wmNull, 0, 0)
	pDestroyMenu.Call(m)
	switch cmd {
	case 1:
		openUI(uiURL)
	case 2:
		toggleAutoStart()
	case 4:
		notesQ <- Note{"⚡ naHora funcionando", "É assim que os avisos vão aparecer."}
	case 5:
		go func() {
			checkUpdate()
			i := updateSnapshot()
			if i.Status != "ready" && i.Status != "applying" {
				notesQ <- Note{"naHora " + appVersion, i.Message}
			}
		}()
	case 3:
		if w := findAppWindow(); w != 0 {
			pPostMessageW.Call(w, wmClose, 0, 0)
		}
		removeTrayIcon()
		pDestroyWindow.Call(hwnd)
	}
}

func wndProc(hwnd, msg, wp, lp uintptr) uintptr {
	switch {
	case msg == wmTray:
		switch lp & 0xFFFF {
		case wmLButtonUp, ninBalloonClk:
			go openUI(uiURL)
		case wmRButtonUp:
			showMenu(hwnd)
		}
		return 0
	case msg == taskbarCreated && taskbarCreated != 0:
		addTrayIcon() // o Explorer reiniciou: recoloca o ícone
		return 0
	case msg == wmDestroy:
		pPostQuitMessage.Call(0)
		return 0
	}
	r, _, _ := pDefWindowProcW.Call(hwnd, msg, wp, lp)
	return r
}

func platformRun() {
	hInst, _, _ = pGetModuleHandleW.Call(0)
	sm, _, _ := pGetSystemMetrics.Call(49) // SM_CXSMICON
	if sm == 0 {
		sm = 16
	}
	lg, _, _ := pGetSystemMetrics.Call(11) // SM_CXICON
	if lg == 0 {
		lg = 32
	}
	iconSmall = loadIcon(sm)
	iconLarge = loadIcon(lg * 2)
	if iconLarge == 0 {
		iconLarge = loadIcon(lg)
	}
	taskbarCreated, _, _ = pRegisterWindowMessageW.Call(uintptr(unsafe.Pointer(u16("TaskbarCreated"))))

	cls := u16("naHoraTray")
	wc := wndClassEx{
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInst,
		HIcon:         iconLarge,
		HIconSm:       iconSmall,
		LpszClassName: cls,
	}
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	pRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))
	hwndMain, _, _ = pCreateWindowExW.Call(0, uintptr(unsafe.Pointer(cls)),
		uintptr(unsafe.Pointer(u16("naHora"))), 0, 0, 0, 0, 0, 0, 0, hInst, 0)
	if hwndMain == 0 {
		fatal("Não consegui criar o ícone da bandeja.")
		return
	}
	addTrayIcon()
	trimMemory()

	var m winMsg
	for {
		r, _, _ := pGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		pTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		pDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

// ---------- janela da interface ----------

var (
	enumCB     = syscall.NewCallback(enumProc)
	enumFound  uintptr
	enumMu     sync.Mutex
	openMu     sync.Mutex
	lastLaunch time.Time
)

func enumProc(hwnd, _ uintptr) uintptr {
	if v, _, _ := pIsWindowVisible.Call(hwnd); v == 0 {
		return 1
	}
	var cls [64]uint16
	pGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&cls[0])), 64)
	if syscall.UTF16ToString(cls[:]) != "Chrome_WidgetWin_1" {
		return 1
	}
	var title [256]uint16
	pGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&title[0])), 256)
	t := syscall.UTF16ToString(title[:])
	if t == "naHora" || strings.HasPrefix(t, "naHora - ") || strings.HasPrefix(t, "naHora — ") {
		enumFound = hwnd
		return 0
	}
	return 1
}

func findAppWindow() uintptr {
	enumMu.Lock()
	defer enumMu.Unlock()
	enumFound = 0
	pEnumWindows.Call(enumCB, 0)
	return enumFound
}

func findBrowser() string {
	pf, pf86, local := os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)"), os.Getenv("LOCALAPPDATA")
	for _, p := range []string{
		filepath.Join(pf86, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pf, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(local, `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(pf, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(pf86, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(local, `Google\Chrome\Application\chrome.exe`),
		filepath.Join(pf, `BraveSoftware\Brave-Browser\Application\brave.exe`),
	} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func openUI(url string) {
	if w := findAppWindow(); w != 0 {
		if ic, _, _ := pIsIconic.Call(w); ic != 0 {
			pShowWindow.Call(w, 9) // SW_RESTORE
		}
		pSetForegroundWindow.Call(w)
		return
	}
	openMu.Lock()
	defer openMu.Unlock()
	if time.Since(lastLaunch) < 4*time.Second {
		return // evita abrir duas janelas num clique duplo
	}
	lastLaunch = time.Now()

	if b := findBrowser(); b != "" {
		base := os.Getenv("LOCALAPPDATA")
		if base == "" {
			base = dataDir()
		}
		profile := filepath.Join(base, appName, "janela")
		cmd := exec.Command(b,
			"--app="+url,
			"--user-data-dir="+profile,
			"--window-size=1180,820",
			"--no-first-run",
			"--no-default-browser-check",
			"--disable-sync",
			"--disable-extensions",
			"--disable-background-mode",
			"--disable-features=Translate,msEdgeSidebarV2,msUndersideButton,msHubApps",
		)
		if cmd.Start() == nil {
			cmd.Process.Release()
			return
		}
	}
	pShellExecuteW.Call(0, uintptr(unsafe.Pointer(u16("open"))), uintptr(unsafe.Pointer(u16(url))), 0, 0, 1)
}

// ---------- iniciar com o Windows ----------

const runKey = `Software\Microsoft\Windows\CurrentVersion\Run`

func setAutoStart(on bool) error {
	var hk uintptr
	r, _, _ := pRegCreateKeyExW.Call(0x80000001 /*HKCU*/, uintptr(unsafe.Pointer(u16(runKey))), 0, 0, 0,
		0x0002 /*KEY_SET_VALUE*/, 0, uintptr(unsafe.Pointer(&hk)), 0)
	if r != 0 {
		return syscall.Errno(r)
	}
	defer pRegCloseKey.Call(hk)
	if !on {
		r, _, _ = pRegDeleteValueW.Call(hk, uintptr(unsafe.Pointer(u16(appName))))
		if r != 0 && r != 2 {
			return syscall.Errno(r)
		}
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	val, _ := syscall.UTF16FromString(`"` + exe + `" --tray`)
	r, _, _ = pRegSetValueExW.Call(hk, uintptr(unsafe.Pointer(u16(appName))), 0, 1, /*REG_SZ*/
		uintptr(unsafe.Pointer(&val[0])), uintptr(len(val)*2))
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

// trimMemory devolve ao Windows a memória que não está em uso.
func trimMemory() {
	debug.FreeOSMemory()
	h, _, _ := pGetCurrentProcess.Call()
	pSetProcessWorkingSetSize.Call(h, ^uintptr(0), ^uintptr(0))
}

func fatal(msg string) {
	pMessageBoxW.Call(0, uintptr(unsafe.Pointer(u16(msg))), uintptr(unsafe.Pointer(u16("naHora"))), 0x10)
}

// ---------- som personalizado (MCI: toca MP3, WAV e WMA sem bibliotecas extras) ----------

var (
	winmm           = syscall.NewLazyDLL("winmm.dll")
	pMciSendStringW = winmm.NewProc("mciSendStringW")
	soundMu         sync.Mutex
)

func mci(cmd string) (string, uintptr) {
	var buf [128]uint16
	r, _, _ := pMciSendStringW.Call(uintptr(unsafe.Pointer(u16(cmd))), uintptr(unsafe.Pointer(&buf[0])), 128, 0)
	return syscall.UTF16ToString(buf[:]), r
}

func playSound(path string, vol int) {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		soundMu.Lock()
		defer soundMu.Unlock()
		mci("close nhsom")
		if _, r := mci(`open "` + path + `" type mpegvideo alias nhsom`); r != 0 {
			if _, r = mci(`open "` + path + `" alias nhsom`); r != 0 {
				pMessageBeep.Call(0x40)
				return
			}
		}
		mci(fmt.Sprintf("setaudio nhsom volume to %d", vol*10))
		mci("play nhsom")
		// toca no máximo 15 s (se o áudio for uma música inteira, corta)
		for i := 0; i < 75; i++ {
			time.Sleep(200 * time.Millisecond)
			if mode, _ := mci("status nhsom mode"); mode != "playing" {
				break
			}
		}
		mci("stop nhsom")
		mci("close nhsom")
		trimMemory()
	}()
}

var pMessageBeep = user32.NewProc("MessageBeep")

func platformBeforeExit() {
	removeTrayIcon()
}

func exeHeaderOK(h []byte) bool { return len(h) == 2 && h[0] == 'M' && h[1] == 'Z' }
