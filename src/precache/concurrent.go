package precache

import (
	"context"
	"sync"
	"time"
)

// ConcurrentExecutor executes pre-caching with worker pool
type ConcurrentExecutor struct {
	processor Processor
	workers   int
	progress  ProgressReporter
}

// NewConcurrentExecutor creates a new concurrent executor
func NewConcurrentExecutor(processor Processor, workers int, progress ProgressReporter) *ConcurrentExecutor {
	if workers <= 0 {
		workers = 4 // Default to 4 workers
	}
	return &ConcurrentExecutor{
		processor: processor,
		workers:   workers,
		progress:  progress,
	}
}

// Execute processes images concurrently using a worker pool
func (e *ConcurrentExecutor) Execute(ctx context.Context, imagePaths []string) (*Stats, error) {
	startTime := time.Now()
	
	stats := &Stats{
		TotalImages: len(imagePaths),
		StartTime:   startTime,
	}
	
	if len(imagePaths) == 0 {
		stats.EndTime = time.Now()
		stats.Duration = stats.EndTime.Sub(stats.StartTime)
		return stats, nil
	}
	
	// Start progress tracking
	e.progress.Start(len(imagePaths))
	
	// Create channels for work distribution
	jobs := make(chan string, len(imagePaths))
	
	// Result tracking
	var mu sync.Mutex
	processedOK := 0
	skipped := 0
	errors := 0
	processed := 0
	
	// Start worker pool using WaitGroup.Go (Go 1.24+)
	var wg sync.WaitGroup
	
	for i := 0; i < e.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for {
				select {
				case <-ctx.Done():
					// Context cancelled
					return
				case imagePath, ok := <-jobs:
					if !ok {
						// Channel closed, no more work
						return
					}
					
					// Process the image
					err := e.processor.Process(ctx, imagePath)
					
					mu.Lock()
					processed++
					if err == nil {
						processedOK++
					} else {
						errors++
						e.progress.Error(imagePath, err)
					}
					e.progress.Update(processed, imagePath)
					mu.Unlock()
				}
			}
		}()
	}
	
	// Send jobs to workers
	go func() {
		for _, imagePath := range imagePaths {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- imagePath:
			}
		}
		close(jobs)
	}()
	
	// Wait for all workers to finish
	wg.Wait()
	
	// Calculate final stats
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	
	stats.ProcessedOK = processedOK
	stats.Skipped = skipped
	stats.Errors = errors
	stats.EndTime = endTime
	stats.Duration = duration
	
	// Complete progress tracking
	e.progress.Complete(processedOK, skipped, errors, duration)
	
	return stats, nil
}
