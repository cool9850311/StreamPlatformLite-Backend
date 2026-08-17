package mock_data

import (
	"sync"
	"time"

	"github.com/stretchr/testify/mock"
)

// MockFfmpegLibrary is a testify mock for the ffmpeg library.
//
// On top of the regular testify behaviour it offers three concurrency safe
// knobs required by the record-status tests:
//   - a call counter (CallCount) readable while a conversion is in flight,
//   - an injectable delay (SetDelay) so ConvertStreamToMp4 can block and keep
//     the "converting" window observable,
//   - an injectable error (SetError) so the failure path can be exercised.
//
// The mock stays permissive by default: when no expectation has been declared
// with .On("ConvertStreamToMp4", ...) it simply applies the injected delay and
// returns the injected error, so tests that never convert are unaffected.
type MockFfmpegLibrary struct {
	mock.Mock

	mu        sync.Mutex
	callCount int
	delay     time.Duration
	err       error
}

// SetDelay makes every subsequent ConvertStreamToMp4 call block for d.
func (m *MockFfmpegLibrary) SetDelay(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.delay = d
}

// SetError makes every subsequent ConvertStreamToMp4 call return err.
// Ignored when an explicit testify expectation is registered.
func (m *MockFfmpegLibrary) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

// CallCount returns how many times ConvertStreamToMp4 has been entered.
// It is safe to call while a conversion is still running.
func (m *MockFfmpegLibrary) CallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

func (m *MockFfmpegLibrary) ConvertStreamToMp4(filePath string) error {
	m.mu.Lock()
	m.callCount++
	delay := m.delay
	injectedErr := m.err
	m.mu.Unlock()

	if delay > 0 {
		time.Sleep(delay)
	}

	// Honour explicit testify expectations when the test declared any,
	// otherwise fall back to the injected behaviour without failing.
	if len(m.ExpectedCalls) > 0 {
		args := m.Called(filePath)
		return args.Error(0)
	}

	return injectedErr
}
