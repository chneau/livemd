package livemd

import (
	"fmt"
	"io/ioutil"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
	"github.com/shurcooL/github_flavored_markdown"
)

// Manager ...
type Manager struct {
	Directory string
	Watcher   *fsnotify.Watcher
	conns     map[*websocket.Conn]interface{}
	read      chan interface{}
	latest    interface{}
}

func (m *Manager) keepDispatching() {
	for {
		message := <-m.read
		m.latest = message
		for c := range m.conns {
			if c.WriteJSON(message) != nil {
				delete(m.conns, c)
			}
		}
	}
}

// AddConn ...
func (m *Manager) AddConn(ws *websocket.Conn) {
	m.conns[ws] = nil
	if m.latest != nil {
		ws.WriteJSON(m.latest)
	}
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
			m.read <- map[string]string{
				event.Name: string(github_flavored_markdown.Markdown(b)),
			}
		case err := <-m.Watcher.Errors:
			if err != nil {
				panic(err)
			}
		}
	}
}

func (m *Manager) init() {
	ff, err := MarkdownFiles(m.Directory)
	if err != nil {
		panic(err)
	}
	for i := range ff {
		fmt.Println(ff[i])
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
		read:      make(chan interface{}),
		conns:     map[*websocket.Conn]interface{}{},
	}
	m.init()
	return &m
}
