package api

import (
	"encoding/json"
	"strconv"
	"time"
)

// JSONInt64 handles JSON numbers that may be strings or integers
type JSONInt64 int64

func (j *JSONInt64) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		*j = JSONInt64(v)
		return nil
	}
	// Try int64
	var i int64
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	*j = JSONInt64(i)
	return nil
}

func (j JSONInt64) Int64() int64 {
	return int64(j)
}

// JSONLevel handles level field that can be string or int
type JSONLevel int

func (l *JSONLevel) UnmarshalJSON(data []byte) error {
	// Try string first (e.g., "error", "warning")
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*l = JSONLevel(levelFromString(s))
		return nil
	}
	// Try int
	var i int
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	*l = JSONLevel(i)
	return nil
}

func (l JSONLevel) Int() int {
	return int(l)
}

func levelFromString(s string) int {
	switch s {
	case "debug":
		return 10
	case "info":
		return 20
	case "warning":
		return 30
	case "error":
		return 40
	case "critical":
		return 50
	default:
		return 0
	}
}

// ItemsResponse represents the response from GET /api/1/items
type ItemsResponse struct {
	Err    int         `json:"err"`
	Result ItemsResult `json:"result"`
}

// ItemsResult is the result object from the items endpoint
type ItemsResult struct {
	Items []Item `json:"items"`
	Page  int    `json:"page"`
}

// ItemResponse represents the response from GET /api/1/item/{id}
type ItemResponse struct {
	Err    int  `json:"err"`
	Result Item `json:"result"`
}

// Item represents a Rollbar item (error group)
type Item struct {
	ID                       JSONInt64 `json:"id"`
	Counter                  int       `json:"counter"`
	Title                    string    `json:"title"`
	Level                    JSONLevel `json:"level"`
	LevelString              string    `json:"-"` // Computed
	Status                   string    `json:"status"`
	Environment              string    `json:"environment"`
	Framework                string    `json:"framework"`
	Platform                 string    `json:"platform"`
	TotalOccurrences         int       `json:"total_occurrences"`
	LastOccurrenceTimestamp  int64     `json:"last_occurrence_timestamp"`
	LastOccurrenceTime       time.Time `json:"-"` // Computed
	FirstOccurrenceTimestamp int64     `json:"first_occurrence_timestamp"`
	FirstOccurrenceTime      time.Time `json:"-"` // Computed
	ActivatingOccurrenceID   int64     `json:"activating_occurrence_id"`
	ProjectID                int       `json:"project_id"`
	Hash                     string    `json:"hash"`
	UniqueOccurrences        int       `json:"unique_occurrences"`
}

// LevelToString converts numeric level to string
func LevelToString(level int) string {
	switch level {
	case 10:
		return "debug"
	case 20:
		return "info"
	case 30:
		return "warning"
	case 40:
		return "error"
	case 50:
		return "critical"
	default:
		return "unknown"
	}
}

// ComputeFields populates computed fields from raw data
func (i *Item) ComputeFields() {
	i.LevelString = LevelToString(i.Level.Int())
	if i.LastOccurrenceTimestamp > 0 {
		i.LastOccurrenceTime = time.Unix(i.LastOccurrenceTimestamp, 0)
	}
	if i.FirstOccurrenceTimestamp > 0 {
		i.FirstOccurrenceTime = time.Unix(i.FirstOccurrenceTimestamp, 0)
	}
}

// InstancesResponse represents the response from occurrence listing endpoints
type InstancesResponse struct {
	Err    int             `json:"err"`
	Result InstancesResult `json:"result"`
}

// InstancesResult is the result object from instances endpoints
type InstancesResult struct {
	Instances []Instance `json:"instances"`
	Page      int        `json:"page"`
}

// InstanceResponse represents the response from GET /api/1/instance/{id}
type InstanceResponse struct {
	Err    int      `json:"err"`
	Result Instance `json:"result"`
}

