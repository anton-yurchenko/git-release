package release

import "time"

// SetRetryDelay collapses the upload backoff for the duration of a test.
//
// The production backoff sleeps 9, 27 and 81 real seconds. Tests need the retry
// LOGIC, not the wall clock, and a suite slow enough to discourage running it is
// a suite that stops catching things.
func SetRetryDelay(d time.Duration) func() {
	previous := retryDelay
	retryDelay = func(int) time.Duration { return d }

	return func() { retryDelay = previous }
}

// UploadHandler exposes a single upload attempt.
//
// Upload() swallows a retryable error and reports only the outcome of the last
// attempt, so the recovery branches are invisible from outside. Testing one
// attempt at a time is what makes them assertable.
func (a *Asset) UploadHandler(release *Release, cli RepositoriesClient, id int64, lastTry bool) error {
	return a.uploadHandler(release, cli, id, lastTry)
}
