package components

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"strings"

	cohere "github.com/cohere-ai/cohere-go/v2"
	"github.com/gabriel-vasile/mimetype"
	gemini "github.com/google/generative-ai-go/genai"
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

// LLMResponse instructor provider chat response
type LLMResponse struct {
	ID        string      `json:"id,omitempty"`
	Role      MessageRole `json:"role,omitempty"`
	Model     string      `json:"model,omitempty"`
	Usage     *LLMUsage   `json:"usage,omitempty"`
	Timestamp int64       `json:"ts,omitempty"`
	Details   any         `json:"content,omitempty"`
}

// FromOpenAI convnert response from openai
func (r *LLMResponse) FromOpenAI(v *openai.ChatCompletionResponse) {
	r.ID = v.ID
	r.Role = AssistantRole
	r.Model = v.Model
	r.Usage = &LLMUsage{
		InputTokens:  v.Usage.PromptTokens,
		OutputTokens: v.Usage.CompletionTokens,
	}
	r.Details = v.Choices
}

// FromAnthropic convert response from anthropic
func (r *LLMResponse) FromAnthropic(v *anthropic.MessagesResponse) {
	r.ID = v.ID
	r.Role = AssistantRole
	r.Model = string(v.Model)
	r.Usage = &LLMUsage{
		InputTokens:  v.Usage.InputTokens,
		OutputTokens: v.Usage.OutputTokens,
	}
	r.Details = v.Content
}

