package components

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"

	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/gabriel-vasile/mimetype"
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

// ApiResponse instructor provider chat response
type ApiResponse struct {
	ID        string      `json:"id,omitempty"`
	Role      MessageRole `json:"role,omitempty"`
	Model     string      `json:"model,omitempty"`
	Usage     *ApiUsage   `json:"usage,omitempty"`
	Timestamp int64       `json:"ts,omitempty"`
	Details   any         `json:"content,omitempty"`
}

// FromOpenAI convnert response from openai
func (r *ApiResponse) FromOpenAI(v *openai.ChatCompletionResponse) {
	r.ID = v.ID
	r.Role = AssistantRole
	r.Model = v.Model
	r.Usage = &ApiUsage{
		InputTokens:  v.Usage.PromptTokens,
		OutputTokens: v.Usage.CompletionTokens,
	}
	r.Details = v.Choices
}

// FromAnthropic convert response from anthropic
func (r *ApiResponse) FromAnthropic(v *anthropic.MessagesResponse) {
	r.ID = v.ID
	r.Role = AssistantRole
	r.Model = string(v.Model)
	r.Usage = &ApiUsage{
		InputTokens:  v.Usage.InputTokens,
		OutputTokens: v.Usage.OutputTokens,
	}
	r.Details = v.Content
}

// FromCohere convert response from cohere
func (r *ApiResponse) FromCohere(v *cohere.NonStreamedChatResponse) {
	if v.GenerationId != nil {
		r.ID = *v.GenerationId
	}
	r.Role = AssistantRole
	if meta := v.Meta; meta != nil {
		if usage := meta.Tokens; usage != nil {
			r.Usage = new(ApiUsage)
			if usage.InputTokens != nil {
				r.Usage.InputTokens = int(*usage.InputTokens)
			}
			if usage.OutputTokens != nil {
				r.Usage.OutputTokens = int(*usage.OutputTokens)
			}
		}
		if version := meta.ApiVersion; version != nil {
			r.Model = version.Version
		}
	}
	r.Details = v
}

type ApiUsage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
}

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
			Text: schema.Stringify(m.content),
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
		dist.Content = schema.Stringify(m.content)
	}
}

// ToAnthropic convert message to anthropic Message
func (m Message) ToAnthropic(dist *anthropic.Message) {
	dist.Role = anthropic.ChatRole(m.role)
	if attachement := m.Attachement(); attachement != nil && (len(attachement.ImageURLs) > 0 || len(attachement.Files) > 0) {
		images := getImages(attachement.ImageURLs)
		dist.Content = make([]anthropic.MessageContent, 0, len(images)+len(attachement.Files)+1)
		buf := new(bytes.Buffer)
		for _, img := range images {
			buf.Reset()
			png.Encode(buf, img)
			encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
			imgSource := anthropic.MessageContentSource{
				Type:      "base64",
				MediaType: "image/png",
				Data:      fmt.Sprintf("data:image/png;base64,%s", encodedString),
			}
			dist.Content = append(dist.Content, anthropic.NewImageMessageContent(imgSource))
		}
		for _, f := range attachement.Files {
			buf.Reset()
			tee := io.TeeReader(f, buf)
			mimeType, _ := mimetype.DetectReader(tee)
			encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
			docSource := anthropic.MessageContentSource{
				Type:      "base64",
				MediaType: mimeType.String(),
				Data:      fmt.Sprintf("data:%s;base64,%s", mimeType.String(), encodedString),
			}
			dist.Content = append(dist.Content, anthropic.NewDocumentMessageContent(docSource))
		}
	}
	dist.Content = []anthropic.MessageContent{anthropic.NewTextMessageContent(schema.Stringify(m.content))}
}

// ToCohere convert message to cohere Message
func (m Message) ToCohere(dist *cohere.Message) {
	dist.Role = m.role
	switch m.role {
	case SystemRole:
		dist.Role = "SYSTEM"
		dist.System = &cohere.ChatMessage{
			Message: schema.Stringify(m.content),
		}
	case AssistantRole:
		dist.Role = "CHATBOT"
		dist.System = &cohere.ChatMessage{
			Message: schema.Stringify(m.content),
		}
	case UserRole:
		dist.Role = "USER"
		dist.User = &cohere.ChatMessage{
			Message: schema.Stringify(m.content),
		}
	}
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
