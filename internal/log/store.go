package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

// 레코드 크기와 인덱스 항목을 저장할 때의 인코딩을 정의한 것
var (
	enc = binary.BigEndian
)

// 레코드 길이를 저장하는 바이트 개수
const lenWidth = 8

// 저장 파일 : 레코드를 저장하는 파일
type store struct {
	*os.File
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64
}

/*
bufio.Writer는 파일에 데이터를 쓸 때 내부적으로 버퍼링하여 여러 작은 쓰기 작업을 묶어서 한 번에 큰 덩어리로 보내는 것을 가능
*/

func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name()) //주어진 파일 또는 디렉토리의 정보를 조회
	if err != nil {
		return nil, err
	}
	size := uint64(fi.Size())
	return &store{
		File: f,
		size: size,
		buf:  bufio.NewWriter(f), //f에 대한 새로운 버퍼를 생성
	}, nil
}

func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	/*
		이 패턴을 사용하는 이유는 동시에 여러 고루틴이 공유 자원에 접근할 때 경쟁 조건을 방지하고, 데이터의 일관성과 안정성을 보장
	*/
	s.mu.Lock()         //잠금
	defer s.mu.Unlock() //해제
	pos = s.size
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}
	w, err := s.buf.Write(p)
	if err != nil {
		return 0, 0, err
	}
	w += lenWidth
	s.size += uint64(w)
	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// 읽으려는 레코드가 아직 버퍼에 있을 때를 대비해서 우선은 쓰기 버퍼의 내용을 플래시해서 디스크에 쓴다
	if err := s.buf.Flush(); err != nil {
		return nil, err
	}
	size := make([]byte, lenWidth)
	if _, err := s.File.ReadAt(size, int64(pos)); err != nil {
		return nil, err
	}
	b := make([]byte, enc.Uint64(size))
	if _, err := s.File.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return 0, err
	}
	return s.File.ReadAt(p, off)
}

// 파일을 닫기 전 버퍼의 데이터를 파일에 쓴다
func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.buf.Flush(); err != nil {
		return err
	}
	return s.File.Close()
}
