/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

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
	threads map[string]string
	tasks   map[string][]parl.Record
}

func NewTracer() (tracer parl.Tracer) {
	return &Tracer{threads: map[string]string{}, tasks: map[string][]parl.Record{}}
}

func (tr *Tracer) Assign(threadID, task string) (tr2 parl.Tracer) {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	tr2 = tr

	// find out thread’s current assignment
	var task0 string
	var record0 []parl.Record
	var ok0 bool
	if task0, ok0 = tr.threads[threadID]; ok0 {
		record0 = tr.tasks[task0]
		// unassign from previous task as appropriate
		if task == "" || task != task0 {
			record0 = append(record0, NewRecordDo(trUnassigned+threadID))
			tr.tasks[task0] = record0
		}
	}

	// unassignment
	if task == "" {
		delete(tr.tasks, threadID)
		return
	}

	// new assignment
	tr.threads[threadID] = task
	record1 := tr.tasks[task]
	if len(record1) == 0 {
		record1 = []parl.Record{NewRecordDo(task)} // first record is task name
	}
	record1 = append(record1, NewRecordDo(trAssigned+threadID))
	tr.tasks[task] = record1
	return
}

func (tr *Tracer) Record(threadID, text string) (tr2 parl.Tracer) {
	tr.lock.Lock()
	defer tr.lock.Unlock()
	tr2 = tr

	// if thread is unassigned, use thread task
	var task0 string
	var ok0 bool
	var record []parl.Record
	if task0, ok0 = tr.threads[threadID]; !ok0 {
		task0 = trDefaultTask + threadID
		record = tr.tasks[task0]
		if len(record) == 0 {
			record = []parl.Record{NewRecordDo(task0)} // first record is task name
		}
	} else {
		record = tr.tasks[task0]
	}

	record = append(record, NewRecordDo(text))
	tr.tasks[task0] = record
	return
}

func (tr *Tracer) Records(clear bool) (records map[string][]parl.Record) {
	tr.lock.Lock()
	defer tr.lock.Unlock()

	records = map[string][]parl.Record{}
	for key, recordSlice := range tr.tasks {
		slice2 := make([]parl.Record, len(recordSlice))
		copy(slice2, recordSlice)
		records[key] = slice2
	}
	if clear {
		tr.threads = map[string]string{}
		tr.tasks = map[string][]parl.Record{}
	}
	return
}
