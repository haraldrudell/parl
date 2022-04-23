/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"time"
)

// Tracer lists events in terms of tasks rather than per time or thread.
// A task is executed by threads assigned to it.
// Threads are uniquely identified by threadID.
// A task can have zero or one threads assigned at any one time.
// A thread can be assigned to zero or one tasks.
// Each task has an ID, a name and a list of events and thread assignments
// Tracer can record branching in the code and return that for a particular
// item being processed. For an item processed incorrectly, or when
// threads hang, Tracer will find unfavorable branching and last known locations.
type Tracer interface {
	// AssignTaskToThread assigns a Thread to a task
	AssignTaskToThread(threadID ThreadID, task TracerTaskID) (tracer Tracer)
	// RecordTaskEvent adds an event to the task threadID is currently assigned to.
	// If threadID is not assigned, a new task is created.
	RecordTaskEvent(threadID ThreadID, text string) (tracer Tracer)
	// Records returns the current map of tasks and their events
	Records(clear bool) (records map[TracerTaskID][]TracerRecord)
}

type TracerTaskID string

type TracerRecord interface {
	Values() (at time.Time, text string)
}

type TracerFactory interface {
	// NewTracer creates tracer storage.
	// use false indicates a nil Tracer whose output will not be used
	NewTracer(use bool) (tracer Tracer)
}
