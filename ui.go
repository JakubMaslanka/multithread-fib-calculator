package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UI page identifiers for navigation
const (
	PageMain  = "main"  // Main page displaying operation details table, control buttons and logger
	PageForm  = "form"  // Form page to input Fibonacci parameters (n)
	PageModal = "modal" // Modal for confirmation prompts (e.g., canceling operations)
)

// setupUI initializes and configures the UI components using tview library
func setupUI(app *tview.Application, pages *tview.Pages, tm *ThreadManager) {
	// Create main layout container with vertical direction
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	// Header displaying the app title
	header := tview.NewTextView().
		SetText("Fibonacci Multi-threaded Calculator in Go Lang").
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetTextColor(tcell.ColorWhite)
	mainFlex.AddItem(header, 1, 0, false)

	// Table displaying operation details
	table := tview.NewTable().SetBorders(true)
	headers := []string{"ID", "Operation", "Fib(n)", "Progress", "Status", "Result"}
	for i, h := range headers {
		table.SetCell(0, i, tview.NewTableCell(h).
			SetAlign(tview.AlignCenter).
			SetSelectable(false).
			SetBackgroundColor(tcell.ColorBlue).
			SetTextColor(tcell.ColorWhite))
	}
	mainFlex.AddItem(table, 0, 3, true)

	// Controls for starting and canceling operations
	controls := tview.NewFlex().SetDirection(tview.FlexColumn)
	startBtn := tview.NewButton("Start Fibonacci").SetSelectedFunc(func() {
		// Navigate to form page for input fibonacci parameters
		pages.SwitchToPage(PageForm)
	})

	cancelBtn := tview.NewButton("Cancel Operation").SetSelectedFunc(func() {
		// Display a confirmation modal for canceling all operations
		modal := tview.NewModal().
			SetText("Are you sure you want to cancel all pending operations?").
			AddButtons([]string{"Yes", "No"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				if buttonLabel == "Yes" {
					tm.CancelAllOperations()
				}
				pages.RemovePage(PageModal)
			})
		pages.AddPage(PageModal, modal, true, true)
	})

	controls.AddItem(startBtn, 0, 1, false)
	controls.AddItem(cancelBtn, 0, 1, false)
	mainFlex.AddItem(controls, 3, 0, false)

	// Logger for displaying app logs and updates
	logger := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		}).
		SetTextColor(tcell.ColorGreen)
	logger.SetBorder(true).SetTitle(" Logger: ").SetTitleAlign(tview.AlignLeft)
	mainFlex.AddItem(logger, 0, 2, false)

	// Add main page to the application
	pages.AddPage(PageMain, mainFlex, true, true)

	// Form for entering Fibonacci parameters
	var form *tview.Form
	form = tview.NewForm().
		AddInputField("n", "", 15, func(text string, lastChar rune) bool {
			// Validate input as a non-negative integer
			_, err := strconv.Atoi(text)
			return err == nil || text == ""
		}, nil).
		AddButton("Start", func() {
			// Parse and validate input
			inputField := form.GetFormItem(0).(*tview.InputField)
			nStr := inputField.GetText()
			n, err := strconv.Atoi(nStr)
			if err != nil || n < 0 {
				tm.logChan <- "Invalid input for n. Please enter a non-negative integer."
				pages.SwitchToPage(PageMain)
				return
			}

			// Start Fibonacci operation
			id := generateID()
			tm.StartFibonacci(id, n, FibonacciOperation)

			// Clear input and return to main page
			inputField.SetText("")
			pages.SwitchToPage(PageMain)
		}).
		AddButton("Cancel", func() {
			pages.SwitchToPage(PageMain)
		})
	form.SetBorder(true).SetTitle(" Enter n for Fibonacci(n) ").SetTitleAlign(tview.AlignCenter)

	// Add form page to the application
	pages.AddPage(PageForm, form, true, false)

	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			ops := tm.GetOperations()
			app.QueueUpdateDraw(func() {
				// Clear and repopulate table rows
				for row := table.GetRowCount() - 1; row > 0; row-- {
					table.RemoveRow(row)
				}
				for i, op := range ops {
					table.SetCell(i+1, 0, tview.NewTableCell(op.ID).SetAlign(tview.AlignCenter).SetSelectable(false))
					table.SetCell(i+1, 1, tview.NewTableCell(op.Name).SetAlign(tview.AlignCenter).SetSelectable(false))
					table.SetCell(i+1, 2, tview.NewTableCell(strconv.Itoa(op.N)).SetAlign(tview.AlignCenter).SetSelectable(false))
					progressText := fmt.Sprintf("%d%%", op.Progress)
					table.SetCell(i+1, 3, tview.NewTableCell(progressText).SetAlign(tview.AlignCenter).SetSelectable(false))
					statusCell := tview.NewTableCell(string(op.Status)).SetAlign(tview.AlignCenter).SetSelectable(false)
					switch op.Status {
					case StatusRunning:
						statusCell.SetTextColor(tcell.ColorYellow)
					case StatusStopped:
						statusCell.SetTextColor(tcell.ColorRed)
					case StatusCompleted:
						statusCell.SetTextColor(tcell.ColorGreen)
					}
					table.SetCell(i+1, 4, statusCell)
					result := op.Result
					if op.Status != StatusCompleted {
						result = "-"
					}
					table.SetCell(i+1, 5, tview.NewTableCell(result).SetAlign(tview.AlignCenter).SetSelectable(false))
				}
			})
		}
	}()

	// Goroutine responsible for listening on the log channel (tm.logChan) for new log messages sent by other parts of the application
	//
	// The channel tm.logChan acts as a thread-safe communication mechanism between goroutines, ensuring proper synchronization
	// and avoiding race conditions. When a new message is sent to the channel, this goroutine retrieves it and updates the
	// logger view in the UI. The `app.QueueUpdateDraw` function schedules the UI update, queuing it for safe execution in the
	// main thread to prevent rendering issues.
	go func() {
		for logMsg := range tm.logChan {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(logger, "%s\n", logMsg)
			})
		}
	}()
}
