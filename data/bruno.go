package data

type BrunoMeta struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Seq  int    `json:"seq"`
}

type BrunoRequest struct {
	URL  string `json:"url"`
	Body string `json:"body,omitempty"`
	Auth string `json:"auth,omitempty"`

	Method string `json:"method,omitempty"`
}

type BrunoBody struct {
	Body string
	Type string
}

type BrunoCollectionDefinition struct {
	Version string   `json:"version"`
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Ignore  []string `json:"ignore,omitempty"`
}
