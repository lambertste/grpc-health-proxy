// Package latch provides a one-shot gate (latch) that starts in the closed
// state and can be opened exactly once.
//
// # Overview
//
// A Latch is useful when a service must block traffic until some initialisation
// step completes — for example, waiting for a configuration reload, a database
// migration, or a leader-election result — without the overhead of a full
// warmup timer.
//
// # Usage
//
//	l := latch.New()
//
//	// Block until ready.
//	go func() {
//		runMigrations()
//		l.Open()
//	}()
//
//	http.Handle("/", l.Middleware(appHandler))
//
// Once Open is called all subsequent Allow calls return true immediately via
// an atomic load with no mutex contention.
package latch
