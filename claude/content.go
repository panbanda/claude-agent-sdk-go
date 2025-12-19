package claude

// ContentBlockKind discriminates the type of content in a ContentBlock.
//
// Design rationale: using a single struct with Kind discriminator (Genkit pattern)
// rather than separate types with an interface because:
//   - JSON marshaling is trivial with struct tags.
//   - All 4 variants have similar sizes, minimal memory waste.
//   - Avoids custom UnmarshalJSON complexity for discriminated unions.
//   - Helper methods (IsText, IsToolUse, etc.) provide type checking.
type ContentBlockKind int8

const (
	// BlockText represents text content from the assistant.
	BlockText ContentBlockKind = iota

	// BlockThinking represents extended thinking content.
	BlockThinking

	// BlockToolUse represents a tool invocation request.
	BlockToolUse

	// BlockToolResult represents the result of a tool invocation.
	BlockToolResult
)

// ContentBlock represents a block of content in a message.
// Use the Kind field to determine which fields are relevant,
// or use the Is*() helper methods.
type ContentBlock struct {
	Kind ContentBlockKind `json:"kind"`

	// Text content (Kind == BlockText)
	Text string `json:"text,omitempty"`

	// Thinking content (Kind == BlockThinking)
	Thinking  string `json:"thinking,omitempty"`
	Signature string `json:"signature,omitempty"`

	// Tool use fields (Kind == BlockToolUse or BlockToolResult)
	ToolUseID string         `json:"id,omitempty"`
	ToolName  string         `json:"name,omitempty"`
	ToolInput map[string]any `json:"input,omitempty"`

	// Tool result fields (Kind == BlockToolResult)
	ToolResult any  `json:"content,omitempty"`
	IsError    bool `json:"is_error,omitempty"`
}

// IsText returns true if this is a text content block.
func (b *ContentBlock) IsText() bool {
	return b.Kind == BlockText
}

// IsThinking returns true if this is a thinking content block.
func (b *ContentBlock) IsThinking() bool {
	return b.Kind == BlockThinking
}

// IsToolUse returns true if this is a tool use request block.
func (b *ContentBlock) IsToolUse() bool {
	return b.Kind == BlockToolUse
}

// IsToolResult returns true if this is a tool result block.
func (b *ContentBlock) IsToolResult() bool {
	return b.Kind == BlockToolResult
}

// NewTextBlock creates a new text content block.
func NewTextBlock(text string) *ContentBlock {
	return &ContentBlock{
		Kind: BlockText,
		Text: text,
	}
}

// NewThinkingBlock creates a new thinking content block.
func NewThinkingBlock(thinking, signature string) *ContentBlock {
	return &ContentBlock{
		Kind:      BlockThinking,
		Thinking:  thinking,
		Signature: signature,
	}
}

// NewToolUseBlock creates a new tool use request block.
func NewToolUseBlock(id, name string, input map[string]any) *ContentBlock {
	return &ContentBlock{
		Kind:      BlockToolUse,
		ToolUseID: id,
		ToolName:  name,
		ToolInput: input,
	}
}

// NewToolResultBlock creates a new tool result block.
func NewToolResultBlock(toolUseID string, result any, isError bool) *ContentBlock {
	return &ContentBlock{
		Kind:       BlockToolResult,
		ToolUseID:  toolUseID,
		ToolResult: result,
		IsError:    isError,
	}
}
