/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package pmaps

import (
	"fmt"
	"testing"
	"unsafe"

	"github.com/haraldrudell/parl/pmaps/swissmap"
)

// a nil map value is of unknown implementation
// and has no internal structure
func TestGoMapSizeNil(t *testing.T) {
	//t.Error("LoggingOn")

	if swissmap.IsBucketMap() {
		t.Skip("older Go featuring legacy bucket map")
	}

	const (
		testLabel = "nil map"
		// the hash table size of a nil map
		expLoadFactor = float32(0)
		// the allocated entries for a nil map
		expAllocatedEntries = 0
	)

	var (
		actLoadFactor       float32
		actAllocatedEntries int
	)

	// map has no methods
	//	- 5 operations: Get Put Delete Length Range
	//	- a map can be created using make, new or composite literal
	//	- make can have optional size argument
	//	- new returns pointer to ununsable uninitialized value
	//	- initializes: testLabel m swissMap actDirLen actGroupCount actGroupStrings
	var m map[int]int

	// load factor should match
	actLoadFactor, actAllocatedEntries = GoMapSize(m)
	// nil map: loadFactor: 0 allocated entries: 0
	t.Logf("%s: loadFactor: %v allocated entries: %d",
		testLabel, actLoadFactor, actAllocatedEntries,
	)
	if actLoadFactor != expLoadFactor {
		t.Errorf("FAIL %s loadFactor %v exp %v",
			testLabel, actLoadFactor, expLoadFactor,
		)
	}

	// allocs should match
	if actAllocatedEntries != expAllocatedEntries {
		t.Errorf("FAIL %s allocatedEntries %d exp %d",
			testLabel, actAllocatedEntries, expAllocatedEntries,
		)
	}
}

// make without size argument creates a
// swiss map of zero entry-allocations
//   - make states: The size may be omitted, in which case a
//     small starting size is allocated
//   - swiss map uses zero
func TestGoMapSizeMakeNoSize(t *testing.T) {
	//t.Error("LoggingOn")

	if swissmap.IsBucketMap() {
		t.Skip("older Go featuring legacy bucket map")
	}

	const (
		expLoadFactor = float32(0)
		// the allocated entries for the map
		expAllocatedEntries = 0
	)

	var (
		tester              *swissMapTester
		actLoadFactor       float32
		actAllocatedEntries int
	)

	// map has no methods
	//	- 5 operations: Get Put Delete Length Range
	//	- a map can be created using make, new or composite literal
	//	- make can have optional size argument
	//	- new returns pointer to ununsable uninitialized value
	//	- initializes: testLabel m swissMap actDirLen actGroupCount actGroupStrings
	var m map[int]int
	tester = newSwissMapTester(&m, t)

	// unused map created by one-argument make
	//	- make map: The size may be omitted, in which case a small starting size is allocated
	//	- swiss map is created with 0 groups
	tester.reset("make omitted size", make(map[int]int))

	// “make omitted size: loadFactor 0 dirLen: 0 len: 0 allocs: 0
	// tables: 0 groups: 0 Directory: 0 Groups: ‘’ Tables: 0”
	t.Logf("%s: loadFactor %v dirLen: %d len: %d allocs: %d tables: %d groups: %d"+
		" Directory: %d Groups: ‘%s’ Tables: %d",
		tester.label,
		tester.actLoadFactor,
		tester.actDirLen, len(m), tester.actEntryAllocationCount,
		tester.actTableCount, tester.actGroupCount,
		tester.actDirectory,
		tester.groupStrings(),
		tester.actTables,
	)

	// make omitted size: swissMap_used_0_dirlen_0_depth_0
	// _shift_0_dirptr_0x0_seq_0_seed_cb0b283adbc8b74
	t.Logf("%s: %s", tester.label, tester.actSwissMap.String())

	// loadFactor should be correct
	actLoadFactor, actAllocatedEntries = GoMapSize(m)
	// make omitted size: loadFactor: 0 allocated: 0
	t.Logf("%s: loadFactor: %v allocated: %d",
		tester.label, actLoadFactor, actAllocatedEntries,
	)
	if actLoadFactor != expLoadFactor {
		t.Errorf("FAIL %s loadFactor %v exp %v",
			tester.label, actLoadFactor, expLoadFactor,
		)
	}

	// allocatedEntries should be correct
	if actAllocatedEntries != expAllocatedEntries {
		t.Errorf("FAIL %s allocatedEntries %d exp %d",
			tester.label, actAllocatedEntries, expAllocatedEntries,
		)
	}
}

