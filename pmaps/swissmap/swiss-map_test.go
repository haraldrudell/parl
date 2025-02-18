/*
© 2023–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package swissmap

import (
	"strings"
	"testing"
	"unsafe"
)

func TestSwissMap(t *testing.T) {
	//t.Error("Logging on")

	if IsBucketMap() {
		t.Skip("older Go featuring legacy bucket map")
	}

	type swissMapNil bool
	const (
		swissMapNilYes swissMapNil = true
		swissMapNilNo  swissMapNil = false
	)
	type testFixture struct {
		label           string
		m               map[int]int
		expSwisssMapNil swissMapNil
		expString       string
		expLoadFactor   float32
		expDirLen       int
		expAllocation   int
		expTableCount   int
		expGroupCount   int
		expDirectory    int
		expTables       int
		expGroups       int
	}
	const (
		size10Ki = 10240
	)
	var (
		tests = []testFixture{{
			label:           "nil map",
			m:               nil,
			expSwisssMapNil: swissMapNilYes,
		}, {
			label:           "make omitted size",
			m:               make(map[int]int),
			expSwisssMapNil: swissMapNilNo,
			expString:       "swissMap_used_0_dirlen_0_depth_0_shift_0_dirptr_0x_seq_0_seed_",
			expLoadFactor:   float32(0),
			expDirLen:       0,
			expAllocation:   0,
			expTableCount:   0,
			expGroupCount:   0,
			expDirectory:    0,
			expTables:       0,
			expGroups:       0,
		}, {
			label:           "make size 9",
			m:               make(map[int]int, 9),
			expSwisssMapNil: swissMapNilNo,
			expString:       "swissMap_used_0_dirlen_1_depth_0_shift_64_dirptr_0x_seq_0_seed_",
			expLoadFactor:   float32(0),
			expDirLen:       1,
			expAllocation:   16,
			expTableCount:   1,
			expGroupCount:   2,
			expDirectory:    1,
			expTables:       2,
			expGroups:       3,
		}, {
			label:           "literal",
			m:               map[int]int{},
			expSwisssMapNil: swissMapNilNo,
			expString:       "swissMap_used_0_dirlen_0_depth_0_shift_0_dirptr_0x_seq_0_seed_",
			expLoadFactor:   float32(0),
			expDirLen:       0,
			expAllocation:   0,
			expTableCount:   0,
			expGroupCount:   0,
			expDirectory:    3,
			expTables:       3,
			expGroups:       3,
		}, {
			label: "10Ki",
			m: func() (m map[int]int) {
				m = make(map[int]int)
				for i := range size10Ki {
					m[i] = i + size10Ki
				}
				return
			}(),
			expSwisssMapNil: swissMapNilNo,
			expString:       "swissMap_used_10,240_dirlen_16_depth_4_shift_60_dirptr_0x_seq_0_seed_",
			expLoadFactor:   float32(0.625),
			expDirLen:       16,
			expAllocation:   16384,
			expTableCount:   16,
			expGroupCount:   2048,
			expDirectory:    19,
			expTables:       35,
			expGroups:       51,
		}}
	)

	var (
		actSwissMap                                                                     *SwissMap
		actDirPtr                                                                       uintptr
		actLoadFactor                                                                   float32
		actDirLen, actEntryAllocationCount, actGroupCount, actIterations, actTableCount int
		actString                                                                       string
	)

	// sizeof map: 8
	// the map type has no methods. There are 5 operations:
	//	- read: value := m[key]; value, ok := m[key]
	//	- assign: m[key]=value
	//	- delete: delete(m, key)
	//	- length: len(m)
	//	- range: for key, value := range m {
	//	-
	//	- the debugger cannot step into those operations
	//	- a map is created by make, new or literal
	//	- — make takes zero or one additional arguments
	//	- — new returns an unitialized, unusable map pointer
	//	- a map value is a pointer to an internal structure
	t.Logf("sizeof map: %d", unsafe.Sizeof(map[int]int{}))

	for _, mapTest := range tests {

		// swissMap should match
		actSwissMap = GetSwissMap(mapTest.m)
		if actSwissMap == nil {
			if mapTest.expSwisssMapNil == swissMapNilNo {
				t.Errorf("FAIL %s swissMap nil", mapTest.label)
			}
			continue
		}
		// swissMap not nil

		// Seq() String() Structure()
		var _ *SwissMap

		// String() should match
		actString = actSwissMap.String()
		// make omitted size: swissMap_used_0_dirlen_0_depth_0_shift_0
		// _dirptr_0x0_seq_0_seed_b1b2084ea6f2be16
		// make size 9: swissMap_used_0_dirlen_1_depth_0_shift_64
		// _dirptr_0x1400010e0c8_seq_0_seed_37b17311d6ac5efa
		// literal: swissMap_used_0_dirlen_0_depth_0_shift_0
		// _dirptr_0x0_seq_0_seed_c92b3ab7273f9ecd
		// 10Ki: swissMap_used_10,240_dirlen_16_depth_4_shift_60
		// _dirptr_0x1400012e200_seq_0_seed_914428e1209306ff
		t.Logf("%s: %s", mapTest.label, actString)
		if !compareString(actString, mapTest.expString, t) {
			t.Errorf("FAIL %s swissMap String\n%q exp\n%q", mapTest.label, actString, mapTest.expString)
		}

		// Structure() dirlen should match
		actLoadFactor, actDirLen, actEntryAllocationCount, actTableCount, actGroupCount = actSwissMap.Structure()
		actDirPtr = uintptr(actSwissMap.DirPtr)
		// make omitted size: loadFactor 0 dirLen: 0 alloc: 0 table: 0 group: 0 0x0
		// make size 9: loadFactor 0 dirLen: 1 alloc: 16 table: 1 group: 2 0x140000560b0
		// literal: loadFactor 0 dirLen: 0 alloc: 0 table: 0 group: 0 0x0
		// 10Ki: loadFactor 0.625 dirLen: 16 alloc: 16384 table: 16 group: 2048 0x1400012e200
		t.Logf("%s: loadFactor %v dirLen: %d alloc: %d table: %d group: %d 0x%x",
			mapTest.label,
			actLoadFactor, actDirLen, actEntryAllocationCount, actTableCount, actGroupCount,
			actDirPtr,
		)
		if actLoadFactor != mapTest.expLoadFactor {
			t.Errorf("FAIL %s swissMap Load factor %v exp %v", mapTest.label, actLoadFactor, mapTest.expLoadFactor)
		}

		// dirlen should match
		if actDirLen != mapTest.expDirLen {
			t.Errorf("FAIL %s swissMap.DirLen %d exp %d", mapTest.label, actDirLen, mapTest.expDirLen)
		}

		// entry allocation should match
		if actEntryAllocationCount != mapTest.expAllocation {
			t.Errorf("FAIL %s swissMap.Allocations %d exp %d", mapTest.label, actEntryAllocationCount, mapTest.expAllocation)
		}

		// table count should match
		if actTableCount != mapTest.expTableCount {
			t.Errorf("FAIL %s swissMap table count %d exp %d", mapTest.label, actTableCount, mapTest.expTableCount)
		}

		// group count should match
		if actGroupCount != mapTest.expGroupCount {
			t.Errorf("FAIL %s swissMap.GroupCount %d exp %d", mapTest.label, actGroupCount, mapTest.expGroupCount)
		}

		// Directory iterations should match
		for table := range actSwissMap.Directory {
			_ = table
			actIterations++
		}
		if actIterations != mapTest.expDirectory {
			t.Errorf("FAIL %s swissMap.Directory iterations %d exp %d", mapTest.label, actIterations, mapTest.expDirectory)
		}

		// Table iterations should match
		for table := range actSwissMap.Tables {
			_ = table
			actIterations++
		}
		if actIterations != mapTest.expTables {
			t.Errorf("FAIL %s swissMap.Tables iterations %d exp %d", mapTest.label, actIterations, mapTest.expTables)
		}

		// Group iterations should match
		for group := range actSwissMap.Groups {
			_ = group
			actIterations++
		}
		if actIterations != mapTest.expGroups {
			t.Errorf("FAIL %s swissMap.Groups iterations %d exp %d", mapTest.label, actIterations, mapTest.expGroups)
		}
	}
}

// compareString verifies String result filtering non-constant characters
func compareString(actual, expected string, t *testing.T) (isOk bool) {
	var toCompare = actual

	// “…dirptr_0x140000560d0_…” → “…dirptr_0x_…”
	// i is the start of dirptr: “…” i “dirptr_0x_…”
	var i int
	if i = strings.Index(toCompare, dirLabel); i != -1 {
		// nextIndex is after “…dirptr_0x”
		var nextIndex = i + len(dirLabel)
		// i2 is how many characters to delete at nextIndex
		var i2 int
		if i2 = strings.Index(toCompare[nextIndex:], underscore); i2 != -1 {
			// i2
			toCompare = toCompare[:nextIndex] + toCompare[nextIndex+i2:]
		}
	}

	// cut after seed
	if j := strings.Index(toCompare, seedLabel); j != -1 {
		toCompare = toCompare[:j+len(seedLabel)]
	}

	// compare
	isOk = toCompare == expected
	if !isOk {
		t.Logf("compareString:\n%q\n%q\n", toCompare, expected)
	}

	return
}

const (
	// the seed is random, last field, cut
	seedLabel = "seed_"
	// start of dirptr to cut non-constant pointer
	dirLabel = "dirptr_0x"
	// end of dirptr
	underscore = "_"
)
