package precache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Progress_Tracking(t *testing.T) {
	progress := &consoleProgress{
		started: false,
	}
	
	// Start tracking
	progress.Start(10)
	assert.True(t, progress.started)
	assert.Equal(t, 10, progress.total)
	
	// Update progress
	progress.Update(5, "test.jpg")
	assert.Equal(t, 5, progress.processed)
	assert.Equal(t, "test.jpg", progress.current)
	
	// Complete
	duration := 2 * time.Second
	progress.Complete(8, 1, 1, duration)
	assert.Equal(t, 8, progress.processedOK)
	assert.Equal(t, 1, progress.skipped)
	assert.Equal(t, 1, progress.errors)
}

func Test_Progress_Error(t *testing.T) {
	progress := &consoleProgress{
		started: false,
	}
	
	progress.Start(5)
	
	// Report error
	progress.Error("bad.jpg", assert.AnError)
	
	// Errors should be tracked
	assert.Equal(t, 1, len(progress.errorList))
}

func Test_Progress_MultipleUpdates(t *testing.T) {
	progress := &consoleProgress{}
	
	progress.Start(3)
	
	progress.Update(1, "image1.jpg")
	assert.Equal(t, 1, progress.processed)
	
	progress.Update(2, "image2.jpg")
	assert.Equal(t, 2, progress.processed)
	
	progress.Update(3, "image3.jpg")
	assert.Equal(t, 3, progress.processed)
}
