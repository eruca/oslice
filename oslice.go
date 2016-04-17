package oslice

import (
	"bytes"
	"log"
	"sort"
)

type OSlice struct {
	buf        *bytes.Buffer
	regionList []int
	idList     []RegionID
	isSorted   bool
}

type RegionID int

func New() *OSlice {
	return &OSlice{
		buf: new(bytes.Buffer),
	}
}

// 如果是结构体内部的一个项，而且不是指针，需要用到Init()
func (o *OSlice) Init() {
	if o.buf == nil {
		o.buf = new(bytes.Buffer)
	}
}

func (o *OSlice) Len() int {
	if len(o.regionList) != len(o.idList) {
		panic("regionList is not sync with idList")
	}
	return len(o.regionList)
}

func (o *OSlice) Less(i, j int) bool {
	return bytes.Compare(o.Query(o.idList[i]), o.Query(o.idList[j])) < 0
}

func (o *OSlice) Swap(i, j int) {
	o.idList[i], o.idList[j] = o.idList[j], o.idList[i]
}

func (o *OSlice) Sort() {
	if !o.isSorted {
		sort.Sort(o)
		o.isSorted = true
	}
}

// reserve <= 0 表示不预留空间
// 如果要求预留空间比实际cap-len大，也就直接返回实际可预留空间cap-len
func (o *OSlice) Shrink(reserve int) (reserved int) {
	length := o.buf.Len()
	capacity := o.buf.Cap()

	if reserve < 0 {
		reserve = 0
	}

	if capacity-length > reserve {
		src := o.buf.Bytes()[:length : length+reserve]
		o.buf = bytes.NewBuffer(src)
		return reserve
	}

	return capacity - length
}

func (o *OSlice) Append(words []byte) (id RegionID) {
	begin := o.buf.Len()

	_, err := o.buf.Write(words)
	if err != nil {
		log.Panicln("bytes.Buffer write []byte failed: ", err)
	}

	id = RegionID(len(o.regionList))

	o.regionList = append(o.regionList, begin)
	o.idList = append(o.idList, id)

	return id
}

func (o *OSlice) Search(text []byte) bool {
	if !o.isSorted {
		o.Sort()
	}

	_, ok := o.search(text)
	return ok
}

func (o *OSlice) FoundOrInsert(text []byte) (id RegionID) {
	if !o.isSorted {
		o.Sort()
	}

	i, ok := o.search(text)
	if ok {
		return o.idList[i]
	}

	id = o.Append(text)
	copy(o.idList[i+1:], o.idList[i:])
	o.idList[i] = id

	return id
}

// 返回查找位置及是否找到
// 如果found是true, 则表示idList查找到位置 i
// 如果found是false,则表示为插入位置 i
func (o *OSlice) search(text []byte) (i int, found bool) {
	data := o.buf.Bytes()
	first, last := 0, len(o.idList)-1

	begin, end := o.byteRange(o.idList[0])

	cmp := bytes.Compare(text, data[begin:end])
	if cmp == 0 {
		return 0, true
	} else if cmp < 0 {
		return 0, false
	}

	begin, end = o.byteRange(o.idList[last])

	cmp = bytes.Compare(text, data[begin:end])
	if cmp == 0 {
		return len(o.idList) - 1, true
	} else if cmp > 0 {
		return len(o.idList), false
	}

	mid := 0
	for last-first > 1 {
		mid = (last + first) / 2
		begin, end = o.byteRange(o.idList[mid])

		cmp = bytes.Compare(text, data[begin:end])
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

func (o *OSlice) byteRange(id RegionID) (begin int, end int) {
	begin = o.regionList[int(id)]

	if int(id) == len(o.regionList)-1 {
		end = o.buf.Len()
	} else {
		end = o.regionList[int(id)+1]
	}
	return
}

func (o *OSlice) Query(id RegionID) []byte {
	begin, end := o.byteRange(id)
	return o.buf.Bytes()[begin:end]
}