// make to specific size cause directory to be created for size 9+
//   - swiss map group size is 8
//   - for 9, there are two groups and directory must be created
//   - allocated entries are 16, multiple of 8
//   - bucket size is 1
func TestGoMapSizeMake9(t *testing.T) {
	//t.Error("LoggingOn")

	if swissmap.IsBucketMap() {
		t.Skip("older Go featuring legacy bucket map")
	}

	const (
		// the requested dimension for the map
		//	 values larger than 9 causes dirctory allocation
		makeArg       = 9
		expLoadFactor = float32(0)
		// the allocated entries for the map
		expAllocatedEntries = 16
	)

	var (
		tester                             *swissMapTester
		actLoadFactor                      float32
		actBucketSize, actAllocatedEntries int
	)

	// map has no methods
	//	- 5 operations: Get Put Delete Length Range
	//	- a map can be created using make, new or composite literal
	//	- make can have optional size argument
	//	- new returns pointer to ununsable uninitialized value
	//	- initializes: testLabel m swissMap actDirLen actGroupCount actGroupStrings
	var m map[int]int
	tester = newSwissMapTester(&m, t)

	// unused map created by one-argument make
	//	- make map: The size may be omitted, in which case a small starting size is allocated
	//	- swiss map is created with 0 groups
	tester.reset(
		fmt.Sprintf("make size %d", makeArg),
		make(map[int]int, makeArg),
	)

	// “make size 9: loadFactor 0 dirLen: 1 len: 0 allocs: 16
	// tables: 1 groups: 2 Directory: 1 Groups: ‘group0: alloc: 2 0x14000102330’ Tables: 1”
	t.Logf("%s: loadFactor %v dirLen: %d len: %d allocs: %d tables: %d groups: %d"+
		" Directory: %d Groups: ‘%s’ Tables: %d",
		tester.label,
		tester.actLoadFactor,
		tester.actDirLen, len(m), tester.actEntryAllocationCount,
		tester.actTableCount, tester.actGroupCount,
		tester.actDirectory,
		tester.groupStrings(),
		tester.actTables,
	)

	// “make size 9: swissMap_used_0_dirlen_1_depth_0_shift_64
	// _dirptr_0x140000560d8_seq_0_seed_f36dcc6399aa838e”
	t.Logf("%s: %s", tester.label, tester.actSwissMap.String())

	// load factor should be correct
	actLoadFactor, actAllocatedEntries = GoMapSize(m)
	// make size 9: buckets: 0 allocated: 16
	t.Logf("%s: buckets: %d allocated: %d",
		tester.label, actBucketSize, actAllocatedEntries,
	)
	if actLoadFactor != expLoadFactor {
		t.Errorf("FAIL %s loadFactor %v exp %v",
			tester.label, actLoadFactor, expLoadFactor,
		)
	}

	// allocatedEntries should be correct
	if actAllocatedEntries != expAllocatedEntries {
		t.Errorf("FAIL %s allocatedEntries %d exp %d",
			tester.label, actAllocatedEntries, expAllocatedEntries,
		)
	}
}

