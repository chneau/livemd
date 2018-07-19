package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/chneau/tt"

	"github.com/chneau/livemd/pkg/livemd"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
	if runtime.GOOS == "windows" {
		gin.DisableConsoleColor()
	}
	gracefulExit()
	log.SetFlags(log.LstdFlags)
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
	<-m.Done
}
