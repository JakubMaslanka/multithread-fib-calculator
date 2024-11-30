package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/rivo/tview"
)

func main() {
    // The logger is initialized with a buffered channel to handle log messages asynchronously
    logChan := make(chan string, 100)
    defer close(logChan)
    
    // The ThreadManager is created to manage concurrent operations within the application
    tm := NewThreadManager(logChan)
    
    // The user interface is built using the tview library, with an application and pages setup
    app := tview.NewApplication()
    pages := tview.NewPages()
    
    // The setupUI function is called to configure the UI components and link them with the ThreadManager
    setupUI(app, pages, tm)
    
    // Finally, the application is started with mouse support enabled. If the application fails to run,
    // it will panic and display the error message
    if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
        panic(err)
    }
}

// generateID creates a unique ID based on random bytes and hexadecimal encoding
func generateID() string {
    bytes := make([]byte, 4)
    _, err := rand.Read(bytes)
    if err != nil {
        fmt.Println("Failed to generate random ID:", err)
        panic(err)
    }
    return hex.EncodeToString(bytes)
}