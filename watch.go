// Copyright 2013 bee authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/howeyc/fsnotify"
)

var (
	cmd          *exec.Cmd
	state        sync.Mutex
	eventTime    = make(map[string]int64)
	scheduleTime time.Time
	watchExts    = []string{".go"}
)

// NewWatcher ...
func NewWatcher(paths []string, extensions []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		ColorLog("[ERRO] Fail to create new Watcher[ %s ]\n", err)
		os.Exit(2)
	}
	if extensions != nil {
		watchExts = extensions
	}

	go func() {
		for {
			select {
			case e := <-watcher.Event:
				isbuild := true

				// Skip TMP files for Sublime Text.
				if checkTMPFile(e.Name) {
					continue
				}
				if !checkExtension(e.Name, watchExts) {
					continue
				}

				mt := getFileModTime(e.Name)
				if t := eventTime[e.Name]; mt == t {
					ColorLog("[SKIP] # %s #\n", e.String())
					isbuild = false
				}

				eventTime[e.Name] = mt

				if isbuild {
					ColorLog("[EVEN] %s\n", e)
					go func() {
						// Wait 1s before autobuild util there is no file change.
						scheduleTime = time.Now().Add(1 * time.Second)
						for {
							time.Sleep(scheduleTime.Sub(time.Now()))
							if time.Now().After(scheduleTime) {
								break
							}
							return
						}

						Autobuild()
					}()
				}
			case err := <-watcher.Error:
				ColorLog("[WARN] %s\n", err.Error()) // No need to exit here
			}
		}
	}()

	ColorLog("[INFO] Initializing watcher...\n")
	for _, path := range paths {
		ColorLog("[TRAC] Directory( %s )\n", path)
		err = watcher.Watch(path)
		if err != nil {
			ColorLog("[ERRO] Fail to watch directory[ %s ]\n", err)
			os.Exit(2)
		}
	}

}

// getFileModTime retuens unix timestamp of `os.File.ModTime` by given path.
func getFileModTime(path string) int64 {
	path = strings.Replace(path, "\\", "/", -1)
	f, err := os.Open(path)
	if err != nil {
		ColorLog("[ERRO] Fail to open file[ %s ]\n", err)
		return time.Now().Unix()
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		ColorLog("[ERRO] Fail to get file information[ %s ]\n", err)
		return time.Now().Unix()
	}

	return fi.ModTime().Unix()
}

// Autobuild ...
func Autobuild() {
	state.Lock()
	defer state.Unlock()

	ColorLog("[INFO] Start building...\n")
	path, _ := os.Getwd()
	os.Chdir(path)

	cmdName := "go"

	appName := appname
	if runtime.GOOS == "windows" {
		appName += ".exe"
	}

	args := []string{"build"}
	args = append(args, "-o", appName)
	if buildTags != "" {
		args = append(args, "-tags", buildTags)
	}

	if mainPath != "" {
		args = append(args, mainPath)
	}

	bcmd := exec.Command(cmdName, args...)
	bcmd.Env = append(os.Environ(), "GOGC=off")
	bcmd.Stdout = os.Stdout
	bcmd.Stderr = os.Stderr
	err := bcmd.Run()
	if err != nil {
		ColorLog("[ERRO] ============== Build failed ===================\n")
		return
	}
	ColorLog("[SUCC] Build was successful\n")
	Restart(appname)
}

// Kill ...
func Kill() {
	defer func() {
		if e := recover(); e != nil {
			fmt.Println("Kill.recover -> ", e)
		}
	}()
	if cmd != nil && cmd.Process != nil {
		err := cmd.Process.Kill()
		if err != nil {
			fmt.Println("Kill -> ", err)
		}
	}
}

// Restart ...
func Restart(appname string) {
	Debugf("kill running process")
	Kill()
	go Start(appname)
}

// Start ...
func Start(appname string) {
	ColorLog("[INFO] Restarting %s ...\n", appname)
	if !strings.Contains(appname, "./") {
		appname = "./" + appname
	}

	cmd = exec.Command(appname)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Args = []string{appname}
	cmd.Env = os.Environ()

	go cmd.Run()
	ColorLog("[INFO] %s is running...\n", appname)
}

// checkTMPFile returns true if the event was for TMP files.
func checkTMPFile(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".tmp")
}
