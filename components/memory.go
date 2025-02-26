package components

import (
	"fmt"
	"sync"

	"github.com/bububa/atomic-agents/schema"
)

type MemoryStore interface {
	MaxMessages() int
	TurnID() string
	NewTurn() MemoryStore
	NewMessage(MessageRole, schema.Schema) *Message
	History() []Message
	Reset() MemoryStore
	Copy(MemoryStore)
	MessageCount() int
}

// Memory Manages the chat history for an AI agent.
// threadsafe
type Memory struct {
	//	history is a list of messages representing the chat history.
	history []Message
	//	turnID is the ID of the current turn.
	turnID string
	// maxMessages is the maximum number of messages to keep in history.
	// When exceeded, oldest messages are removed first.
	maxMessages int
	// mtx sync lock
	mtx *sync.RWMutex
}

var _ MemoryStore = (*Memory)(nil)

// NewMemory initializes the Memory with an empty history and optional constraints.
func NewMemory(maxMessages int) *Memory {
	return &Memory{
		maxMessages: maxMessages,
		history:     make([]Message, 0, maxMessages+1),
		mtx:         new(sync.RWMutex),
	}
}

// MaxMessages returns the max number of messages
func (m Memory) MaxMessages() int {
	return m.maxMessages
}

// SetMaxMessages set the max number of messages
func (m *Memory) SetMaxMessages(maxMessages int) *Memory {
	m.maxMessages = maxMessages
	return m
}

// TurnID returns the current turn ID
func (m Memory) TurnID() string {
	return m.turnID
}

// SetTurnID set the current turn ID
func (m *Memory) SetTurnID(turnID string) MemoryStore {
	m.turnID = turnID
	return m
}

// NewTurn initializes a new turn by generating a random turn ID.
func (m *Memory) NewTurn() MemoryStore {
	return m.SetTurnID(NewTurnID())
}

// NewMessage adds a message to the chat history and manages overflow.
func (m *Memory) NewMessage(role MessageRole, content schema.Schema) *Message {
	msg := NewMessage(role, content).SetTurnID(m.turnID)
	m.mtx.Lock()
	// Manages the chat history overflow based on max_messages constraint.
	m.history = append(m.history, *msg)
	l := len(m.history)
	if m.maxMessages > 0 && l > m.maxMessages {
		m.history = m.history[1:]
	}
	m.mtx.Unlock()
	return msg
}

// SetHistory set a copy of chat history
func (m *Memory) SetHistory(history []Message) *Memory {
	m.mtx.Lock()
	m.history = make([]Message, len(history))
	copy(m.history, history)
	m.mtx.Unlock()
	return m
}

// History retrieves the chat history, filtering out unnecessary fields and serializing content.
func (m *Memory) History() []Message {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return m.history
}

// Copy creates a copy of the chat memory.
func (m *Memory) Copy(src MemoryStore) {
	m.SetMaxMessages(src.MaxMessages()).SetTurnID(src.TurnID())
	m.SetHistory(src.History())
}

func (m *Memory) Reset() MemoryStore {
	m.mtx.Lock()
	m.history = make([]Message, 0, m.maxMessages)
	m.mtx.Unlock()
	return m
}

// DeleteTurn delete messages from the memory by its turn ID.
// returns Error if the specified turn ID is not found in the memory
func (m *Memory) DeleteTurn(turnID string) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	l := len(m.history)
	list := make([]Message, 0, l)
	for _, v := range m.history {
		if v.TurnID() == turnID {
			continue
		}
		list = append(list, v)
	}
	m.history = list
	num := len(list)
	if num == l {
		return fmt.Errorf("TurnID %s not found in memory", turnID)
	}
	// Update current_turn_id if necessary
	if len(list) == 0 {
		m.turnID = ""
	} else if turnID == m.turnID {
		m.turnID = m.history[num-1].TurnID()
	}
	return nil
}

// MessageCount returns the number of messages in the chat history.
func (m *Memory) MessageCount() int {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	return len(m.history)
}
