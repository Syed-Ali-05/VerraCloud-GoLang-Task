package main

import (
    "os"
    log "github.com/sirupsen/logrus"
	"github.com/Syed-Ali-05/VerraCloud-GoLang-Task/internal/app"
)

func main() {
    log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
    log.SetLevel(log.InfoLevel)

    addr := ":8080"
    if envAddr := os.Getenv("ADDR"); envAddr != "" {
        addr = envAddr
    }

    log.Infof("Starting server on %s", addr)
    if err := app.Run(addr); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
