package oslice

import (
	"encoding/binary"
	"log"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkAppend(b *testing.B) {
	o := OSlice{}
	o.Init(false)

	word := []byte("中国")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o.Append(word)
	}
}

func BenchmarkMapInsert(b *testing.B) {
	m := make(map[string]bool)

	word := "中国"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m[word] = true
	}
}

func BenchmarkSearch(b *testing.B) {
	var o OSlice
	o.Init(false)

	for i := uint64(0); i < 588888; i++ {
		size := binary.Size(i)
		data := make([]byte, size)
		binary.LittleEndian.PutUint64(data, i)
		o.Append(data)
	}

	o.SortIfNot()

	size := binary.Size(uint64(10000))
	data := make([]byte, size)
	binary.LittleEndian.PutUint64(data, 23)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		o.Search(data)
	}
}

func BenchmarkMapSearch(b *testing.B) {
	m := map[string]bool{"中国": true}

	word := "中国"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = m[word]
	}
}

func TestAppend(t *testing.T) {
	var o OSlice
	o.Init(false)

	o.Append([]byte("a"))
	o.Append([]byte("c"))
	o.Append([]byte("b"))

	assert.Equal(t, o.BufLen(), 3)
	assert.Equal(t, o.buf, []byte{'a', 'c', 'b'})
	assert.Equal(t, o.regionList, []Region{{0, 1}, {1, 2}, {2, 3}})
	assert.Equal(t, o.idList, []RegionID{0, 1, 2})

	assert.Equal(t, o.SortIfNot(), true)
	log.Printf("len:%d, cap:%d", o.BufLen(), o.BufCap())
	assert.True(t, o.Shrink(0))
	assert.Equal(t, o.idList, []RegionID{0, 2, 1})
}

func TestAppend2(t *testing.T) {
	var o OSlice
	o.Init(true)

	o.Append([]byte("a"))
	o.Append([]byte("c"))
	o.Append([]byte("b"))

	assert.Equal(t, o.BufLen(), 3)
	assert.Equal(t, o.buf, []byte{'a', 'c', 'b'})
	assert.Equal(t, o.regionList, []Region{{0, 1}, {1, 2}, {2, 3}})
	assert.Equal(t, o.idList, []RegionID{0, 2, 1})

	assert.Equal(t, o.SortIfNot(), false)
	assert.Equal(t, o.idList, []RegionID{0, 2, 1})
}

func TestParrel(t *testing.T) {
	var o OSlice
	o.Init(false)

	strs := []string{"aa", "cc", "dd", "ee", "bb", "ff", "gg", "hh", "ii"}

	for i := 0; i < 3; i++ {
		o.Append([]byte(strs[i]))
	}

	var wg sync.WaitGroup

	assert.Equal(t, o.SortIfNot(), true)
	for i := 3; i < len(strs); i++ {
		wg.Add(1)
		go func(data []byte) {
			o.Append(data)
			wg.Done()
		}([]byte(strs[i]))
	}

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			assert.Equal(t, o.ToByte(RegionID(i)), []byte(strs[i]))
			wg.Done()
		}(i)
	}

	wg.Wait()
	assert.False(t, o.SortIfNot())

	for i := 0; i < len(strs)-1; i++ {
		// log.Printf("%q - %q", o.Id(o.idList[i]), o.Id(o.idList[i+1]))
		assert.Equal(t, o.Less(i, i+1), true)
	}
}

func TestParralAppend(t *testing.T) {
	var o OSlice
	o.Init(false)

	strs := []string{"aa", "cc", "dd", "ee", "bb", "ff"}

	var wg sync.WaitGroup

	wg.Add(len(strs))
	for _, str := range strs {
		go func(data []byte) {
			o.Append(data)
			wg.Done()
		}([]byte(str))
	}

	wg.Wait()

	o.SortIfNot()

	assert.Equal(t, o.Len(), len(strs))
	assert.Equal(t, o.regionList, []Region{{0, 2}, {2, 4}, {4, 6}, {6, 8}, {8, 10}, {10, 12}})
	for i := 0; i < len(strs)-1; i++ {
		assert.Equal(t, o.Less(i, i+1), true)
	}
}

func TestParrelAppend2(t *testing.T) {
	var o OSlice
	o.Init(true)

	strs := []string{"aa", "cc", "dd", "ee", "bb", "ff"}

	var wg sync.WaitGroup

	wg.Add(len(strs))
	for _, str := range strs {
		go func(data []byte) {
			o.Append(data)
			wg.Done()
		}([]byte(str))
	}

	wg.Wait()
	// log.Printf("%q", o.buf)

	assert.Equal(t, o.SortIfNot(), false)
	assert.Equal(t, o.Len(), len(strs))
	assert.Equal(t, o.regionList, []Region{{0, 2}, {2, 4}, {4, 6}, {6, 8}, {8, 10}, {10, 12}})
	for i := 0; i < len(strs)-1; i++ {
		// log.Printf("%q - %q", o.Id(o.idList[i]), o.Id(o.idList[i+1]))
		assert.Equal(t, o.Less(i, i+1), true)
	}
}
