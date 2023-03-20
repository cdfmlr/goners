package goners

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type SessionID string

type PcapSessionConfig struct {
	Device  string        `json:"device"`
	Filter  string        `json:"filter"`
	Snaplen int           `json:"snaplen"`
	Promisc bool          `json:"promisc"`
	Timeout time.Duration `json:"timeout"`

	Format PacketsFormater
	Output Outputer
}

type pcapSession struct {
	ID     SessionID
	Config *PcapSessionConfig
	cancel context.CancelFunc // stop CaptureLivePackets
}

type PcapSessionsManager interface {
	StartSession(config *PcapSessionConfig) (SessionID, error)
	CloseSession(id SessionID) error
}

type pcapSessionsManager struct {
	sessions map[SessionID]*pcapSession
	mutex    sync.RWMutex
}

func (m *pcapSessionsManager) StartSession(config *PcapSessionConfig) (SessionID, error) {
	ctx, cancel := context.WithCancel(context.Background())

	packets, err := CaptureLivePackets(
		ctx,
		config.Device,
		config.Filter,
		int32(config.Snaplen),
		config.Promisc,
		config.Timeout)
	if err != nil {
		cancel()
		return SessionID(""), err
	}

	if config.Format == nil || config.Output == nil {
		cancel()
		return SessionID(""), fmt.Errorf("bad config: unexpected nil format or nil output")
	}
	go config.Output.Output(config.Format.FormatPackets(packets))

	sessionID := m.newSessionID(config)

	session := pcapSession{
		ID:     sessionID,
		Config: config,
		cancel: cancel,
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.sessions[sessionID] = &session

	return sessionID, nil
}

func (m *pcapSessionsManager) newSessionID(config *PcapSessionConfig) SessionID {
	var sessionID SessionID
	u, err := uuid.NewRandom()
	if err != nil {
		sessionID = SessionID(fmt.Sprint(time.Now().UnixNano()))
	} else {
		sessionID = SessionID(u.String())
	}
	return sessionID
}

func (m *pcapSessionsManager) CloseSession(id SessionID) error {
	m.mutex.RLock()
	session, ok := m.sessions[id]
	m.mutex.RUnlock()

	if !ok {
		return fmt.Errorf("session not found")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	session.cancel()
	delete(m.sessions, id)

	return nil
}

// deprecated
func getPacketsFormater(format string) (PacketsFormater, error) {
	var formater PacketsFormater
	switch format {
	case "text":
		formater = StringPacketsFormater
	case "json":
		formater = JsonPacketsFormater
	default:
		return nil, fmt.Errorf("unknown format: %v", format)
	}

	return formater, nil
}

var pcapSessionsManagerSingleton *pcapSessionsManager

func init() {
	pcapSessionsManagerSingleton = &pcapSessionsManager{
		sessions: make(map[SessionID]*pcapSession),
	}
}

func GetPcapSessionsManager() PcapSessionsManager {
	return pcapSessionsManagerSingleton
}
