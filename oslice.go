package oslice

import (
	"bytes"
	"sort"
	"sync"
	"sync/atomic"
)

var (
	_rate = 0.9
)

func SetShrinkRate(rate float64) {
	_rate = rate
}

func GetShrinkRate() float64 {
	return _rate
}

type RegionID int

type Region struct {
	start int
	end   int
}

type OSlice struct {
	buf        []byte
	regionList []Region
	idList     []RegionID
	isSorted   uint64
	sync.RWMutex
}

// 单线程中执行
func (o *OSlice) Init(withSorted bool) {
	if withSorted {
		o.isSorted = 1
	}
}

func (o *OSlice) ToByte(id RegionID) []byte {
	o.RLock()
	defer o.RUnlock()

	return o.toByte(id)
}

func (o *OSlice) toByte(id RegionID) []byte {
	return o.buf[o.regionList[int(id)].start:o.regionList[int(id)].end]
}

func (o *OSlice) foundOrInsert(text []byte) (id RegionID) {
	o.RLock()
	pos, found := o.search(text)
	if found {
		o.RUnlock()
		return o.idList[pos]
	}
	o.RUnlock()

	o.Lock()
	if pos, found = o.search(text); found {
		o.Unlock()
		return o.idList[pos]
	}

	begin := o.BufLen()
	o.buf = append(o.buf, text...)
	length := len(text)

	id = RegionID(len(o.regionList))
	o.regionList = append(o.regionList, Region{
		start: begin,
		end:   begin + length,
	})

	o.idList = append(o.idList, 0)
	copy(o.idList[pos+1:], o.idList[pos:])
	o.idList[pos] = id
	o.Unlock()

	return id
}

func (o *OSlice) Search(text []byte) bool {
	o.RLock()
	_, found := o.search(text)
	o.RUnlock()

	return found
}

func (o *OSlice) search(text []byte) (pos int, found bool) {
	if len(o.idList) == 0 {
		return 0, false
	}

	first, last := 0, len(o.idList)-1

	cmp := bytes.Compare(text, o.toByte(o.idList[0]))
	if cmp == 0 {
		return 0, true
	} else if cmp < 0 {
		return 0, false
	}

	cmp = bytes.Compare(text, o.toByte(o.idList[last]))
	if cmp == 0 {
		return last, true
	} else if cmp > 0 {
		return last + 1, false
	}

	mid := 0
	for last-first > 1 {
		mid = (last + first) / 2
		cmp = bytes.Compare(text, o.toByte(o.idList[mid]))
		if cmp == 0 {
			return mid, true
		} else if cmp < 0 {
			last = mid
		} else {
			first = mid
		}
	}

	return last, false
}

func (o *OSlice) Shrink(reserve int) bool {
	o.Lock()
	defer o.Unlock()

	if float64(o.BufLen()/o.BufCap()) > _rate {
		return false
	}

	buf := make([]byte, len(o.buf)+reserve)
	copy(buf, o.buf)
	o.buf = buf

	return true
}

func (o *OSlice) Append(term []byte) (id RegionID) {
	if atomic.LoadUint64(&o.isSorted) == 0 {
		return o.append(term)
	}

	return o.foundOrInsert(term)
}

func (o *OSlice) append(term []byte) (id RegionID) {
	o.Lock()

	begin := o.BufLen()
	o.buf = append(o.buf, term...)
	length := len(term)

	id = RegionID(len(o.regionList))
	o.regionList = append(o.regionList, Region{
		start: begin,
		end:   begin + length,
	})

	o.idList = append(o.idList, id)
	o.Unlock()

	return id
}

func (o *OSlice) BufLen() int {
	return len(o.buf)
}

func (o *OSlice) BufCap() int {
	return cap(o.buf)
}

func (o *OSlice) Len() int {
	return len(o.idList)
}

func (o *OSlice) Less(i, j int) bool {
	return bytes.Compare(o.toByte(o.idList[i]), o.toByte(o.idList[j])) < 0
}

func (o *OSlice) Swap(i, j int) {
	o.idList[i], o.idList[j] = o.idList[j], o.idList[i]
}

func (o *OSlice) SortIfNot() bool {
	if atomic.CompareAndSwapUint64(&o.isSorted, 0, 1) {
		sort.Sort(o)
		return true
	}

	return false
}
