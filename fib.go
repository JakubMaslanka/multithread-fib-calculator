package main

import (
	"context"
	"fmt"
	"math/big"
	"time"
)

// FibonacciOperation oblicza n-ty wyraz ciągu Fibonacciego.
func FibonacciOperation(ctx context.Context, op *Operation, n int, tm *ThreadManager) {
    tm.logChan <- fmt.Sprintf("Fibonacci operation %s started for n=%d", op.ID, n)

    // Użycie math/big dla dużych liczb
    fib := make([]*big.Int, n+1)
    fib[0] = big.NewInt(0)
    if n > 0 {
        fib[1] = big.NewInt(1)
    }

    ticker := time.NewTicker(10 * time.Millisecond)
    defer ticker.Stop()

    for i := 2; i <= n; i++ {
        select {
			case <-ctx.Done():
				tm.logChan <- fmt.Sprintf("Fibonacci operation %s canceled at i=%d", op.ID, i)
				return
			case <-ticker.C:
				fib[i] = new(big.Int).Add(fib[i-1], fib[i-2])
				// Aktualizacja postępu
				op.Progress = (i * 100) / n
				tm.UpdateOperation(op)
		}

		// Ustawienie wyniku po zakończeniu pętli
		op.Result = fib[n].String()
		op.Status = StatusCompleted
		tm.UpdateOperation(op)
		tm.logChan <- fmt.Sprintf("Fibonacci operation %s completed: fib(%d) = %s", op.ID, n, op.Result)
	}
}