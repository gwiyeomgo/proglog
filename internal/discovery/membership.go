package discovery

import (
	"net"

	"go.uber.org/zap"

	"github.com/hashicorp/serf/serf"
)

type Membership struct {
	Config
	handler Handler
	serf    *serf.Serf
	events  chan serf.Event
	logger  *zap.Logger
}

func New(handler Handler, config Config) (*Membership, error) {
	c := &Membership{
		Config:  config,
		handler: handler,
		logger:  zap.L().Named("membership"),
	}
	if err := c.setupSerf(); err != nil {
		return nil, err
	}
	return c, nil
}

type Config struct {
	NodeName       string // 서프 클러스터 안에서 노드의 고유한 id
	BindAddr       string // 가십을 듣기 위한 주소와 포트
	Tags           map[string]string
	StartJoinAddrs []string
}

func (m *Membership) setupSerf() (err error) {
	addr, err := net.ResolveTCPAddr("tcp", m.BindAddr)
	if err != nil {
		return err
	}
	config := serf.DefaultConfig()
	config.Init()
	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port
	m.events = make(chan serf.Event)
	config.EventCh = m.events //이벤트채널
	config.Tags = m.Tags      //서프는 클러스터 내의 노드들에 태그를 공유하는데,클러스터에 이 노드를 어떻게 다룰지 간단히 알려준다.
	config.NodeName = m.Config.NodeName
	m.serf, err = serf.Create(config)
	if err != nil {
		return err
	}
	//고루틴을 시작해서 서프 이벤트를 처리하게 한다
	go m.eventHandler()

	//새로운 노드를 만들어 클러스터에 추가한다면, 새 노드는 이미 클러스터에 있는 노드 중 하나 이상을 가리켜야 한다.
	//새 노드가 기존의 클러스터에 어떻게 조인할지 설정한다.
	if m.StartJoinAddrs != nil {
		_, err = m.serf.Join(m.StartJoinAddrs, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// 인터베이스는 서비스 내의 컴포넌트를 의미하며,어떤 서버가 클러스터에 조인하거나 떠나는 것을 알 수 있어야 한다.
type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

// 서프가 보낸 이벤트를 읽고, 자료형에 따라 적절한 이벤트 채널로 보내는 루프를 실행한다.
// 노드가 클러스터에 조인하거나 떠나면, 서프는 해당 노드까지 포함한 모든 노드에 이벤트를 보낸다.
// 이벤트가 로컬 서버 자신이 발생시킨 것이라면 처리하지 않도록 한다.
func (m *Membership) eventHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					return
				}
				m.handleLeave(member)
			}
		}
	}
}

func (m *Membership) handleJoin(member serf.Member) {
	if err := m.handler.Join(
		member.Name,
		member.Tags["rpc_addr"],
	); err != nil {
		m.logError(err, "failed to join", member)
	}
}

func (m *Membership) handleLeave(member serf.Member) {
	if err := m.handler.Leave(
		member.Name,
	); err != nil {
		m.logError(err, "failed to leave", member)
	}
}

// 해당 서프 멤버가 로컬 멤버인지 맴버명을 확인
func (m *Membership) isLocal(member serf.Member) bool {
	return m.serf.LocalMember().Name == member.Name
}

// 호출 시점의 클러서터 서프 멤버들의 스냅숏을 리턴한다.
func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}

// 자신이 서프 클러스터를 떠난다고 알려준다.
func (m *Membership) Leave() error {
	return m.serf.Leave()
}

// 받은 에러와 메세지를 로그로 저장한다.
func (m *Membership) logError(err error, msg string, member serf.Member) {
	m.logger.Error(
		msg,
		zap.Error(err),
		zap.String("name", member.Name),
		zap.String("rpc_addr", member.Tags["rpc_addr"]),
	)
}
