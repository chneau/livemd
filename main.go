package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/chneau/tt"
	"github.com/gorilla/websocket"

	"github.com/chneau/livemd/pkg/livemd"

	"github.com/gin-gonic/gin"
)

var (
	port string
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	if runtime.GOOS == "windows" {
		gin.DisableConsoleColor()
	}
	gracefulExit()
	log.SetFlags(log.LstdFlags)
	flag.StringVar(&port, "port", "8888", "port to listen on")
}

func gracefulExit() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		os.Exit(0)
	}()
}

// checkError
func ce(err error, msg string) {
	if err != nil {
		log.Panicln(msg, err)
	}
}
func main() {
	defer tt.Track(time.Now(), "main")
	m := livemd.NewManager(".")
	r := gin.Default()
	r.Use(gin.Recovery())
	r.GET("/ws", func(c *gin.Context) {
		conn, _ := websocket.Upgrade(c.Writer, c.Request, c.Writer.Header(), 1024, 1024)
		_ = conn
		// hub.AddConn(conn)
	})
	r.GET("/", func(c *gin.Context) {
		c.Redirect(307, "/livemd")
	})
	r.Static("/livemd", "static")
	hostname, _ := os.Hostname()
	log.Printf("Listening on http://%[1]s:%[2]s/ , http://localhost:%[2]s/\n", hostname, port)
	r.Run(":" + port)
	<-m.Done
}
