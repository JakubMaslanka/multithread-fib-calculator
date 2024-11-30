package main

import (
	"context"
	"fmt"
	"math/big"
	"time"
)

// FibonacciOperation calculates the n-th Fibonacci number using the math/big package to handle large integers
// It runs in a separate goroutine and supports cancellation via context.Context
//
// Parameters:
// - ctx: A context.Context object that allows for cancellation of the operation
// - op: A pointer to an Operation struct that holds metadata and progress information about the operation
// - n: An integer representing the position in the Fibonacci sequence to calculate
// - tm: A pointer to a ThreadManager struct that manages logging and operation updates
func FibonacciOperation(ctx context.Context, op *Operation, n int, tm *ThreadManager) {
	// Sends a log message when the operation starts
	tm.logChan <- fmt.Sprintf("Fibonacci operation %s started for n=%d", op.ID, n)

	// Initializes a slice of *big.Int to store Fibonacci numbers up to the n-th position
	fib := make([]*big.Int, n+1)
	fib[0] = big.NewInt(0)
	if n > 0 {
		fib[1] = big.NewInt(1)
	}

	// Ticker periodically updates the Fibonacci calculation progress and checks for cancellation signals
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	for i := 2; i <= n; i++ {
		select {
		case <-ctx.Done(): // If the cancel signal is sent via the context, the function logs the cancellation and exits early
			tm.logChan <- fmt.Sprintf("Fibonacci operation %s canceled at i=%d", op.ID, i)
			return
		case <-ticker.C: // At every tick, the progress bar is updated and the current step of the calculation is displayed in operation table
			fib[i] = new(big.Int).Add(fib[i-1], fib[i-2])
			op.Progress = (i * 100) / n
			tm.UpdateOperation(op)
		}

		// When the calculation is complete, the result is stored in the Operation struct, and the status is set to StatusCompleted
		if i == n {
			op.Result = fib[n].String()
			op.Status = StatusCompleted
			tm.UpdateOperation(op)
			tm.logChan <- fmt.Sprintf("Fibonacci operation %s completed: fib(%d) = %s", op.ID, n, op.Result)
		}
	}
}