// FromCohere convert response from cohere
func (r *LLMResponse) FromCohere(v *cohere.NonStreamedChatResponse) {
	if v.GenerationId != nil {
		r.ID = *v.GenerationId
	}
	r.Role = AssistantRole
	if meta := v.Meta; meta != nil {
		if usage := meta.Tokens; usage != nil {
			r.Usage = new(LLMUsage)
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

func (r *LLMResponse) FromGemini(v *gemini.GenerateContentResponse) {
	r.Role = AssistantRole
	if v.UsageMetadata != nil {
		r.Usage = &LLMUsage{
			InputTokens:  int(v.UsageMetadata.PromptTokenCount),
			OutputTokens: int(v.UsageMetadata.CandidatesTokenCount),
		}
	}
	r.Details = v.Candidates
}

type LLMUsage struct {
	InputTokens  int `json:"input_tokens,omitempty"`
	OutputTokens int `json:"output_tokens,omitempty"`
}

func (u *LLMUsage) Merge(v *LLMUsage) {
	if v == nil {
		return
	}
	u.InputTokens += v.InputTokens
	u.OutputTokens += v.OutputTokens
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

// Attachement returns message attachement
func (m Message) Attachement() *schema.Attachement {
	return m.content.Attachement()
}

// FileIDs returns message attachement file ids
func (m Message) FileIDs() []string {
	attachment := m.content.Attachement()
	if attachment == nil {
		return nil
	}
	return attachment.FileIDs
}

// ImageURLs returns message attachement image urls
func (m Message) ImageURLs() []string {
	attachment := m.content.Attachement()
	if attachment == nil {
		return nil
	}
	return attachment.ImageURLs
}

// Files returns message attachement files
func (m Message) Files() []io.Reader {
	attachment := m.content.Attachement()
	if attachment == nil {
		return nil
	}
	return attachment.Files
}

// Chunks returns message attachement chunks
func (m Message) Chunks() []schema.Schema {
	return m.content.Chunks()
}

// turnID returns message turnID
func (m Message) TurnID() string {
	return m.turnID
}

func (m Message) TryAttachChunkPrompt(idx int) string {
	var txt string
	if idx == 0 {
		if v, ok := m.content.(schema.Markdownable); ok {
			txt = v.ToMarkdown()
		} else {
			txt = schema.Stringify(m.content)
		}
	}
	if l := len(m.Chunks()); l > 0 {
		if idx < l {
			return fmt.Sprintf(`Do not answer yet. This is just another part of the text I want to send you. Just receive and acknowledge as "Part %[1]d/%[2]d received" and wait for the next part.
        [START PART %[1]d/%[2]d]
        %[3]s
        [END PART %[1]d/%[2]d]
        Remember not answering yet. Just acknowledge you received this part with the message "Part %[1]d/%[2]d received" and wait for the next part.`, idx+1, l, txt)
		} else {
			return fmt.Sprintf(`[START PART %[1]d/%[2]d]
        %[3]s
        [END PART %[1]d/%[2]d]
        ALL PARTS SENT. Now you can continue processing the request.`, idx+1, l, txt)
		}
	}
	return txt
}

// ToOpenAI convert message to openai ChatCompletionMessage
func (m Message) ToOpenAI(dist *openai.ChatCompletionMessage) []openai.ChatCompletionMessage {
	m.toOpenAI(dist, 0)
	if l := len(m.Chunks()); l > 0 {
		list := make([]openai.ChatCompletionMessage, 0, l)
		for idx := range l {
			var llmMsg openai.ChatCompletionMessage
			if err := m.toOpenAI(&llmMsg, idx+1); err == nil {
				list = append(list, llmMsg)
			}
		}
		return list
	}
	return nil
}

func (m Message) toOpenAI(dist *openai.ChatCompletionMessage, idx int) error {
	src := m
	chunks := m.Chunks()
	if idx > 0 {
		if len(chunks) > idx {
			src = Message{content: chunks[idx], role: m.role}
		} else {
			return errors.New("invalid chunk index")
		}
	}
	dist.Role = m.role
	txt := m.TryAttachChunkPrompt(idx)
	if attachement := src.Attachement(); attachement != nil && len(attachement.ImageURLs) > 0 {
		dist.MultiContent = make([]openai.ChatMessagePart, 0, len(attachement.ImageURLs)+1)
		dist.MultiContent = append(dist.MultiContent, openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeText,
			Text: txt,
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
		dist.Content = txt
	}
	return nil
}

// ToAnthropic convert message to anthropic Message
func (m Message) ToAnthropic(dist *anthropic.Message) []anthropic.Message {
	m.toAnthropic(dist, 0)
	if l := len(m.Chunks()); l > 0 {
		list := make([]anthropic.Message, 0, l)
		for idx := range l {
			var llmMsg anthropic.Message
			if err := m.toAnthropic(&llmMsg, idx+1); err == nil {
				list = append(list, llmMsg)
			}
		}
		return list
	}
	return nil
}

func (m Message) toAnthropic(dist *anthropic.Message, idx int) error {
	src := m
	chunks := m.Chunks()
	if idx > 0 {
		if len(chunks) > idx {
			src = Message{content: chunks[idx], role: m.role}
		} else {
			return errors.New("invalid chunk index")
		}
	}
	dist.Role = anthropic.ChatRole(m.role)
	txt := m.TryAttachChunkPrompt(idx)
	if attachement := src.Attachement(); attachement != nil && (len(attachement.ImageURLs) > 0 || len(attachement.Files) > 0) {
		images := getImages(attachement.ImageURLs)
		dist.Content = make([]anthropic.MessageContent, 0, len(images)+len(attachement.Files)+1)
		buf := new(bytes.Buffer)
		for _, img := range images {
			buf.Reset()
			jpeg.Encode(buf, img, nil)
			encodedString := base64.StdEncoding.EncodeToString(buf.Bytes())
			imgSource := anthropic.MessageContentSource{
				Type:      "base64",
				MediaType: "image/jpeg",
				Data:      fmt.Sprintf("data:image/jpeg;base64,%s", encodedString),
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
	dist.Content = []anthropic.MessageContent{anthropic.NewTextMessageContent(txt)}
	return nil
}

// ToCohere convert message to cohere Message
func (m Message) ToCohere(dist *cohere.Message) []*cohere.Message {
	m.toCohere(dist, 0)
	if l := len(m.Chunks()); l > 0 {
		list := make([]*cohere.Message, 0, l)
		for idx := range l {
			var llmMsg cohere.Message
			if err := m.toCohere(&llmMsg, idx+1); err == nil {
				list = append(list, &llmMsg)
			}
		}
		return list
	}
	return nil
}

func (m Message) toCohere(dist *cohere.Message, idx int) error {
	chunks := m.Chunks()
	if idx > 0 && len(chunks) <= idx {
		return errors.New("invalid chunk index")
	}
	dist.Role = m.role
	txt := m.TryAttachChunkPrompt(idx)
	switch m.role {
	case SystemRole:
		dist.Role = "SYSTEM"
		dist.System = &cohere.ChatMessage{
			Message: txt,
		}
	case AssistantRole:
		dist.Role = "CHATBOT"
		dist.System = &cohere.ChatMessage{
			Message: txt,
		}
	case UserRole:
		dist.Role = "USER"
		dist.User = &cohere.ChatMessage{
			Message: txt,
		}
	}
	return nil
}

// ToGemini convert message to openai Content
func (m Message) ToGemini(dist *gemini.Content) []*gemini.Content {
	m.toGemini(dist, 0)
	if l := len(m.Chunks()); l > 0 {
		list := make([]*gemini.Content, 0, l)
		for idx := range l {
			var llmMsg gemini.Content
			if err := m.toGemini(&llmMsg, idx+1); err == nil {
				list = append(list, &llmMsg)
			}
		}
		return list
	}
	return nil
}

func (m Message) toGemini(dist *gemini.Content, idx int) error {
	src := m
	chunks := m.Chunks()
	if idx > 0 {
		if len(chunks) > idx {
			src = Message{content: chunks[idx], role: m.role}
		} else {
			return errors.New("invalid chunk index")
		}
	}
	dist.Role = m.role
	if dist.Role == FunctionRole {
		bs := schema.ToBytes(m.content)
		resp := make(map[string]any)
		if err := json.Unmarshal(bs, &resp); err == nil {
			dist.Parts = append(dist.Parts, gemini.FunctionResponse{
				Response: resp,
			})
			return nil
		}
	}
	txt := m.TryAttachChunkPrompt(idx)
	dist.Parts = append(dist.Parts, gemini.Text(txt))
	if attachement := src.Attachement(); attachement != nil && len(attachement.ImageURLs) > 0 {
		images := getImages(attachement.ImageURLs)
		dist.Parts = make([]gemini.Part, 0, len(images)+1)
		buf := new(bytes.Buffer)
		for _, img := range images {
			buf.Reset()
			jpeg.Encode(buf, img, nil)
			bs := make([]byte, buf.Len())
			copy(bs, buf.Bytes())
			dist.Parts = append(dist.Parts, gemini.ImageData("jpeg", bs))
		}
	}
	return nil
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
	var r io.Reader
	if strings.HasPrefix(imgURL, "data") && strings.Contains(imgURL, ";base64,") {
		parts := strings.Split(imgURL, ",")
		if len(parts) != 2 {
			return nil, errors.New("invalid image url")
		}
		bs, err := base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(bs)
	} else {
		resp, err := http.Get(imgURL)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		r = resp.Body
	}
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// func getImageData(imgURL string) ([]byte, error) {
// 	if strings.HasPrefix(imgURL, "data") && strings.Contains(imgURL, ";base64,") {
// 		parts := strings.Split(imgURL, ",")
// 		if len(parts) != 2 {
// 			return nil, errors.New("invalid image url")
// 		}
// 		return base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
// 	}
// 	resp, err := http.Get(imgURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()
// 	return io.ReadAll(resp.Body)
// }
