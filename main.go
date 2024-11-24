// main.go
package main

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/rivo/tview"
)

func main() {
    // Inicjalizacja loggera
    logChan := make(chan string, 100)
    defer close(logChan)

    // Inicjalizacja ThreadManager
    tm := NewThreadManager(logChan)

    // Inicjalizacja UI
    app := tview.NewApplication()
    pages := tview.NewPages()

    setupUI(app, pages, tm)

    // Uruchomienie aplikacji
    if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
        panic(err)
    }
}

// generateID tworzy unikalne ID na podstawie losowych bajtów i kodowania heksadecymalnego.
func generateID() string {
    bytes := make([]byte, 4) // 4 bajty = 8 znaków hex
    _, err := rand.Read(bytes)
    if err != nil {
        panic(err) // Możesz obsłużyć błąd w inny sposób
    }
    return hex.EncodeToString(bytes)
}