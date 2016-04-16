package oslice

import (
	"bufio"
	// "fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

var o = New()
var isRead = false
var dictFile = "/Users/nick/Lang/Go/cedar-go/testdata/dict.txt"

func expect(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Errorf("expect %v type:%v, but get %v type:%v",
			b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func read() {
	f, err := os.Open(dictFile)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	index := 0
	for scanner.Scan() {
		text := scanner.Text()
		o.Append([]byte(strings.Split(text, " ")[0]))
		index++
	}
	// fmt.Printf("插入 %d 数据, buf: %d\n", index, cap(o.buf.Bytes()))

	o.Sort()
	o.Shrink()

	// fmt.Printf("shrink 后 buf: %d\n", cap(o.buf.Bytes()))
}

func TestRead(t *testing.T) {
	read()
}

func Benchmark_OSlice(b *testing.B) {
	if !isRead {
		read()
	}
	str := []byte("气温计")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// o.Search([]byte("一一对"))
		o.Search(str)
		// o.Search([]byte("龟苓膏"))
	}
}

func Benchmark_SmallOSlice(b *testing.B) {
	o := New()

	str := []byte("气温计")
	o.Append([]byte("一一对"))
	o.Append(str)
	o.Append([]byte("龟苓膏"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		o.Search(str)
	}
}

func Benchmark_Map(b *testing.B) {
	m := make(map[string]int)
	m["一一对"] = 0
	m["气温计"] = 1
	m["龟苓膏"] = 2

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// _, _ = m["一一对"]
		_, _ = m["气温计"]
		// _, _ = m["龟苓膏"]
	}
}

func Benchmark_Query(b *testing.B) {
	if !isRead {
		read()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// o.Query(1)
		o.Query(1000)
		// o.Query(10000)
	}
}
