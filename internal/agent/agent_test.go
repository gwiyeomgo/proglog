package agent_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	api "github.com/gwiyeomgo/proglog/api/v1"
	"github.com/gwiyeomgo/proglog/internal/agent"
	"github.com/gwiyeomgo/proglog/internal/config"
)

func TestAgent(t *testing.T) {
	//클라이언트에 전달할 인증서 설정
	serverTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		Server:        true,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)
	//서버들 사이의 인증서 설정을 정의하여 서로 연결하고 복제할 수 있도록
	peerTLSConfig, err := config.SetupTLSConfig(config.TLSConfig{
		CertFile:      config.RootClientCertFile,
		KeyFile:       config.RootClientKeyFile,
		CAFile:        config.CAFile,
		Server:        false,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	var agents []*agent.Agent
	for i := 0; i < 3; i++ {
		ports := dynaport.Get(2)
		bindAddr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
		rpcPort := ports[1]

		dataDir, err := ioutil.TempDir("", "agent-test-log")
		require.NoError(t, err)

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(
				startJoinAddrs,
				agents[0].Config.BindAddr,
			)
		}
		//두번째,세 번째 노드가 첫번째 클러스터에 조인=> 노드가 3개인 클러스터 생성
		agent, err := agent.New(agent.Config{
			NodeName:        fmt.Sprintf("%d", i),
			StartJoinAddrs:  startJoinAddrs,
			BindAddr:        bindAddr,
			RPCPort:         rpcPort,
			DataDir:         dataDir,
			ACLModelFile:    config.ACLModelFile,
			ACLPolicyFile:   config.ACLPolicyFile,
			ServerTLSConfig: serverTLSConfig,
			PeerTLSConfig:   peerTLSConfig,
		})
		require.NoError(t, err)

		agents = append(agents, agent)
	}
	// Shutdown 작동,테스트 데이터 모두 삭제
	defer func() {
		for _, agent := range agents {
			err := agent.Shutdown()
			require.NoError(t, err)
			require.NoError(t,
				os.RemoveAll(agent.Config.DataDir),
			)
		}
	}()
	//노드들이 서로를 찾을 시간 3초
	time.Sleep(3 * time.Second)

	//하나의 노드에서 생산과 소비를 할 수 있는지 확인
	//다른 노드가 레코드를 복제하는지 확인
	//서버 간 복제는 비동기로 이루어진다
	leaderClient := client(t, agents[0], peerTLSConfig)
	produceResponse, err := leaderClient.Produce(
		context.Background(),
		&api.ProduceRequest{
			Record: &api.Record{
				Value: []byte("foo"),
			},
		},
	)
	require.NoError(t, err)
	consumeResponse, err := leaderClient.Consume(
		context.Background(),
		&api.ConsumeRequest{
			Offset: produceResponse.Offset,
		},
	)
	require.NoError(t, err)
	require.Equal(t, consumeResponse.Record.Value, []byte("foo"))

	// wait until replication has finished
	time.Sleep(3 * time.Second)

	followerClient := client(t, agents[1], peerTLSConfig)
	consumeResponse, err = followerClient.Consume(
		context.Background(),
		&api.ConsumeRequest{
			Offset: produceResponse.Offset,
		},
	)
	require.NoError(t, err)
	require.Equal(t, consumeResponse.Record.Value, []byte("foo"))

}

// 도우미함수
func client(
	t *testing.T,
	agent *agent.Agent,
	tlsConfig *tls.Config,
) api.LogClient {
	tlsCreds := credentials.NewTLS(tlsConfig)
	opts := []grpc.DialOption{grpc.WithTransportCredentials(tlsCreds)}
	rpcAddr, err := agent.Config.RPCAddr()
	require.NoError(t, err)
	conn, err := grpc.Dial(fmt.Sprintf(
		"%s",
		rpcAddr,
	), opts...)
	require.NoError(t, err)
	client := api.NewLogClient(conn)
	return client
}