// map composite literal
//   - allocated entries zero
//   - bucket size is 1
func TestGoMapSizeLiteral(t *testing.T) {
	//t.Error("LoggingOn")

	if swissmap.IsBucketMap() {
		t.Skip("older Go featuring legacy bucket map")
	}

	const (
		// the hash table size of the map
		expLoadFactor = float32(0)
		// the allocated entries for the map
		expAllocatedEntries = 0
	)

	var (
		tester              *swissMapTester
		actLoadFactor       float32
		actAllocatedEntries int
	)

	// map has no methods
	//	- 5 operations: Get Put Delete Length Range
	//	- a map can be created using make, new or composite literal
	//	- make can have optional size argument
	//	- new returns pointer to ununsable uninitialized value
	//	- initializes: testLabel m swissMap actDirLen actGroupCount actGroupStrings
	var m map[int]int
	tester = newSwissMapTester(&m, t)

	// unused map created by one-argument make
	//	- make map: The size may be omitted, in which case a small starting size is allocated
	//	- swiss map is created with 0 groups
	tester.reset("literal", map[int]int{})

	// “literal: loadFactor 0 dirLen: 0 len: 0 allocs: 0
	// tables: 0 groups: 0 Directory: 0 Groups: ‘’ Tables: 0”
	t.Logf("%s: loadFactor %v dirLen: %d len: %d allocs: %d tables: %d groups: %d"+
		" Directory: %d Groups: ‘%s’ Tables: %d",
		tester.label,
		tester.actLoadFactor,
		tester.actDirLen, len(m), tester.actEntryAllocationCount,
		tester.actTableCount, tester.actGroupCount,
		tester.actDirectory,
		tester.groupStrings(),
		tester.actTables,
	)

	// “literal: swissMap_used_0_dirlen_0_depth_0_shift_0
	// _dirptr_0x0_seq_0_seed_945a6890812c4932”
	t.Logf("%s: %s", tester.label, tester.actSwissMap.String())

	// load factor should be correct
	actLoadFactor, actAllocatedEntries = GoMapSize(m)
	// literal: buckets: 1 allocated: 0
	t.Logf("%s: loadFactor: %v allocated: %d",
		tester.label, actLoadFactor, actAllocatedEntries,
	)
	if actLoadFactor != expLoadFactor {
		t.Errorf("FAIL %s loadFactor %v exp %v",
			tester.label, actLoadFactor, expLoadFactor,
		)
	}

	// allocatedEntries should be correct
	if actAllocatedEntries != expAllocatedEntries {
		t.Errorf("FAIL %s allocatedEntries %d exp %d",
			tester.label, actAllocatedEntries, expAllocatedEntries,
		)
	}
}

// make map and add 10 Ki mappings
//   - bucket list size is 16
//   - entries 16,384, significantly larger than 10,240
//   - load factor is 10,240 / 16
func TestGoMapSize10K(t *testing.T) {
	//t.Error("LoggingOn")

	if swissmap.IsBucketMap() {
		t.Skip("older Go featuring legacy bucket map")
	}

	const (
		// the hash table size of the map
		expLoadFactor = float32(0.625)
		// the allocated entries for the map
		expAllocatedEntries = 16384
		// size for making map
		mapLength = 10240
	)

	var (
		tester              *swissMapTester
		actLoadFactor       float32
		actAllocatedEntries int
	)

	// map has no methods
	//	- 5 operations: Get Put Delete Length Range
	//	- a map can be created using make, new or composite literal
	//	- make can have optional size argument
	//	- new returns pointer to ununsable uninitialized value
	//	- initializes: testLabel m swissMap actDirLen actGroupCount actGroupStrings
	var m map[int]int
	tester = newSwissMapTester(&m, t)

	// map with one mapping created by one-argument make
	m = make(map[int]int)
	tester.populateMap(mapLength)
	tester.reset(fmt.Sprintf("map length %d", mapLength), m)

	// “map length 10240: loadFactor 0.625 dirLen: 16 len: 10240 allocs: 16384
	// tables: 16 groups: 2048 Directory: 16
	// Groups: ‘group0: alloc: 128 0x140000a49b0…group15: alloc: 128 0x140000a4a90’
	// Tables: 16”
	t.Logf("%s: loadFactor %v dirLen: %d len: %d allocs: %d tables: %d groups: %d"+
		" Directory: %d Groups: ‘%s’ Tables: %d",
		tester.label,
		tester.actLoadFactor,
		tester.actDirLen, len(m), tester.actEntryAllocationCount,
		tester.actTableCount, tester.actGroupCount,
		tester.actDirectory,
		tester.groupStrings(),
		tester.actTables,
	)

	// “map length 10240: swissMap_used_10,240_dirlen_16_depth_4_shift_60
	// _dirptr_0x1400012e200_seq_0_seed_3cda70bbc505ca25”
	t.Logf("%s: %s", tester.label, tester.actSwissMap.String())

	// load factor should match
	actLoadFactor, actAllocatedEntries = GoMapSize(m)
	// map length 10240: load factor: 0.625 allocated: 16384
	t.Logf("%s: load factor: %v allocated: %d",
		tester.label, actLoadFactor, actAllocatedEntries,
	)
	if actLoadFactor != expLoadFactor {
		t.Errorf("FAIL %s: loadFactor: %v exp %v", tester.label, actLoadFactor, expLoadFactor)
	}

	// alloc entries should match
	if actAllocatedEntries != expAllocatedEntries {
		t.Errorf("FAIL %s allocatedEntries: %d exp %d",
			tester.label, actAllocatedEntries, expAllocatedEntries,
		)
	}
}

