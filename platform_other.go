//go:build !windows

package main

// Versão de testes (Linux/macOS): sem bandeja, só imprime os avisos.

import (
	"fmt"
	"os"
	"os/signal"
)

func platformRun() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
func notify(title, msg string, silent bool) {
	fmt.Printf("[NOTIFICAÇÃO silent=%v] %s | %s\n", silent, title, msg)
}
func playSound(path string, vol int) { fmt.Printf("[SOM] %s vol=%d\n", path, vol) }
func openUI(url string)              { fmt.Println("abrir:", url) }
func setAutoStart(on bool) error     { return nil }
func trimMemory()                    {}
func fatal(msg string)               { fmt.Fprintln(os.Stderr, msg) }
func platformBeforeExit()            { fmt.Println("[SAINDO para atualizar]") }
func exeHeaderOK(h []byte) bool      { return len(h) == 2 }