// Instance represents a single occurrence of an error
type Instance struct {
	ID        int64        `json:"id"`
	ItemID    int64        `json:"item_id"`
	Timestamp int64        `json:"timestamp"`
	Time      time.Time    `json:"-"` // Computed
	Version   int          `json:"version"`
	Data      InstanceData `json:"data"`
}

// ComputeFields populates computed fields from raw data
func (i *Instance) ComputeFields() {
	if i.Timestamp > 0 {
		i.Time = time.Unix(i.Timestamp, 0)
	}
}

// ClientInfo contains client-side context (browser, runtime info)
type ClientInfo struct {
	JavaScript *ClientJavaScript `json:"javascript"`
}

// ClientJavaScript contains JavaScript runtime info
type ClientJavaScript struct {
	Browser             string `json:"browser"`
	CodeVersion         string `json:"code_version"`
	SourceMapEnabled    bool   `json:"source_map_enabled"`
	GuessUncaughtFrames bool   `json:"guess_uncaught_frames"`
}

// InstanceData contains the occurrence payload
type InstanceData struct {
	Body        Body                   `json:"body"`
	Level       string                 `json:"level"`
	Environment string                 `json:"environment"`
	Framework   string                 `json:"framework"`
	Platform    string                 `json:"platform"`
	Language    string                 `json:"language"`
	Request     *Request               `json:"request"`
	Server      *Server                `json:"server"`
	Person      *Person                `json:"person"`
	Client      *ClientInfo            `json:"client"`
	Custom      map[string]interface{} `json:"custom"`
	Timestamp   int64                  `json:"timestamp"`
	CodeVersion string                 `json:"code_version"`
}

// Body contains the error details
type Body struct {
	Trace       *Trace       `json:"trace"`
	TraceChain  []Trace      `json:"trace_chain"`
	Message     *Message     `json:"message"`
	CrashReport *CrashReport `json:"crash_report"`
}

// Trace represents a stack trace
type Trace struct {
	Exception Exception `json:"exception"`
	Frames    []Frame   `json:"frames"`
}

// Exception represents exception details
type Exception struct {
	Class       string `json:"class"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

// Frame represents a stack frame
type Frame struct {
	Filename string       `json:"filename"`
	Lineno   int          `json:"lineno"`
	Colno    int          `json:"colno"`
	Method   string       `json:"method"`
	Code     string       `json:"code"`
	Context  FrameContext `json:"context"`
	Argspec  []string     `json:"argspec"`
}

// FrameContext contains code context around the error line
type FrameContext struct {
	Pre  []string `json:"pre"`
	Post []string `json:"post"`
}

// Message represents a message-type error body
type Message struct {
	Body string `json:"body"`
}

// CrashReport for crash-type errors
type CrashReport struct {
	Raw string `json:"raw"`
}

// Request contains HTTP request data
type Request struct {
	URL         string                 `json:"url"`
	Method      string                 `json:"method"`
	Headers     map[string]string      `json:"headers"`
	Params      map[string]interface{} `json:"params"`
	GET         map[string]interface{} `json:"GET"`
	POST        map[string]interface{} `json:"POST"`
	Body        string                 `json:"body"`
	UserIP      string                 `json:"user_ip"`
	QueryString string                 `json:"query_string"`
}

// Server contains server information
type Server struct {
	Host        string   `json:"host"`
	Root        string   `json:"root"`
	Branch      string   `json:"branch"`
	CodeVersion string   `json:"code_version"`
	Argv        []string `json:"argv"`
	PID         int      `json:"pid"`
}

// Person represents the affected user
type Person struct {
	ID       JSONString `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
}

// JSONString handles JSON values that may be strings or numbers
type JSONString string

func (j *JSONString) UnmarshalJSON(data []byte) error {
	// Try string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*j = JSONString(s)
		return nil
	}
	// Try number (convert to string)
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*j = JSONString(n.String())
		return nil
	}
	// Fallback: store raw
	*j = JSONString(string(data))
	return nil
}

func (j JSONString) String() string {
	return string(j)
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Err     int    `json:"err"`
	Message string `json:"message"`
}