// swissMapTester provides test functions for go1.24 swiss map
type swissMapTester struct {
	mp                                                               *map[int]int
	label                                                            string
	actSwissMap                                                      *swissmap.SwissMap
	actLoadFactor                                                    float32
	actDirLen, actEntryAllocationCount, actTableCount, actGroupCount int
	actDirectory, actTables                                          int
	actGroupStrings                                                  []string
	t                                                                *testing.T
}

// newSwissMapTester returns a tester for go1.24 swiss map
func newSwissMapTester(
	mp *map[int]int,
	t *testing.T,
) (s *swissMapTester) {
	return &swissMapTester{
		mp: mp,
		t:  t,
	}
}

// reset starts test name label based on map m0
func (s *swissMapTester) reset(label string, m0 map[int]int) {
	var t = s.t
	s.label = label
	*s.mp = m0
	s.actSwissMap = swissmap.GetSwissMap(m0)
	if s.actSwissMap == nil {
		t.Fatalf("FAIL %s: GetSwissMap nil", label)
	}
	s.actLoadFactor, s.actDirLen, s.actEntryAllocationCount, s.actTableCount, s.actGroupCount = s.actSwissMap.Structure()
	for table := range s.actSwissMap.Directory {
		_ = table
		s.actDirectory++
	}
	for table := range s.actSwissMap.Tables {
		_ = table
		s.actTables++
	}
	s.actGroupStrings = nil
	var i int
	for group := range s.actSwissMap.Groups {
		var gp = uintptr(unsafe.Pointer(group))
		var g = fmt.Sprintf("group%d: alloc: %d 0x%x",
			i, group.LengthMask+1, gp,
		)
		s.actGroupStrings = append(s.actGroupStrings, g)
		i++
	}
}

// populateMap creates n mappings in m
//   - invoke prior to reset
func (s *swissMapTester) populateMap(n int) {
	// key: 0…n-1, value: n…2n-1
	for key := range n {
		var value = key + n
		(*s.mp)[key] = value
	}
}

// groupStrings returns abbreviated representation of actGroupStrings
func (s *swissMapTester) groupStrings() (s2 string) {
	switch length := len(s.actGroupStrings); length {
	case 0:
	case 1:
		s2 = s.actGroupStrings[0]
	case 2:
		s2 = fmt.Sprintf("%s, %s",
			s.actGroupStrings[0],
			s.actGroupStrings[1],
		)
	default:
		s2 = fmt.Sprintf("%s…%s",
			s.actGroupStrings[0],
			s.actGroupStrings[length-1],
		)
	}

	return
}
