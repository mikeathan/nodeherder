package utils_test

import (
	"errors"
	"node-herder/utils"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewActiveStateTimer(t *testing.T) {
	mockCallback := func(bool) error { return nil }
	timer := utils.NewActiveStateTimer(mockCallback)

	assert.NotNil(t, timer, "NewActiveStateTimer should return a non-nil timer")
}

func TestStartActiveState_AlreadyActive(t *testing.T) {
	mockCallback := func(bool) error { return nil }
	timer := utils.NewActiveStateTimer(mockCallback)

	err := timer.Start(time.Second)
	assert.NoError(t, err, "First Start should not return an error")

	err = timer.Start(time.Second)
	assert.EqualError(t, err, "state already active", "Second Start should return an error")
}

func TestStartActiveState_CallbackError(t *testing.T) {
	mockCallback := func(bool) error { return errors.New("callback error") }
	timer := utils.NewActiveStateTimer(mockCallback)

	err := timer.Start(time.Second)

	assert.EqualError(t, err, "callback error", "Start should return the callback error")
}

func TestStartActiveState_Success(t *testing.T) {
	mockCallback := func(bool) error { return nil }
	timer := utils.NewActiveStateTimer(mockCallback)

	duration := 100 * time.Millisecond

	err := timer.Start(time.Microsecond)
	assert.NoError(t, err, "Start should not return an error")

	time.Sleep(duration + 20*time.Millisecond) // Wait for timer to expire

	err = timer.Start(duration) // Subsequent start should now succeed
	assert.NoError(t, err, "Subsequent Start should now succeed")
}

func TestStopActiveStateTimer_NoTimer(t *testing.T) {
	mockCallback := func(bool) error { return nil }
	timer := utils.NewActiveStateTimer(mockCallback)

	err := timer.Stop(true)
	assert.NoError(t, err, "Stop should not return an error")
}

func TestStopActiveStateTimer_WithTimer(t *testing.T) {
	mockCallback := func(bool) error { return nil }
	timer := utils.NewActiveStateTimer(mockCallback)

	duration := 100 * time.Millisecond
	timer.Start(duration)

	err := timer.Stop(true)
	assert.NoError(t, err, "Stop should not return an error")

	time.Sleep(duration + 20*time.Millisecond) // Wait a bit after stop
	err = timer.Start(duration)                // Subsequent start should now succeed
	assert.NoError(t, err, "Subsequent Start should now succeed")
}

func TestStopActiveStateTimer_WithoutCallback(t *testing.T) {
	mockCallback := func(bool) error { return nil }
	timer := utils.NewActiveStateTimer(mockCallback)

	duration := 100 * time.Millisecond
	timer.Start(duration)

	err := timer.Stop(false)
	assert.NoError(t, err, "Stop should not return an error")

	time.Sleep(duration + 20*time.Millisecond) // Wait a bit after stop
	err = timer.Start(duration)                // Subsequent start should now succeed
	assert.NoError(t, err, "Subsequent Start should now succeed")
}

func TestConcurrentStartStop(t *testing.T) {
	mockCallback := func(bool) error { return nil }
	timer := utils.NewActiveStateTimer(mockCallback)

	var wg sync.WaitGroup
	numRoutines := 10

	for i := 0; i < numRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				timer.Start(10 * time.Millisecond)
				time.Sleep(5 * time.Millisecond)
				timer.Stop(false)
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
}
