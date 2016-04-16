package oslice

import (
	"bytes"
	"log"
	"sort"
)

type region struct {
	begin uint32
	end   uint32
}

type OSlice struct {
	buf        *bytes.Buffer
	regionList []region
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
	o.buf = new(bytes.Buffer)
}

func (o *OSlice) Len() int {
	if len(o.regionList) != len(o.idList) {
		panic("regionList is not sync with idList")
	}
	return len(o.regionList)
}

func (o *OSlice) Less(i, j int) bool {
	data := o.buf.Bytes()

	rgI := o.regionList[i]
	rgJ := o.regionList[j]

	return bytes.Compare(data[int(rgI.begin):int(rgI.end)],
		data[int(rgJ.begin):int(rgJ.end)]) < 0
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

func (o *OSlice) Shrink() {
	length := o.buf.Len()
	src := o.buf.Bytes()[:length:length]
	o.buf = bytes.NewBuffer(src)
}

func (o *OSlice) Append(words []byte) (id RegionID) {
	begin := o.buf.Len()

	n, err := o.buf.Write(words)
	if err != nil {
		log.Panicln("bytes.Buffer write []byte failed: ", err)
	}

	rg := region{
		begin: uint32(begin),
		end:   uint32(begin + n),
	}

	id = RegionID(len(o.regionList))

	o.regionList = append(o.regionList, rg)
	o.idList = append(o.idList, id)

	return id
}

func (o *OSlice) Search(text []byte) bool {
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
// 如果found是true, 则表示查找到位置 i
// 如果found是false,则表示为插入位置 i
func (o *OSlice) search(text []byte) (i int, found bool) {
	data := o.buf.Bytes()
	first, last := 0, len(o.idList)-1

	begin, end := int(o.regionList[o.idList[0]].begin), int(o.regionList[o.idList[0]].end)
	cmp := bytes.Compare(text, data[begin:end])
	if cmp == 0 {
		return 0, true
	} else if cmp < 0 {
		return 0, false
	}

	begin, end = int(o.regionList[o.idList[last]].begin), int(o.regionList[o.idList[last]].end)
	cmp = bytes.Compare(text, data[begin:end])
	if cmp == 0 {
		return len(o.idList) - 1, true
	} else if cmp > 0 {
		return len(o.idList), false
	}

	mid := 0
	for last-first > 1 {
		mid = (last + first) / 2
		begin = int(o.regionList[o.idList[mid]].begin)
		end = int(o.regionList[o.idList[mid]].end)

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

func (o *OSlice) Query(regionID RegionID) []byte {
	rg := o.regionList[int(regionID)]

	return o.buf.Bytes()[int(rg.begin):int(rg.end)]
}
