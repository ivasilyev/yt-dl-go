package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"sync"
)

var (
	globalYtDlBinFile        string
	globalYtDlCliArgs        string
	globalYtDlNamingTemplate string

	cmdArraysToProcessChannel = make(chan []string, 0)
	mutex                     sync.Mutex
)

func processCmdArray(ytDlArray []string) {
	log.Println("Process Cmd Array")
	cmd := exec.Command(globalYtDlBinFile, ytDlArray...)
	stdout, stderr := cmd.CombinedOutput()
	log.Println(string(stdout))
	if stderr != nil {
		log.Println(string(stdout))
		return
	}
}

func checkCmdArrays() {
	log.Println("Check Cmd Arrays")
	for ytDlArray := range cmdArraysToProcessChannel {
		processCmdArray(ytDlArray)
	}
}

func main() {
	var port int
	flag.IntVar(&port, "port", 8080, "Port to listen on")
	flag.StringVar(&globalYtDlBinFile, "bin", "yt-dlp", "Path to the Youtube-dl(p) binary file")
	flag.StringVar(&globalYtDlCliArgs, "args", "", "Youtube-dl(p) arguments string")
	flag.StringVar(&globalYtDlNamingTemplate, "name", "", "Youtube-dl(p) output naming template")
	flag.Parse()
	log.Printf("Start server on port: %d", port)
	log.Printf("Youtube-dl(p) binary file: %s", globalYtDlBinFile)
	log.Printf("Youtube-dl(p) arguments: %s", globalYtDlCliArgs)
	log.Printf("Youtube-dl(p) output naming template: %s", globalYtDlNamingTemplate)

	http.HandleFunc("/", formHandler)
	http.HandleFunc("/run", runHandler)

	go checkCmdArrays()
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Send web page")
	tmpl := template.Must(template.ParseFiles("form.html"))
	_ = tmpl.Execute(w, nil)
}

func runHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "The request method is not POST: ", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	videoUrl := r.FormValue("argument")
	ytDlArgs := strings.Fields(globalYtDlCliArgs)
	ytDlArgs = append(ytDlArgs, globalYtDlNamingTemplate, videoUrl)
	mutex.Lock()
	cmdArraysToProcessChannel <- ytDlArgs
	mutex.Unlock()
}
