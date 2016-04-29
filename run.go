package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	path "path/filepath"
	"runtime"
	"strings"
)

var cmdRun = &Command{
	UsageLine: "run [-t=.go -ext=.ini] [-e=Godeps -e=folderToExclude] [-tags=goBuildTags]",
	Short:     "run the app and start a Web server for development",
	Long: `
Run command will supervise the file system of the go project using inotify,
it will recompile and restart the app after any modifications.
`,
}

// The extension list of the paths excluded from watching
var extensions strFlags

// The flags list of the paths excluded from watching
var excludedPaths strFlags

// Pass through to -tags arg of "go build"
var buildTags string

func init() {
	cmdRun.Run = runApp
	cmdRun.Flag.Var(&extensions, "t", "extension, default .go")
	cmdRun.Flag.Var(&excludedPaths, "e", "Excluded paths[].")
	cmdRun.Flag.StringVar(&buildTags, "tags", "", "Build tags (https://golang.org/pkg/go/build/)")
	extensions = append(extensions, ".go")
}

var appname string

func runApp(cmd *Command, args []string) int {
	fmt.Println("bat   :" + version)
	goversion, err := exec.Command("go", "version").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Go    :" + strings.TrimSpace(string(goversion)))
	fmt.Printf("Ext   :%v\n", extensions)
	fmt.Println("")

	exit := make(chan bool)
	crupath, _ := os.Getwd()

	appname = path.Base(crupath)
	ColorLog("[INFO] Uses '%s' as 'appname'\n", appname)
	if strings.HasSuffix(appname, ".go") && isExist(path.Join(crupath, appname)) {
		ColorLog("[WARN] The appname has conflic with crupath's file, do you want to build appname as %s\n", appname)
		ColorLog("[INFO] Do you want to overwrite it? [yes|no]]  ")
	}
	Debugf("current path:%s\n", crupath)

	var paths []string

	readAppDirectories(crupath, &paths)

	// Because monitor files has some issues, we watch current directory
	// and ignore non-go files.
	gps := GetGOPATHs()
	if len(gps) == 0 {
		ColorLog("[ERRO] Fail to start[ %s ]\n", "$GOPATH is not set or empty")
		os.Exit(2)
	}

	NewWatcher(paths, extensions)
	Autobuild()

	for {
		select {
		case <-exit:
			runtime.Goexit()
		}
	}
}

func readAppDirectories(directory string, paths *[]string) {
	fileInfos, err := ioutil.ReadDir(directory)
	if err != nil {
		return
	}

	useDirectory := false
	for _, fileInfo := range fileInfos {
		if strings.HasSuffix(fileInfo.Name(), "docs") {
			continue
		}

		if isExcluded(path.Join(directory, fileInfo.Name())) {
			continue
		}

		if fileInfo.IsDir() == true && fileInfo.Name()[0] != '.' {
			readAppDirectories(directory+"/"+fileInfo.Name(), paths)
			continue
		}

		if useDirectory == true {
			continue
		}

		if checkExtension(fileInfo.Name(), extensions) {
			*paths = append(*paths, directory)
			useDirectory = true
		}
	}

	return
}

// If a file is excluded
func isExcluded(filePath string) bool {
	for _, p := range excludedPaths {
		absP, err := path.Abs(p)
		if err != nil {
			ColorLog("[ERROR] Can not get absolute path of [ %s ]\n", p)
			continue
		}
		absFilePath, err := path.Abs(filePath)
		if err != nil {
			ColorLog("[ERROR] Can not get absolute path of [ %s ]\n", filePath)
			break
		}
		if strings.HasPrefix(absFilePath, absP) {
			ColorLog("[INFO] Excluding from watching [ %s ]\n", filePath)
			return true
		}
	}
	return false
}
