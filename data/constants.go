package data

var (
	TriggerType = map[string]string{
		"Http":      "http",
		"Timer":     "timer",
		"EventGrid": "event-grid",
		"CosmosDB":  "cosmos",
		"UNKNOWN":   "unknown",
	}
)
