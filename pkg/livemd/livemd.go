package livemd

import (
	"io/ioutil"
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/microcosm-cc/bluemonday"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// Manager ...
type Manager struct {
	Directory string
	Files     map[string][]byte
	Watcher   *fsnotify.Watcher
	Done      chan interface{}
}

// AllFiles ...
func (m *Manager) AllFiles() []string {
	ff := []string{}
	for f := range m.Files {
		ff = append(ff, f)
	}
	return ff
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
			m.Files[event.Name] = bluemonday.UGCPolicy().SanitizeBytes(blackfriday.Run(b))
			log.Println(event.Name, len(m.Files[event.Name]))
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
		m.Files[ff[i]] = nil
	}
	w, err := Watcher(ff)
	if err != nil {
		panic(err)
	}
	m.Watcher = w
	go m.watch()
}

// NewManager ...
func NewManager(directory string) *Manager {
	m := Manager{
		Directory: directory,
		Files:     map[string][]byte{},
		Done:      make(chan interface{}),
	}
	m.init()
	return &m
}
