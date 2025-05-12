package data

import "fmt"

type FileMetaData struct {
	Name string
	Path string
}

type EndpointMetaData struct {
	Name           string   `json:"name"`
	Authentication []string `json:"authentication,omitempty"`
	Route          string   `json:"route,omitempty"`
	Methods        []string `json:"methods,omitempty"`
	PathParameters []string `json:"pathParameters,omitempty"`
	Description    string   `json:"description,omitempty"` // TODO: use ai to generate this?
	Body           string   `json:"body,omitempty"`        // potentially a json string...parse the cs classes to get the body?
	ResponseCodes  []int    `json:"responseCodes,omitempty"`
	Interval       string   `json:"interval,omitempty"` // for time triggers, the cron expression
	TriggerType    string   `json:"triggerType,omitempty"`
}

func (e EndpointMetaData) String() string {
	return fmt.Sprintf("Name: %s, Authentication: %v, Route: %s, Methods: %v", e.Name, e.Authentication, e.Route, e.Methods)
}

type ApiMetaData struct {
	Version    string `json:"version"`
	Extensions struct {
		Http struct {
			RoutePrefix string `json:"routePrefix"`
		} `json:"http"`
	} `json:"extensions"`
}
