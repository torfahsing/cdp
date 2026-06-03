package daemon

import "encoding/json"

type Request struct {
	Command string          `json:"command"`
	Args    json.RawMessage `json:"args,omitempty"`
	Port    int             `json:"port"`
	Timeout int             `json:"timeout"`
	JSON    bool            `json:"jsonMode"`
}

type Response struct {
	OK     bool            `json:"ok"`
	Text   string          `json:"text,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
	Error  string          `json:"error,omitempty"`
	Binary string          `json:"binary,omitempty"`
}

type SnapshotArgs struct {
	Verbose bool `json:"verbose,omitempty"`
}

type ClickArgs struct {
	UID      string `json:"uid"`
	DblClick bool   `json:"dbl,omitempty"`
	Snapshot bool   `json:"snapshot,omitempty"`
}

type HoverArgs struct {
	UID string `json:"uid"`
}

type DragArgs struct {
	FromUID string `json:"fromUid"`
	ToUID   string `json:"toUid"`
}

type EvalArgs struct {
	Script string `json:"script"`
	Dialog string `json:"dialog,omitempty"`
}

type NavigateArgs struct {
	URL         string `json:"url,omitempty"`
	Back        bool   `json:"back,omitempty"`
	Forward     bool   `json:"forward,omitempty"`
	Reload      bool   `json:"reload,omitempty"`
	IgnoreCache bool   `json:"ignoreCache,omitempty"`
	WaitLoad    bool   `json:"waitLoad,omitempty"`
}

type SelectArgs struct {
	PageID string `json:"pageId"`
	Focus  bool   `json:"focus,omitempty"`
}

type ClosePageArgs struct {
	PageID string `json:"pageId"`
}

type OpenPageArgs struct {
	URL        string `json:"url"`
	Background bool   `json:"background,omitempty"`
}

type FillArgs struct {
	UID   string `json:"uid"`
	Value string `json:"value"`
}

type FillFormArgs struct {
	Fields map[string]string `json:"fields"`
}

type TypeArgs struct {
	Text      string `json:"text"`
	SubmitKey string `json:"submitKey,omitempty"`
}

type KeyArgs struct {
	Key string `json:"key"`
}

type WaitArgs struct {
	Texts []string `json:"texts"`
}

type DialogArgs struct {
	Action string `json:"action"`
	Text   string `json:"text,omitempty"`
}

type ScreenshotArgs struct {
	Format   string `json:"format,omitempty"`
	Quality  int    `json:"quality,omitempty"`
	FullPage bool   `json:"fullPage,omitempty"`
	UID      string `json:"uid,omitempty"`
}

type ResizeArgs struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type UploadArgs struct {
	UID      string `json:"uid"`
	FilePath string `json:"filePath"`
}

type EmulateArgs struct {
	Viewport    string `json:"viewport,omitempty"`
	ColorScheme string `json:"colorScheme,omitempty"`
	UserAgent   string `json:"userAgent,omitempty"`
	CPUThrottle float64 `json:"cpuThrottle,omitempty"`
	Geo         string `json:"geo,omitempty"`
	Network     string `json:"network,omitempty"`
}

type PerfStartArgs struct {
	Reload   bool `json:"reload,omitempty"`
	AutoStop bool `json:"autoStop,omitempty"`
}

type PerfStopArgs struct {
	FilePath string `json:"filePath,omitempty"`
}

type PerfInsightArgs struct {
	SetID string `json:"setId"`
	Name  string `json:"name"`
}

type MemoryArgs struct {
	FilePath string `json:"filePath"`
}

type ConsoleArgs struct {
	Types []string `json:"types,omitempty"`
}

func MarshalArgs(v any) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}
