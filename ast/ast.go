package ast

// Node interface for all AST nodes
type Node interface {
	NodeType() string
}

// Template represents the root of the parsed template, containing a sequence of top-level nodes.
type Template struct {
	RootNodes []Node // Sequence of FenceSection, ScriptSection, StyleSection, Element, TextNode, etc.
}

func (t *Template) NodeType() string { return "Template" }

// FenceSection represents the fence section of a template
type FenceSection struct {
	Imports    []ImportNode
	Props      []PropNode
	Variables  []VariableNode // Assuming VariableNode exists or will be added
	RawContent string         // Store raw JS content for now
}

func (f *FenceSection) NodeType() string { return "FenceSection" }

// ScriptSection represents the script section
type ScriptSection struct {
	Content string
}

func (s *ScriptSection) NodeType() string { return "ScriptSection" }

// StyleSection represents the style section
type StyleSection struct {
	Content string
}

func (s *StyleSection) NodeType() string { return "StyleSection" }

// ImportNode represents an import statement
type ImportNode struct {
	Name string
	Path string
}

func (i *ImportNode) NodeType() string { return "Import" }

// PropNode represents a prop declaration
type PropNode struct {
	Name         string
	DefaultValue string // Store the default value expression as string
}

func (p *PropNode) NodeType() string { return "Prop" }

// VariableNode represents a variable declaration in the fence
type VariableNode struct {
	Keyword string // let, const, var
	Name    string
	Value   string // Store the value expression as string
}

func (v *VariableNode) NodeType() string { return "Variable" }

// Element represents an HTML element
type Element struct {
	TagName     string
	Attributes  []Attribute
	Children    []Node // Can contain Element, TextNode, Conditional, Loop, Component etc.
	SelfClosing bool   // Might not be needed if parser handles void elements
}

func (e *Element) NodeType() string { return "Element" }

// Attribute represents an HTML attribute with special handling for Alpine.js directives
type Attribute struct {
	Name       string
	Value      string
	Dynamic    bool   // true if value contains {}
	IsAlpine   bool   // true if this is an Alpine.js directive
	AlpineType string // "data", "bind", "on", etc.
	AlpineKey  string // For x-bind:class, this would be "class"
}

// TextNode represents a text node
type TextNode struct {
	Content string
}

func (t *TextNode) NodeType() string { return "Text" }

// CommentNode represents an HTML comment
type CommentNode struct {
	Content string
}

func (c *CommentNode) NodeType() string { return "Comment" }

// ExpressionNode represents a {} expression within text or attributes
type ExpressionNode struct {
	Expression string
}

func (e *ExpressionNode) NodeType() string { return "Expression" }

// Conditional represents an if/else if/else structure
type Conditional struct {
	IfCondition      string // Store expression string
	IfContent        []Node
	ElseIfConditions []string // Store expression strings
	ElseIfContent    [][]Node
	ElseContent      []Node
}

func (c *Conditional) NodeType() string { return "Conditional" }

// Loop represents a for loop
type Loop struct {
	Iterator   string
	Value      string // Optional value variable for 'in' loops
	Collection string // Store expression string
	Content    []Node
	IsOf       bool // true for "of", false for "in"
}

func (l *Loop) NodeType() string { return "Loop" }

// ComponentNode represents a component instance
type ComponentNode struct {
	Name    string // e.g., "Head" or "./path/comp.html" for dynamic
	Props   []ComponentProp
	Dynamic bool // True if tag starts with <=
}

func (c *ComponentNode) NodeType() string { return "Component" }

// ComponentProp represents a prop passed to a component
type ComponentProp struct {
	Name        string
	Value       string // Store expression string or static value string
	IsShorthand bool   // True for {prop} shorthand
	IsDynamic   bool   // True for prop={expression}
}

// --- Simple Directive Nodes ---

// ElseIfNode represents an {else if condition} tag
type ElseIfNode struct {
	Condition string
}

func (n *ElseIfNode) NodeType() string { return "ElseIf" }

// ElseNode represents an {else} tag
type ElseNode struct{}

func (n *ElseNode) NodeType() string { return "Else" }

// IfEndNode represents an {/if} tag
type IfEndNode struct{}

func (n *IfEndNode) NodeType() string { return "IfEnd" }

// ForEndNode represents an {/for} tag
type ForEndNode struct{}

func (n *ForEndNode) NodeType() string { return "ForEnd" }
