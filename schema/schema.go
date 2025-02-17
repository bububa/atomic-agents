package schema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Schema is message schema interface
type Schema interface {
	// Attachement() returns schema attchement
	Attachement() *Attachement
}

type SchemaPointer interface {
	Schema
	SetAttachement(*Attachement)
}

type Markdownable interface {
	ToMarkdown() string
}

func Stringify(s Schema) string {
	if v, ok := s.(String); ok {
		return string(v)
	}
	bs, _ := json.Marshal(s)
	return string(bs)
}

func ToBytes(s Schema) []byte {
	if v, ok := s.(String); ok {
		return []byte(v)
	}
	bs, _ := json.Marshal(s)
	return bs
}

// SchemaToMarkdown converts a struct instance to Markdown based on `jsonschema` tags.
func SchemaToMarkdown(obj interface{}) string {
	val := reflect.ValueOf(obj)
	// Handle pointer values by dereferencing them
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}

	var sb strings.Builder
	switch val.Kind() {
	case reflect.Slice:
		sliceToMarkdown(&sb, val, 0)
	case reflect.Map:
		mapToMarkdown(&sb, val, StructMdTitleStyle, 0)
	case reflect.Struct:
		structToMarkdown(&sb, val, StructMdTitleStyle, 0)
	default:
		fmt.Fprintf(&sb, "%v", val.Interface())
	}
	return sb.String()
}

type StructMarkdownStyle int

const (
	StructMdTitleStyle StructMarkdownStyle = iota
	StructMdListStyle
	StructMdInListStyle
)

// structToMarkdown converts a struct instance to Markdown based on `jsonschema` tags.
func structToMarkdown(sb *strings.Builder, val reflect.Value, style StructMarkdownStyle, indentLevel int) int {
	// Handle pointer values by dereferencing them
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0
		}
		val = val.Elem()
	}
	typ := val.Type()
	var idx int
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		jsonTag := field.Tag.Get("json")
		jsonSchemaTag := field.Tag.Get("jsonschema")

		// Skip empty fields if they have `omitempty`
		if shouldOmitField(field, fieldValue) {
			continue
		}
		title, desc := parseJSONSchemaTag(jsonSchemaTag)

		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			// Dereference the pointer and check if it's a struct or primitive
			fieldValue = fieldValue.Elem()
		}

		if field.Anonymous {
			if v := structToMarkdown(sb, fieldValue, style, indentLevel); v > 0 {
				idx += v
			}
			continue
		}

		if idx > 0 {
			sb.WriteString("\n")
		}

		// Fallback logic for title selection
		mdTitle := title
		if mdTitle == "" {
			mdTitle = parseJSONTag(jsonTag)
		}
		if mdTitle == "" {
			mdTitle = field.Name
		}
		var (
			indent         = strings.Repeat(" ", indentLevel*2)
			subIndent      = indent
			prefix         string
			subIndentLevel = indentLevel
		)
		switch style {
		case StructMdTitleStyle:
			prefix = "# "
		case StructMdListStyle:
			prefix = "- "
			subIndentLevel += 1
			subIndent += "  "
		case StructMdInListStyle:
			if idx > 0 {
				indent += "  "
				prefix = ""
			} else {
				prefix = "- "
			}
			subIndentLevel += 1
			subIndent += "  "
		}

		switch fieldValue.Kind() {
		case reflect.Struct:
			if idx > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(indent)
			sb.WriteString(prefix + "**" + mdTitle + "**")
			if desc != "" {
				sb.WriteString(": " + desc)
			}
			sb.WriteString("\n")
			// Handle nested structs
			structToMarkdown(sb, fieldValue, StructMdListStyle, subIndentLevel)
		case reflect.Slice:
			if idx > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(indent)
			sb.WriteString(prefix + "**" + mdTitle + "**")
			if desc != "" {
				sb.WriteString(": " + desc)
			}
			sb.WriteString("\n")
			sliceToMarkdown(sb, fieldValue, subIndentLevel)
		case reflect.Map:
			if idx > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(indent)
			sb.WriteString(prefix + "**" + mdTitle + "**")
			if desc != "" {
				sb.WriteString(": " + desc)
			}
			sb.WriteString("\n")
			mapToMarkdown(sb, fieldValue, StructMdListStyle, subIndentLevel)
		default:
			sb.WriteString(indent)
			fmt.Fprintf(sb, "%s**%s**: %v", prefix, mdTitle, fieldValue.Interface())
			if desc != "" {
				fmt.Fprintf(sb, "\n\n%s_%s_", subIndent, desc)
			}
			sb.WriteString("\n")
		}
		idx++
	}
	return idx
}

