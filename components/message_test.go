package components

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/bububa/atomic-agents/schema"
)

func TestMessageMarshaler(t *testing.T) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	dec := json.NewDecoder(&buf)
	msg := NewMessage(UserRole, schema.NewString("test string schema"))
	if err := enc.Encode(msg); err != nil {
		t.Fatal(err)
		return
	}
	var decodeMsg Message
	if err := dec.Decode(&decodeMsg); err != nil {
		t.Fatal(err)
		return
	}
	if decodeMsg.StringifiedContent() != msg.StringifiedContent() {
		t.Errorf("string match error, expect:%s, got:%s", msg.StringifiedContent(), decodeMsg.StringifiedContent())
	}
	t.Fatal("failed")
}
