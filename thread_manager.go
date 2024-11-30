// manager.go
package main

import (
	"context"
	"sort"
	"sync"
	"time"
)

// OperationStatus reprezentuje status operacji.
type OperationStatus string

const (
    StatusRunning   OperationStatus = "Running"
    StatusStopped   OperationStatus = "Stopped"
    StatusCompleted OperationStatus = "Completed"
)

// Operation przechowuje informacje o pojedynczej operacji.
type Operation struct {
    ID        string
    Name      string
    N         int
    Progress  int
    Status    OperationStatus
    Result    string
    CreatedAt time.Time
    Cancel    context.CancelFunc
}

// ThreadManager zarządza operacjami (goroutines).
type ThreadManager struct {
    operations map[string]*Operation
    mutex      sync.Mutex
    logChan    chan string
}

// NewThreadManager tworzy nowy ThreadManager.
func NewThreadManager(logChan chan string) *ThreadManager {
    return &ThreadManager{
        operations: make(map[string]*Operation),
        logChan:    logChan,
    }
}

// StartFibonacci uruchamia operację Fibonacciego w nowej goroutine.
func (tm *ThreadManager) StartFibonacci(id string, n int, operationFunc func(ctx context.Context, op *Operation, n int, tm *ThreadManager)) {
    tm.mutex.Lock()
    ctx, cancel := context.WithCancel(context.Background())
    op := &Operation{
        ID:        id,
        Name:      "Fibonacci",
        N:         n,
        Progress:  0,
        Status:    StatusRunning,
        Result:    "",
        CreatedAt: time.Now(),
        Cancel:    cancel,
    }
    tm.operations[id] = op
    tm.mutex.Unlock()

    tm.logChan <- "Started Fibonacci operation with ID " + id

    // Launches a goroutines to manage the Fibonacci operation execution and monitor its status
    // The first goroutine runs the operation function and signals completion via the `done` channel
    // The outer goroutine monitors for either operation completion or timeout/cancellation via a `select` statement
    go func() {
		done := make(chan bool)

		go func() {
			operationFunc(ctx, op, n, tm)
			done <- true
		}()

		select {
            case <-done: // Operation completed successfully
            case <-ctx.Done(): // Handle timeout or cancellation
                if ctx.Err() == context.DeadlineExceeded {
                    tm.logChan <- fmt.Sprintf("Timeout: Fibonacci operation for n=%d (ID: %s) exceeded 10 seconds.", n, id)
                }
                tm.CancelOperation(id)
		}
	}()
}
func (tm *ThreadManager) CancelOperation(id string) {
    tm.mutex.Lock()
    if op, exists := tm.operations[id]; exists && op.Status == StatusRunning {
        op.Cancel()
        op.Status = StatusStopped
        tm.logChan <- "Canceled operation with ID " + id
    }
    tm.mutex.Unlock()
}

// CancelAllOperations anuluje wszystkie bieżące operacje.
func (tm *ThreadManager) CancelAllOperations() {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    for id, op := range tm.operations {
        if op.Status == StatusRunning {
            op.Cancel()
            op.Status = StatusStopped
            tm.logChan <- "Canceled operation with ID " + id
        }
    }
}

// GetOperations zwraca listę wszystkich operacji posortowaną według CreatedAt.
func (tm *ThreadManager) GetOperations() []*Operation {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    ops := make([]*Operation, 0, len(tm.operations))
    for _, op := range tm.operations {
        ops = append(ops, op)
    }
    // Sortowanie operacji według CreatedAt
    sort.Slice(ops, func(i, j int) bool {
        return ops[i].CreatedAt.Before(ops[j].CreatedAt)
    })
    return ops
}

// GetRunningOperations zwraca listę wszystkich bieżących operacji.
func (tm *ThreadManager) GetRunningOperations() []*Operation {
    tm.mutex.Lock()
    defer tm.mutex.Unlock()
    running := []*Operation{}
    for _, op := range tm.operations {
        if op.Status == StatusRunning {
            running = append(running, op)
        }
    }
    return running
}

// UpdateOperation aktualizuje informacje o operacji.
func (tm *ThreadManager) UpdateOperation(op *Operation) {
    tm.mutex.Lock()
    tm.operations[op.ID] = op
    tm.mutex.Unlock()
}