// mapToMarkdown converts a struct instance to Markdown based on `jsonschema` tags.
func mapToMarkdown(sb *strings.Builder, val reflect.Value, style StructMarkdownStyle, indentLevel int) int {
	// Handle pointer values by dereferencing them
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0
		}
		val = val.Elem()
	}
	keys := val.MapKeys()
	var idx int
	for i, key := range keys {
		fieldValue := val.MapIndex(key)
		if fieldValue.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				continue
			}
			fieldValue = fieldValue.Elem()
		}
		var (
			mdTitle        = fmt.Sprintf("%v", key.Interface())
			indent         = strings.Repeat(" ", indentLevel*2)
			prefix         string
			subIndentLevel = indentLevel
		)
		switch style {
		case StructMdTitleStyle:
			prefix = "# "
		case StructMdListStyle:
			prefix = "- "
			subIndentLevel += 1
		case StructMdInListStyle:
			if i > 0 {
				indent += "  "
				prefix = ""
			} else {
				prefix = "- "
			}
			subIndentLevel += 1
		}
		switch fieldValue.Kind() {
		case reflect.Struct:
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(indent)
			sb.WriteString(prefix + "**" + mdTitle + "**")
			sb.WriteString("\n")
			// Handle nested structs
			structToMarkdown(sb, fieldValue, StructMdListStyle, subIndentLevel)
		case reflect.Slice:
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(indent)
			sb.WriteString(prefix + "**" + mdTitle + "**")
			sb.WriteString("\n")
			sliceToMarkdown(sb, fieldValue, subIndentLevel)
		case reflect.Map:
			if i > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(indent)
			sb.WriteString(prefix + "**" + mdTitle + "**")
			sb.WriteString("\n")
			mapToMarkdown(sb, fieldValue, StructMdListStyle, subIndentLevel)
		default:
			sb.WriteString(indent)
			fmt.Fprintf(sb, "%s**%s**: %v", prefix, mdTitle, fieldValue.Interface())
			sb.WriteString("\n")
		}
		idx++
	}
	return idx
}

// sliceToMarkdown converts a struct instance to Markdown based on `jsonschema` tags.
func sliceToMarkdown(sb *strings.Builder, val reflect.Value, indentLevel int) int {
	// Handle pointer values by dereferencing them
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return 0
		}
		val = val.Elem()
	}
	var idx int
	for i := 0; i < val.Len(); i++ {
		value := val.Index(i)
		if value.Kind() == reflect.Ptr {
			if value.IsNil() {
				continue
			}
			value = value.Elem()
		}
		switch value.Kind() {
		case reflect.Struct:
			if i > 0 {
				sb.WriteString("\n")
			}
			structToMarkdown(sb, value, StructMdInListStyle, indentLevel)
		case reflect.Map:
			if i > 0 {
				sb.WriteString("\n")
			}
			mapToMarkdown(sb, value, StructMdInListStyle, indentLevel)
		case reflect.Slice:
			if i > 0 {
				sb.WriteString("\n")
			}
			sliceToMarkdown(sb, value, indentLevel+1)
		default:
			if i > 0 {
				sb.WriteString("\n")
			}
			indent := strings.Repeat(" ", indentLevel*2)
			fmt.Fprintf(sb, "%s- %v", indent, value)
		}
		idx += 1
	}
	return idx
}

// shouldOmitField checks if a field should be omitted based on `omitempty`
func shouldOmitField(field reflect.StructField, fieldValue reflect.Value) bool {
	// Get the json tag to check for `omitempty`
	jsonTag := field.Tag.Get("json")
	omitempty := strings.Contains(jsonTag, "omitempty")

	// If there's no `omitempty`, keep the field
	if !omitempty {
		return false
	}

	// Check if the field is zero-valued
	if !fieldValue.IsValid() || fieldValue.IsZero() {
		return true
	}

	return false
}

// parseJSONSchemaTag extracts title and description from jsonschema tag.
func parseJSONSchemaTag(tag string) (title, description string) {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "title=") {
			title = strings.TrimPrefix(part, "title=")
		} else if strings.HasPrefix(part, "description=") {
			description = strings.TrimPrefix(part, "description=")
		}
	}
	return title, description
}

// parseJSONTag extracts title from json tag.
func parseJSONTag(tag string) string {
	parts := strings.Split(tag, ",")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}
