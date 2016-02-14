package oslice

import (
	// "bytes"
	"reflect"
	"testing"
)

func expect(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expect %v type:%v, but get %v type:%v",
			b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func TestOSlice(t *testing.T) {
	os := New()

	os.Append([]byte("b"))
	os.Append([]byte("a"))
	os.Append([]byte("c"))

	os.sort()
	os.FoundOrInsert([]byte("e"))
	os.FoundOrInsert([]byte("d"))

	expect(t, os.idList, []RegionID{1, 0, 2, 4, 3})
}
