package mcpapi

// MCP Protocol Version
const MCPVersion = "2025-03-26"

// MCP Protocol Message Types
const (
	MessageTypeRequest      = "request"
	MessageTypeResponse     = "response"
	MessageTypeNotification = "notification"
)

// MCP Protocol Methods
const (
	MethodInitialize = "initialize"
	MethodListTools  = "tools/list"
	MethodCallTool   = "tools/call"
)

// MCP JSON-RPC Request
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCP JSON-RPC Response
type Response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *ErrorObj   `json:"error,omitempty"`
}

// MCP Error Object
type ErrorObj struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Error Codes
const (
	ErrorParseError     = -32700
	ErrorInvalidRequest = -32600
	ErrorMethodNotFound = -32601
	ErrorInvalidParams  = -32602
	ErrorInternal       = -32603
)

// Initialize Request Params
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    ClientCapabilities     `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
	Meta            map[string]interface{} `json:"_meta,omitempty"`
}

// Client Capabilities
type ClientCapabilities struct {
	Roots      *RootsCapability      `json:"roots,omitempty"`
	Sampling   map[string]interface{} `json:"sampling,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// Roots Capability
type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Client Info
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Initialize Result
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
	Instructions    string             `json:"instructions,omitempty"`
}

// Server Capabilities
type ServerCapabilities struct {
	Tools        *ToolsCapability       `json:"tools,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Logging      map[string]interface{} `json:"logging,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// Tools Capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Prompts Capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Server Info
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// List Tools Result
type ListToolsResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// Tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// Call Tool Params
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// Call Tool Result
type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError"`
}

// Tool Content
type ToolContent struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
}
