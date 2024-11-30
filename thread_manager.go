package main

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// OperationStatus represents the state of an operation
type OperationStatus string

const (
	StatusRunning   OperationStatus = "Running"   // Indicates the operation is currently in progress
	StatusStopped   OperationStatus = "Stopped"   // Indicates the operation was canceled
	StatusCompleted OperationStatus = "Completed" // Indicates the operation has finished
)

// Operation struct holds all metadata and state information about an individual operation
type Operation struct {
	ID        string             // Unique identifier for the operation
	Name      string             // Human-readable name of the operation
	N         int                // Target value or parameter for the operation
	Progress  int                // An integer between 0 and 100 representing the operation's progress percentage
	Status    OperationStatus    // The current status of the operation (Running, Stopped, or Completed)
	Result    string             // Result of the operation (if completed)
	CreatedAt time.Time          // Timestamp when the operation was created
	Cancel    context.CancelFunc // Context channel that listens for the cancel signal
}

// ThreadManager manages the lifecycle and state of multiple concurrent operations
type ThreadManager struct {
	operations map[string]*Operation // A hashmap of operation IDs to Operation pointers for tracking active and completed operations
	mutex      sync.Mutex            // A sync.Mutex to ensure thread-safe access to the operations map address
	logChan    chan string           // A channel for sending log messages about operations
}

// NewThreadManager creates and initializes a new ThreadManager instance
//
// Parameters:
// - logChan: A channel for transmitting log messages
//
// Returns:
// - A pointer to the newly created ThreadManager instance
func NewThreadManager(logChan chan string) *ThreadManager {
	return &ThreadManager{
		operations: make(map[string]*Operation),
		logChan:    logChan,
	}
}

// StartFibonacci starts a new Fibonacci operation as a goroutine
//
// Parameters:
// - id: A unique identifier for the operation
// - n: The n-th Fibonacci number to calculate
// - operationFunc: A function that performs the operation logic, supporting cancellation via context (see fib.go file)
//
// This method initializes the operation, stores it in the operations map, and starts the goroutine
func (tm *ThreadManager) StartFibonacci(id string, n int, operationFunc func(ctx context.Context, op *Operation, n int, tm *ThreadManager)) {
	tm.mutex.Lock()
	// Creates a new context with a 10-second timeout and a cancel function
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

// CancelOperation cancels an ongoing operation by ID if it is currently running
//
// Parameters:
// - id: The unique identifier of the operation to cancel
//
// This method uses the operation's Cancel function to signal termination and updates its status
func (tm *ThreadManager) CancelOperation(id string) {
	tm.mutex.Lock()
	if op, exists := tm.operations[id]; exists && op.Status == StatusRunning {
		op.Cancel()
		op.Status = StatusStopped
		tm.logChan <- "Canceled operation with ID " + id
	}
	tm.mutex.Unlock()
}

// CancelAllOperations cancels all currently running operations
//
// This method iterates through all stored operations and cancels those with a "Running" status.
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

// GetOperations retrieves a list of all operations, sorted by their creation time
//
// Returns:
// - A slice of Operation pointers sorted in ascending order of their creation timestamps
func (tm *ThreadManager) GetOperations() []*Operation {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	ops := make([]*Operation, 0, len(tm.operations))
	for _, op := range tm.operations {
		ops = append(ops, op)
	}

	sort.Slice(ops, func(i, j int) bool {
		return ops[i].CreatedAt.Before(ops[j].CreatedAt)
	})
	return ops
}

// GetRunningOperations retrieves a list of all currently running operations
//
// Returns:
// - A slice of Operation pointers with a status of "Running"
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

// UpdateOperation updates the state of an existing operation
//
// Parameters:
// - op: A pointer to the Operation to update
//
// This method replaces the operation in the map with the updated instance
func (tm *ThreadManager) UpdateOperation(op *Operation) {
	tm.mutex.Lock()
	tm.operations[op.ID] = op
	tm.mutex.Unlock()
}
