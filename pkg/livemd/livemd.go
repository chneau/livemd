package livemd

import (
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/shurcooL/github_flavored_markdown"
)

// Manager ...
type Manager struct {
	Directory string
	Files     map[string]string
	Watcher   *fsnotify.Watcher
	conns     map[*websocket.Conn]interface{}
	read      chan interface{}
}

// AllFiles ...
func (m *Manager) AllFiles() []string {
	ff := []string{}
	for f := range m.Files {
		ff = append(ff, f)
	}
	return ff
}

func (m *Manager) keepDispatching() {
	for {
		for c := range m.conns {
			message := <-m.read
			if c.WriteJSON(message) != nil {
				delete(m.conns, c)
			}
		}
	}
}

// AddConn ...
func (m *Manager) AddConn(ws *websocket.Conn) {
	m.conns[ws] = nil
	ws.WriteJSON(m.Files)
}

func (m *Manager) watch() {
	for {
		select {
		case event := <-m.Watcher.Events:
			if event.Op != fsnotify.Write {
				continue
			}
			b, err := ioutil.ReadFile(event.Name)
			if err != nil {
				panic(err)
			}
			if len(b) == 0 {
				continue
			}
			m.Files[event.Name] = byteToHTML(b)
			m.read <- map[string]string{
				event.Name: m.Files[event.Name],
			}
		case err := <-m.Watcher.Errors:
			if err != nil {
				panic(err)
			}
		}
	}
}

func byteToHTML(b []byte) string {
	// p := bluemonday.UGCPolicy()
	// p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$")).OnElements("code")
	return string(github_flavored_markdown.Markdown(b))
}

func (m *Manager) init() {
	ff, err := MarkdownFiles(m.Directory)
	if err != nil {
		panic(err)
	}
	for i := range ff {
		b, err := ioutil.ReadFile(ff[i])
		if err != nil {
			panic(err)
		}
		m.Files[ff[i]] = byteToHTML(b)
	}
	w, err := Watcher(ff)
	if err != nil {
		panic(err)
	}
	m.Watcher = w
	go m.watch()
	go m.keepDispatching()
}

// NewManager ...
func NewManager(directory string) *Manager {
	m := Manager{
		Directory: directory,
		Files:     map[string]string{},
		read:      make(chan interface{}),
		conns:     map[*websocket.Conn]interface{}{},
	}
	m.init()
	return &m
}
