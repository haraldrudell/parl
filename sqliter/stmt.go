/*
© 2022–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
All rights reserved.
*/

package sqliter

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/haraldrudell/parl"
	"github.com/haraldrudell/parl/ptime"
)

const (
	// because sub queries somehow lock up penultimate query,
	// have a max wait time
	sqMaxWait = 5 * time.Millisecond
	// sqPrintSlow indicates how much retries must delay a query in order to
	// produce diagnostic printing.
	// c68x at start query #2 of 250 is delayed 106 ms due to database startup delays.
	sqPrintSlow = 200 * time.Millisecond
	// sqPeriodIsIdle defines a time period during which if there is
	// a successful query, a held-up query will wait rathr than produce
	// a busy error.
	// Database write duration typically 20 ms.
	// Database read duration typically 400μs.
	// Value source:  c68x empirics which has 215 initial queries.
	sqPeriodIsIdle = 180 * time.Millisecond
	// sqShortTime is the time to wait when a query indictaes database busy,
	// but the database is not busy by us.
	sqShortTime = time.Millisecond
)

var lastLock sync.Mutex

// lastQuerySucess holds the last time a query against the database was successful
var lastQuerySucess time.Time     // behind lastLock
var longestDuration time.Duration // behind lastLock

var queue = sync.NewCond(&sync.Mutex{})
var lastQueueID int    // behind queue
var executingID int    // behind queue
var executingCount int // behind queue

type Stmt struct {
	*sql.Stmt
	ds *DataSource
}

func (st *Stmt) ExecContext(ctx context.Context, args ...any) (sqlResult sql.Result, err error) {
	st.retry(func() (e error) {
		sqlResult, err = st.Stmt.ExecContext(ctx, args...)
		return err
	})
	return
}
func (st *Stmt) QueryContext(ctx context.Context, args ...any) (sqlRows *sql.Rows, err error) {
	st.retry(func() (e error) {
		sqlRows, err = st.Stmt.QueryContext(ctx, args...)
		return err
	})
	return
}
func (st *Stmt) QueryRowContext(ctx context.Context, args ...any) (sqlRow *sql.Row) {
	st.retry(func() (e error) {
		sqlRow = st.Stmt.QueryRowContext(ctx, args...)
		return sqlRow.Err()
	})
	return
}

func (st *Stmt) retry(query func() (err error)) {
	// use counters to measure query concurrency
	c := st.ds.counters.GetOrCreateCounter(sqStatement).Inc()
	defer c.Dec()
	// diagnostic printing
	t0 := time.Now()  // t0 measures total duration of query
	tLast := t0       // tLast measures execution time of query once successful
	var now time.Time // time that last of this query completed
	defer func() {
		totalQueryDuration := now.Sub(t0)
		queryExecutionTime := now.Sub(tLast)
		retryDelay := totalQueryDuration - queryExecutionTime
		if !shouldQueryPrint(retryDelay) {
			return // no print exit
		}
		incs, decs, max, _, _ := c.CounterValue(false).Get()
		parl.Info("query-count: %d(conc: %d) idle: %s query: total-duration: %s execution-time: %s",
			incs-decs, max, ptime.Duration(idleDuration(now)),
			ptime.Duration(totalQueryDuration), ptime.Duration(queryExecutionTime),
		)
	}()

	// register queries
	registerQuery()
	defer unregisterQuery()

	// handle dequeueing
	var queueID int
	defer dequeue(&queueID)

	for {

		// execute the query
		tLast = time.Now()
		err := query() // query() retains the error internally
		now = time.Now()

		// handle success
		if err == nil {
			updateSuccessTime()
			return // successful query
		}

		// errors other than database unavailable
		if code, _ := Code(err); code != CodeBusy && code != CodeDatabaseIsLocked {
			return // error other than sqlite busy
		}

		// ensure database state is applicable
		if isIdle(now) {
			return // the database is not busy, unavailable errors should not occur
		}

		// enqueue this delayed request
		if enqueue(&queueID) { // blocks until our turn
			time.Sleep(sqShortTime)
		}
	}
}

func updateSuccessTime(t ...time.Time) {
	var t0 time.Time
	if len(t) > 0 {
		t0 = t[0]
	}
	if t0.IsZero() {
		t0 = time.Now()
	}
	lastLock.Lock()
	defer lastLock.Unlock()

	if t0.After(lastQuerySucess) {
		lastQuerySucess = t0
	}
}

func isIdle(t time.Time) (idle bool) {
	lastLock.Lock()
	defer lastLock.Unlock()

	if lastQuerySucess.IsZero() {
		// no query has succeeded yet.
		// to avoid an initial barrage of many queries to print many errors,
		// assume database is not idle.
		return
	}

	// deteremine if database is idle
	durationSinceLastSuccessfulQuery := t.Sub(lastQuerySucess)
	idle = durationSinceLastSuccessfulQuery >= sqPeriodIsIdle
	return
}

func idleDuration(now time.Time) (idleTime time.Duration) {
	lastLock.Lock()
	defer lastLock.Unlock()

	if idleTime = now.Sub(lastQuerySucess); idleTime < 0 {
		idleTime = 0 // queries may complete and update lastQuerySucess while calculating
	}
	return
}

func shouldQueryPrint(d time.Duration) (print bool) {
	if d < sqPrintSlow {
		return // query took less time than idle period
	}
	lastLock.Lock()
	defer lastLock.Unlock()

	if d > longestDuration {
		longestDuration = d
		print = true
	}

	return
}

func registerQuery() {
	queue.L.Lock()
	defer queue.L.Unlock()

	executingCount++
}

func unregisterQuery() {
	queue.L.Lock()
	defer queue.L.Unlock()

	executingCount--
	queue.Signal()
}

func enqueue(queueID *int) (doWait bool) {
	queue.L.Lock()
	defer queue.L.Unlock()

	// ensure queueID and executingID
	if *queueID == 0 {
		lastQueueID++
		*queueID = lastQueueID
		if executingID == 0 {
			executingID = *queueID
		}
	}

	// handle uncalled for busy: when this is the only request
	if doWait = executingCount < 2; doWait {
		return // database busy but not executing our queries: unqueued retry
	}

	var timer *time.Timer
	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	// if join at end of queue, trigger retry for the first item
	if *queueID > executingID {
		queue.Signal() // let first item in queue retry
	} else if executingCount <= lastQueueID-executingID+1 {
		doWait = true
		return // if no unqueued requests, retry after short wait
	}

	// wait in queue until there is reason to retry
	for {

		// head of the line: cap wait at 5 ms
		var ch chan struct{}
		if *queueID == executingID {
			ch := make(chan struct{})
			timer = time.NewTimer(sqMaxWait)
			go func() {
				defer parl.Recover(parl.Annotation(), nil, parl.Infallible)

				select {
				case <-timer.C:
					queue.Signal()
				case <-ch:
					timer.Stop()
				}
			}()
		}
		queue.Wait()
		if ch != nil {
			close(ch)
		}

		// head of the line: retry every time
		if executingID == *queueID {
			break
		}
	}

	return
}

func dequeue(queueID *int) {
	if *queueID == 0 {
		return // query was not enqueued
	}
	queue.L.Lock()
	defer queue.L.Unlock()

	if *queueID == executingID {
		executingID++
	}
}
