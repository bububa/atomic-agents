package components

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"net/http"

	anthropic "github.com/liushuangls/go-anthropic/v2"
	"github.com/rs/xid"
	openai "github.com/sashabaranov/go-openai"

	"github.com/bububa/atomic-agents/schema"
)

// NewTurnID returns a new turn ID.
func NewTurnID() string {
	return xid.New().String()
}

// MessageRole is the role of the message sender (e.g., 'user', 'system', 'tool')
type MessageRole = string

const (
	SystemRole    MessageRole = "system"
	UserRole      MessageRole = "user"
	AssistantRole MessageRole = "assistant"
	ToolRole      MessageRole = "tool"
	FunctionRole  MessageRole = "function"
)

// Message  Represents a message in the chat history.
//
// Attributes:
//
//	role (str): .
//	content: The content of the message.
type Message struct {
	content schema.Schema
	// role is the role of the message sender (e.g., 'user', 'system', 'tool')
	role MessageRole
	//	turnID is Unique identifier for the turn this message belongs to.
	turnID string
}

// NewMessage returns a new Message
func NewMessage(role MessageRole, content schema.Schema) *Message {
	return &Message{
		role:    role,
		content: content,
	}
}

// SetTurnID set message turnID
func (m *Message) SetTurnID(turnID string) *Message {
	m.turnID = turnID
	return m
}

// Role returns message role
func (m Message) Role() MessageRole {
	return m.role
}

// Content returns message content
func (m Message) Content() schema.Schema {
	return m.content
}

// attachement returns message attachement
func (m Message) Attachement() *schema.Attachement {
	return m.content.Attachement()
}

// turnID returns message turnID
func (m Message) TurnID() string {
	return m.turnID
}

// ToOpenAI convert message to openai ChatCompletionMessage
func (m Message) ToOpenAI(dist *openai.ChatCompletionMessage) {
	dist.Role = m.role
	if attachement := m.Attachement(); attachement != nil && len(attachement.ImageURLs) > 0 {
		dist.MultiContent = make([]openai.ChatMessagePart, 0, len(attachement.ImageURLs)+1)
		dist.MultiContent = append(dist.MultiContent, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeText,
			Text: m.content.String(),
		})
		for _, imageURL := range attachement.ImageURLs {
			dist.MultiContent = append(dist.MultiContent, openai.ChatMessagePart{
				Type: openai.ChatMessagePartTypeImageURL,
				ImageURL: &openai.ChatMessageImageURL{
					URL: imageURL,
				},
			})
		}
	} else {
		dist.Content = m.content.String()
	}
}

// ToAnthropic convert message to anthropic Message
func (m Message) ToAnthropic(dist *anthropic.Message) {
	dist.Role = m.role
	if attachement := m.Attachement(); attachement != nil && len(attachement.ImageURLs) > 0 {
		images := getImages(attachement.ImageURLs)
		dist.Content = make([]anthropic.MessageContent, 0, len(images)+1)
		buf := new(bytes.Buffer)
		for _, img := range images {
			png.Encode(buf, img)
			encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
			imgSource := anthropic.MessageContentImageSource{
				Type:      "base64",
				MediaType: "image/png",
				Data:      fmt.Sprintf("data:image/png;base64,%s", encodedString),
			}
			dist.Content = append(dist.Content, anthropic.NewImageMessageContent(imgSource))
		}
	}
	dist.Content = []anthropic.MessageContent{anthropic.NewTextMessageContent(m.content.String())}
}

func getImages(urls []string) []image.Image {
	imgs := make([]image.Image, len(urls))
	for _, link := range urls {
		if img, err := getImage(link); err == nil {
			imgs = append(imgs, img)
		}
	}
	return imgs
}

func getImage(imgURL string) (image.Image, error) {
	resp, err := http.Get(imgURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
}
