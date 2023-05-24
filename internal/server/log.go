package server

import (
	"fmt"
	"sync"
)

/*

sync
뮤텍스(동기화 객체 )는 읽기와 쓰기가 얼마나 이루어지는지 추적할 수 있다
https://mingrammer.com/gobyexample/mutexes/
*/

var ErrOffsetNotFound = fmt.Errorf("offset not found")

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}
type Log struct {
	mu      sync.Mutex
	records []Record
}

func NewLog() *Log {
	return &Log{}
}

// records 배열에 요소를 추가
func (l *Log) Append(record Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	record.Offset = uint64(len(l.records))
	l.records = append(l.records, record)
	return record.Offset, nil
}

// Read 는 records 배열의 요소를 읽어 반환
func (l *Log) Read(offset uint64) (Record, error) {
	l.mu.Lock()         //뮤텍스 잠금
	defer l.mu.Unlock() //뮤텍스 잠금 해제
	//offset 값이 records  배열의 길이가 offset 보다 같거나 크면 에러
	if offset >= uint64(len(l.records)) {
		return Record{}, ErrOffsetNotFound
	}
	return l.records[offset], nil
}
