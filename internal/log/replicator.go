package log

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	api "github.com/gwiyeomgo/proglog/api/v1"
)

// 디스커버드 서버의 요청과 로그 복제
// 다른 서버를 발견하면 서로의 데이터를 복제
// 서버가 클러스터에 조인할 때 복제하거나, 떠날 때 복제를 끝낼 컴포넌트 필요
// 복제 컴포넌트가 발견한 서버에서 소비하고,복사본을 로컬 서버에 생산
// 풀 기반 (pull) 복제 ?
//소비하는 측에서 데이터 소스에 소비할 새로운 데이터가 있는지 주기적으로 확인
//풀 기반 시스템??

// Replicator 는 grpc 클라이언트를 이용해 다른 서버에 연결
type Replicator struct {
	DialOptions []grpc.DialOption //클라이언트가 서버 인증을 할 수 있게 설정
	LocalServer api.LogClient

	logger *zap.Logger

	mu      sync.Mutex
	servers map[string]chan struct{}
	closed  bool
	close   chan struct{}
}

func (r *Replicator) Join(name, addr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}

	if _, ok := r.servers[name]; ok {
		// 이미 복제증이니 건너뛰다.
		return nil
	}
	r.servers[name] = make(chan struct{}) //채널 생성

	go r.replicate(addr, r.servers[name])

	return nil
}

// 복제
func (r *Replicator) replicate(addr string, leave chan struct{}) {
	cc, err := grpc.Dial(addr, r.DialOptions...)
	if err != nil {
		r.logError(err, "failed to dial", addr)
		return
	}
	defer cc.Close()

	client := api.NewLogClient(cc)

	ctx := context.Background()
	stream, err := client.ConsumeStream(ctx,
		&api.ConsumeRequest{
			Offset: 0,
		},
	)
	if err != nil {
		r.logError(err, "failed to consume", addr)
		return
	}
	//클라이언트를 생성하고 스트림을 열어서
	//서버의 모든 로그를 소비한다
	//서버의 로그를 스트림을 통해서 반복 소비함
	//로컬 서버에 생산하여 복사본을 저장한다
	records := make(chan *api.Record)
	go func() {
		for {
			recv, err := stream.Recv()
			if err != nil {
				r.logError(err, "failed to receive", addr)
				return
			}
			records <- recv.Record
		}
	}()

	for {
		select {
		case <-r.close:
			return
		case <-leave:
			return
		case record := <-records:
			_, err = r.LocalServer.Produce(ctx,
				&api.ProduceRequest{
					Record: record,
				},
			)
			if err != nil {
				r.logError(err, "failed to produce", addr)
				return
			}
		}
	}
}

func (r *Replicator) Leave(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()
	if _, ok := r.servers[name]; !ok {
		return nil
	}
	close(r.servers[name])
	delete(r.servers, name)
	return nil
}

func (r *Replicator) init() {
	if r.logger == nil {
		r.logger = zap.L().Named("replicator")
	}
	if r.servers == nil {
		r.servers = make(map[string]chan struct{})
	}
	if r.close == nil {
		r.close = make(chan struct{})
	}
}

func (r *Replicator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.init()

	if r.closed {
		return nil
	}
	r.closed = true
	close(r.close)
	return nil
}

func (r *Replicator) logError(err error, msg, addr string) {
	r.logger.Error(
		msg,
		zap.String("addr", addr),
		zap.Error(err),
	)
}
