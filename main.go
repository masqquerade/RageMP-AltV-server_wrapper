package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

const (
	ModAlt = 1 << iota
	ModCtrl
	ModShift
	ModWin
)

type MSG struct {
	HWND   uintptr
	UINT   uintptr
	WPARAM int16
	LPARAM int64
	DWORD  int32
	POINT  struct{ X, Y int64 }
}

type Config struct {
	LaunchedBefore bool   `json:"launched_before"`
	Platform       string `json:"platform"`
}

func main() {
	scmd := startServer()

	user32 := syscall.MustLoadDLL("user32")
	reghotkey := user32.MustFindProc("RegisterHotKey")

	reghotkey.Call(0, 1, uintptr(ModCtrl+ModAlt), 'P')
	peekmsg := user32.MustFindProc("PeekMessageW")

	for {
		var msg = &MSG{}
		peekmsg.Call(uintptr(unsafe.Pointer(msg)), 0, 0, 0, 1)

		if msg.WPARAM == 1 {
			err := scmd.Process.Kill()
			if err != nil {
				fmt.Println(err)
			}
			clearCmd()
			scmd = startServer()
		}
	}
}

func startServer() exec.Cmd {
	var platform string

	createCfg(&platform)

	startServerCmd := exec.Command(platform)
	startServerCmd.Stdout = os.Stdout
	startServerCmd.Stderr = os.Stderr

	startServerCmd.Start()

	return *startServerCmd
}

func createCfg(pl *string) {
	if _, err := os.Stat("cfg.json"); errors.Is(err, os.ErrNotExist) {
		os.WriteFile("cfg.json", []byte(`{"launched_before": false, "platform": ""}`), 0777)
	}

	f, err := os.OpenFile("cfg.json", os.O_RDWR, 0777)

	if err != nil {
		fmt.Println(err)
	}

	b, _ := ioutil.ReadAll(f)
	f.Close()

	var data Config
	json.Unmarshal(b, &data)

	if !data.LaunchedBefore {
		fmt.Println("Enter your platform (RAGEMP / AltV)")
		fmt.Scan(pl)

		if *pl == "RAGEMP" {
			*pl = "ragemp-server.exe"
		} else if *pl == "AltV" {
			*pl = "altv-server.exe"
		}

		cfg := Config{
			LaunchedBefore: true,
			Platform:       *pl,
		}

		err := os.Remove("cfg.json")
		if err != nil {
			fmt.Println(err)
		}

		b, _ := json.Marshal(cfg)

		os.WriteFile("cfg.json", b, 0777)
	} else {
		*pl = data.Platform
	}
}

func clearCmd() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
