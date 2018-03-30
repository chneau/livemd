package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"

	"github.com/rjeczalik/notify"
	"github.com/russross/blackfriday"
	openlink "github.com/skratchdot/open-golang/open"
	"golang.org/x/net/websocket"
)

var suffixes = [3]string{".md", ".mkd", ".markdown"}

var toc []string
var tocMutex sync.Mutex
var rootTmpl *template.Template
var pageTmpl *template.Template

type state int

var host = flag.String("host", "127.0.0.1", "Host IP to listen on")
var port = flag.String("port", "8080", "Port to listen on")
var path = flag.String("path", ".", "Directory to watch")

const (
	none state = iota
	open
	close
)

type listener struct {
	File   string
	Socket *websocket.Conn
	State  state
}

type update struct {
	File string
}

type browserMsg struct {
	Markdown string
}

func init() {
	var err error
	rootTmpl, err = template.New("root").Parse(rootTemplate)
	if err != nil {
		log.Fatal(err)
	}
	pageTmpl, err = template.New("page").Parse(pageTemplate)
	if err != nil {
		log.Fatal(err)
	}
}

func hasMarkdownSuffix(s string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(strings.ToLower(s), suffix) {
			return true
		}
	}
	return false
}

func addWatch(c chan notify.EventInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			err = notify.Watch(path, c, notify.Write)
			if err != nil {
				fmt.Println("err", err)
				return err
			}
		} else if hasMarkdownSuffix(path) {
			tocMutex.Lock()
			toc = append(toc, path)
			tocMutex.Unlock()
			log.Println("Found", path)
			err = notify.Watch(path, c, notify.Write)
			if err != nil {
				fmt.Println("err", err)
				return err
			}
		}
		return nil
	}
}

func writeFileForListener(l listener) {
	var data []byte
	file, err := os.Open(l.File)
	if err != nil {
		data = []byte("Error: " + err.Error())
	}
	filebytes, err := ioutil.ReadAll(file)
	if err != nil {
		data = []byte("Error: " + err.Error())
	}
	data = blackfriday.MarkdownCommon(filebytes)
	var msg browserMsg
	msg.Markdown = string(data)
	err = websocket.JSON.Send(l.Socket, msg)
	if err != nil {
		log.Println("Error sending message:", err)
	}
}

func updateListeners(updates chan notify.EventInfo, listeners chan listener) {
	currentListeners := make([]listener, 0)
	for {
		select {
		case listener := <-listeners:
			if listener.State == open {
				listener.File = filepath.Join(listener.File)
				// log.Println("New listener on", listener.File)
				currentListeners = append(currentListeners, listener)
				writeFileForListener(listener)
			}
			if listener.State == close {
				for i, l := range currentListeners {
					if l.Socket == listener.Socket {
						// log.Println("Deregistering Listener")
						currentListeners = append(currentListeners[:i], currentListeners[i+1:]...)
					}
				}
			}
		case update := <-updates:
			for _, l := range currentListeners {
				if update.Path() == l.File {
					log.Println("Update on", update.Path())
					writeFileForListener(l)
				}
			}
		}
	}
}

func rootFunc(w http.ResponseWriter, r *http.Request) {
	tocMutex.Lock()
	localToc := make([]string, len(toc))
	copy(localToc, toc)
	tocMutex.Unlock()
	for i, s := range localToc {
		chop := strings.TrimPrefix(s, *path)
		localToc[i] = "* [" + chop + "](/md/" + chop + ")"
	}
	tocMkd := strings.Join(localToc, "\n")
	bytes := blackfriday.MarkdownCommon([]byte(tocMkd))
	rootTmpl.Execute(w, string(bytes))
}

func cssFunc(css string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(css))
	}
}

func pageFunc(w http.ResponseWriter, r *http.Request) {
	subpath := strings.TrimPrefix(r.RequestURI, "/md")
	// log.Println("New watcher on ", subpath)
	pageTmpl.Execute(w, subpath)
}

func handleListener(listeners chan listener) func(ws *websocket.Conn) {
	return func(ws *websocket.Conn) {
		subpath := strings.TrimPrefix(ws.Request().RequestURI, "/ws")
		listeners <- listener{subpath, ws, open}
		var closeMessage string
		err := websocket.Message.Receive(ws, &closeMessage)
		if err != nil && err.Error() != "EOF" {
			log.Println("Error before close:", err)
		}
		listeners <- listener{subpath, ws, close}
	}
}

func main() {
	flag.Parse()
	addr := fmt.Sprintf("%s:%s", *host, *port)
	fulladdr := fmt.Sprintf("http://%s", addr)
	updates := make(chan notify.EventInfo)
	log.Println("Serving on", fulladdr)
	log.Println("Watching directory", *path)
	abspath, _ := filepath.Abs(*path)
	path := &abspath
	err := filepath.Walk(*path, addWatch(updates))
	if err != nil {
		log.Fatal(err)
	}
	listeners := make(chan listener)
	go updateListeners(updates, listeners)
	http.HandleFunc("/", rootFunc)
	http.HandleFunc("/md/", pageFunc)
	http.HandleFunc("/github.css", cssFunc(githubCSS))
	http.Handle("/ws/", websocket.Handler(handleListener(listeners)))
	openlink.Start(fulladdr)
	err = http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalln(err)
	}
}
