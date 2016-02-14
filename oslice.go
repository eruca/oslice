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

func (o *OSlice) sort() {
	if !o.isSorted {
		sort.Sort(o)
		o.isSorted = true
	}
}

func (o *OSlice) shrink() {
	src := o.buf.Bytes()
	dst := make([]byte, len(src))

	copy(dst, src)
	o.buf = bytes.NewBuffer(dst)
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
	i, ok := o.search(text)
	if ok {
		return o.idList[i]
	}

	id = o.Append(text)
	copy(o.idList[i+1:], o.idList[i:])
	o.idList[i] = id

	return id
}

func (o *OSlice) search(text []byte) (i int, found bool) {
	var rg *region
	var toSearch []byte
	var id int

	data := o.buf.Bytes()
	i = sort.Search(len(o.regionList), func(i int) bool {
		id = int(o.idList[i])
		rg = &o.regionList[id]

		toSearch = data[int(rg.begin):int(rg.end)]

		return bytes.Compare(toSearch, text) >= 0
	})

	if i < len(o.regionList) && bytes.Equal(toSearch, text) {
		return i, true
	}

	return i, false
}

func (o *OSlice) Query(regionID int) []byte {
	rg := o.regionList[regionID]

	return o.buf.Bytes()[int(rg.begin):int(rg.end)]
}
