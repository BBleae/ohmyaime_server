package main

import (
	"log"
	"syscall"
	"time"
	"unsafe"
)

const (
	KEYEVENTF_KEYDOWN  = 0x0000
	KEYEVENTF_KEYUP    = 0x0002
	PROCESS_ALL_ACCESS = 0x1F0FFF
	WM_KEYDOWN         = 0x0100
	WM_KEYUP           = 0x0101
	TH32CS_SNAPPROCESS = 0x00000002
	MAX_PATH           = 260
)

type PROCESSENTRY32 struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [MAX_PATH]uint16
}

// IsProcessRunning checks if a process with the given name is running
func IsProcessRunning(processName string) bool {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First := kernel32.NewProc("Process32FirstW")
	procProcess32Next := kernel32.NewProc("Process32NextW")
	procCloseHandle := kernel32.NewProc("CloseHandle")

	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(
		uintptr(TH32CS_SNAPPROCESS),
		uintptr(0))
	if snapshot == uintptr(syscall.InvalidHandle) {
		return false
	}
	defer procCloseHandle.Call(snapshot)

	var entry PROCESSENTRY32
	entry.DwSize = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32First.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return false
	}

	for {
		name := syscall.UTF16ToString(entry.SzExeFile[:])
		if name == processName {
			return true
		}
		ret, _, _ := procProcess32Next.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}
	return false
}

var (
	user32        = syscall.NewLazyDLL("user32.dll")
	procSendInput = user32.NewProc("SendInput")
)

const (
	INPUT_MOUSE    = 0
	INPUT_KEYBOARD = 1
	INPUT_HARDWARE = 2
)

type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uint64
}

// INPUT structure
type INPUT struct {
	Type    uint32
	Ki      KEYBDINPUT
	padding uint64
}

func NewKeyboardInput(wVk, wScan uint16, dwFlags uint32, time uint32, dwExtraInfo uint64) INPUT {
	input := INPUT{Type: INPUT_KEYBOARD}
	input.Ki = KEYBDINPUT{
		WVk:         wVk,
		WScan:       wScan,
		DwFlags:     dwFlags,
		Time:        time,
		DwExtraInfo: dwExtraInfo,
	}
	return input
}

func getKeyCode(key string) uint16 {
	if key == "enter" {
		return 0x0D
	}
	if key == "a" {
		return 0x41
	}
	return 0
}

func sendGlobalEnterKey(key string, delay time.Duration) {
	keyCode := getKeyCode(key)
	input := NewKeyboardInput(uint16(keyCode), 0, KEYEVENTF_KEYDOWN, 0, 0)
	ret, _, err := procSendInput.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&input)),
		uintptr(unsafe.Sizeof(input)),
	)
	if ret != 1 {
		log.Printf("SendInput failed: %v", err)
	}

	time.Sleep(delay)

	input = NewKeyboardInput(uint16(keyCode), 0, KEYEVENTF_KEYUP, 0, 0)
	ret, _, err = procSendInput.Call(
		uintptr(1),
		uintptr(unsafe.Pointer(&input)),
		uintptr(unsafe.Sizeof(input)),
	)
	if ret != 1 {
		log.Printf("SendInput failed: %v", err)
	}
}
