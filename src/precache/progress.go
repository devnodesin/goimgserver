package precache

import (
	"fmt"
	"log"
	"time"
)

// consoleProgress implements ProgressReporter for console output
type consoleProgress struct {
	started     bool
	total       int
	processed   int
	processedOK int
	skipped     int
	errors      int
	current     string
	startTime   time.Time
	errorList   []string
}

// NewProgress creates a new console progress reporter
func NewProgress() ProgressReporter {
	return &consoleProgress{}
}

// Start begins progress tracking
func (p *consoleProgress) Start(total int) {
	p.started = true
	p.total = total
	p.startTime = time.Now()
	p.errorList = make([]string, 0)
	log.Printf("Starting pre-cache: %d images to process", total)
}

// Update updates progress with processed count
func (p *consoleProgress) Update(processed int, current string) {
	p.processed = processed
	p.current = current
	
	if p.total > 0 {
		percentage := float64(processed) / float64(p.total) * 100
		log.Printf("Pre-cache progress: %d/%d (%.1f%%) - %s", processed, p.total, percentage, current)
	}
}

// Complete marks progress as complete
func (p *consoleProgress) Complete(processedOK int, skipped int, errors int, duration time.Duration) {
	p.processedOK = processedOK
	p.skipped = skipped
	p.errors = errors
	
	log.Printf("Pre-cache complete: %d processed, %d skipped, %d errors in %v", 
		processedOK, skipped, errors, duration)
	
	if len(p.errorList) > 0 {
		log.Printf("Errors encountered during pre-cache:")
		for _, err := range p.errorList {
			log.Printf("  - %s", err)
		}
	}
}

// Error reports an error
func (p *consoleProgress) Error(imagePath string, err error) {
	errorMsg := fmt.Sprintf("%s: %v", imagePath, err)
	p.errorList = append(p.errorList, errorMsg)
	log.Printf("Pre-cache error: %s", errorMsg)
}
