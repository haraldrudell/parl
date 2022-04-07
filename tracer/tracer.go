/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package tracer

import (
	"sync"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/goid"
)

const (
	trAssigned    = "Thread assigned: "
	trUnassigned  = "Thread unassigned: "
	trDefaultTask = "Thread:"
)

type Tracer struct {
	lock    sync.Mutex
	threads map[goid.ThreadID]parl.TracerTaskID
	tasks   map[parl.TracerTaskID][]parl.TracerRecord
}

func NewTracer() (tracer parl.Tracer) {
	return &Tracer{
		threads: map[goid.ThreadID]parl.TracerTaskID{},
		tasks:   map[parl.TracerTaskID][]parl.TracerRecord{}}
}

func (tr *Tracer) Assign(threadID goid.ThreadID, task parl.TracerTaskID) (tr2 parl.Tracer) {
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
			recordListBefore = append(recordListBefore, NewRecordDo(trUnassigned+string(threadID)))
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
		recordListNow = []parl.TracerRecord{NewRecordDo(string(task))} // first record is task name
	}
	recordListNow = append(recordListNow, NewRecordDo(trAssigned+string(threadID)))
	tr.tasks[task] = recordListNow
	return
}

func (tr *Tracer) Record(threadID goid.ThreadID, text string) (tr2 parl.Tracer) {
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
			recordList = []parl.TracerRecord{NewRecordDo(string(assignedTask))} // first record is task name
		}
	} else {
		recordList = tr.tasks[assignedTask]
	}

	recordList = append(recordList, NewRecordDo(text))
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
		tr.threads = map[goid.ThreadID]parl.TracerTaskID{}
		tr.tasks = map[parl.TracerTaskID][]parl.TracerRecord{}
	}
	return
}
