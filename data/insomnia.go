package data

type InsomniaCollection struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
	Meta struct {
		Id       string `yaml:"id,omitempty"`
		Created  int64  `yaml:"created,omitempty"`
		Modified int64  `yaml:"modified,omitempty"`
	} `yaml:"meta,omitempty"`
	Collection  []InsomniaCollectionItem `yaml:"collection"`
	Environment InsomniaEnvironment      `yaml:"environments,omitempty"` // despite being plural, this is not an array
	// CookieJar
}

type InsomniaCollectionItem struct {
	Url     string                     `yaml:"url"`
	Name    string                     `yaml:"name"`
	Meta    InsomniaCollectionItemMeta `yaml:"meta"`
	Method  string                     `yaml:"method"`
	Headers map[string]string          `yaml:"headers,omitempty"`
	// Settings       struct{}
	PathParameters []map[string]string `yaml:"pathParameters,omitempty"`
}

type InsomniaCollectionItemMeta struct {
	Id        string `yaml:"id"`
	Created   int64  `yaml:"created"`
	Modified  int64  `yaml:"modified"`
	IsPrivate bool   `yaml:"isPrivate"`
	SortKey   int    `yaml:"sortKey"`
}

type InsomniaEnvironment struct {
	Name string                  `yaml:"name"`
	Meta InsomniaEnvironmentMeta `yaml:"meta"`
	Data map[string]string       `yaml:"data"`
}

type InsomniaEnvironmentMeta struct {
	Id        string `yaml:"id"`
	Created   int64  `yaml:"created"`
	Modified  int64  `yaml:"modified"`
	IsPrivate bool   `yaml:"isPrivate"`
}
