package app

import (
	"fmt"
	"html"
	"math/rand"
	"strings"
	"unicode/utf16"
)

func utf16Slice(s string, offset, length int) string {
	runes := []rune(s)

	utf16Units := utf16.Encode(runes)
	if offset >= len(utf16Units) {
		return ""
	}

	end := min(offset+length, len(utf16Units))

	r := utf16.Decode(utf16Units[offset:end])
	return string(r)
}

func wrapHTML(tag, value, extra string) string {
	if tag == "" {
		return value
	}

	if extra != "" {
		return fmt.Sprintf("<%s %s>%s</%s>", tag, extra, value, tag)
	}

	return fmt.Sprintf("<%s>%s</%s>", tag, value, tag)
}

func UnparseEntitiesToHTML(text string, entities []messageEntity) string {
	if len(entities) == 0 {
		return html.EscapeString(text)
	}

	for i := 0; i < len(entities)-1; i++ {
		for j := i + 1; j < len(entities); j++ {
			if entities[i].Offset > entities[j].Offset {
				entities[i], entities[j] = entities[j], entities[i]
			}
		}
	}

	var result strings.Builder
	lastOffset := 0

	for _, e := range entities {
		if e.Offset < lastOffset {
			continue
		}

		if e.Offset > lastOffset {
			result.WriteString(html.EscapeString(utf16Slice(text, lastOffset, e.Offset-lastOffset)))
		}

		entityText := utf16Slice(text, e.Offset, e.Length)

		switch e.Type {
		case "bold":
			result.WriteString(wrapHTML("b", entityText, ""))
		case "italic":
			result.WriteString(wrapHTML("i", entityText, ""))
		case "underline":
			result.WriteString(wrapHTML("u", entityText, ""))
		case "strikethrough":
			result.WriteString(wrapHTML("s", entityText, ""))
		case "spoiler":
			result.WriteString(wrapHTML("tg-spoiler", entityText, ""))
		case "code":
			result.WriteString(wrapHTML("code", html.EscapeString(entityText), ""))
		case "pre":
			result.WriteString(wrapHTML("pre", html.EscapeString(entityText), ""))
		case "pre_language":
			result.WriteString(wrapHTML("pre", html.EscapeString(entityText), fmt.Sprintf(`class="language-%s"`, *e.Language)))
		case "text_link":
			result.WriteString(wrapHTML("a", entityText, fmt.Sprintf(`href="%s"`, *e.Url)))
		case "text_mention":
			result.WriteString(wrapHTML("a", entityText, fmt.Sprintf(`href="tg://user?id=%d"`, e.User.ID)))
		case "custom_emoji":
			result.WriteString(wrapHTML("tg-emoji", entityText, fmt.Sprintf(`emoji-id="%s"`, *e.CustomEmojiID)))
		case "blockquote":
			result.WriteString(wrapHTML("blockquote", entityText, ""))
		case "expandable_blockquote":
			result.WriteString(wrapHTML("blockquote", entityText, "expandable"))
		default:
			result.WriteString(html.EscapeString(entityText))
		}

		lastOffset = e.Offset + e.Length
	}

	utf16Len := len(utf16.Encode([]rune(text)))

	if lastOffset < utf16Len {
		result.WriteString(html.EscapeString(
			utf16Slice(text, lastOffset, utf16Len-lastOffset),
		))
	}

	return result.String()
}

func parseSpintax(input string) string {
	for {
		start := strings.LastIndex(input, "{")
		if start == -1 {
			break
		}

		end := strings.Index(input[start:], "}")
		if end == -1 {
			break
		}

		end += start

		block := input[start+1 : end]
		options := strings.Split(block, "|")
		if len(options) == 0 {
			break
		}

		choice := options[rand.Intn(len(options))]
		input = input[:start] + choice + input[end+1:]
	}

	return input
}
