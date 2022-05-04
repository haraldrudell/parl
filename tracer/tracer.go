/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

// tracer has events by task then time rather than time or thread
package tracer

import (
	"sync"

	"github.com/haraldrudell/parl"
)

const (
	trAssigned    = "Thread assigned: "
	trUnassigned  = "Thread unassigned: "
	trDefaultTask = "Thread:"
)

type Tracer struct {
	lock    sync.Mutex
	threads map[parl.ThreadID]parl.TracerTaskID
	tasks   map[parl.TracerTaskID][]parl.TracerRecord
}

func NewTracer() (tracer parl.Tracer) {
	return &Tracer{
		threads: map[parl.ThreadID]parl.TracerTaskID{},
		tasks:   map[parl.TracerTaskID][]parl.TracerRecord{}}
}

func (tr *Tracer) AssignTaskToThread(threadID parl.ThreadID, task parl.TracerTaskID) (tr2 parl.Tracer) {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	tr2 = tr

	// find out thread’s current assignment
	var taskBefore parl.TracerTaskID
	var recordListBefore []parl.TracerRecord
	var ok0 bool
	if taskBefore, ok0 = tr.threads[threadID]; ok0 {
		recordListBefore = tr.tasks[taskBefore]
		// unassign from previous task as appropriate
		if task == "" || task != taskBefore {
			recordListBefore = append(recordListBefore, NewTracerRecord(trUnassigned+string(threadID)))
			tr.tasks[taskBefore] = recordListBefore
		}
	}

	// unassignment
	if task == "" {
		delete(tr.threads, threadID)
		return
	}

	// new assignment
	tr.threads[threadID] = task
	recordListNow := tr.tasks[task]
	if len(recordListNow) == 0 {
		recordListNow = []parl.TracerRecord{NewTracerRecord(string(task))} // first record is task name
	}
	recordListNow = append(recordListNow, NewTracerRecord(trAssigned+string(threadID)))
	tr.tasks[task] = recordListNow
	return
}

func (tr *Tracer) RecordTaskEvent(threadID parl.ThreadID, text string) (tr2 parl.Tracer) {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	tr2 = tr

	// if thread is unassigned, use thread task
	var assignedTask parl.TracerTaskID
	var ok0 bool
	var recordList []parl.TracerRecord
	if assignedTask, ok0 = tr.threads[threadID]; !ok0 {
		assignedTask = parl.TracerTaskID(trDefaultTask + string(threadID))
		recordList = tr.tasks[assignedTask]
		if len(recordList) == 0 {
			recordList = []parl.TracerRecord{NewTracerRecord(string(assignedTask))} // first record is task name
		}
	} else {
		recordList = tr.tasks[assignedTask]
	}

	recordList = append(recordList, NewTracerRecord(text))
	tr.tasks[assignedTask] = recordList
	return
}

func (tr *Tracer) Records(clear bool) (records map[parl.TracerTaskID][]parl.TracerRecord) {
	tr.lock.Lock()
	defer tr.lock.Unlock()

	records = map[parl.TracerTaskID][]parl.TracerRecord{}
	for key, recordSlice := range tr.tasks {
		slice2 := make([]parl.TracerRecord, len(recordSlice))
		copy(slice2, recordSlice)
		records[key] = slice2
	}
	if clear {
		tr.threads = map[parl.ThreadID]parl.TracerTaskID{}
		tr.tasks = map[parl.TracerTaskID][]parl.TracerRecord{}
	}
	return
}
