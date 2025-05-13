package documenters

import (
	"documentApi/data"
	"documentApi/utils"
	"os"
	"path"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type InsomniaDocumenter struct{}

func (i InsomniaDocumenter) Extension() string {
	return ".yaml"
}

func (i InsomniaDocumenter) Name() string {
	return "insomnia"
}

func (i InsomniaDocumenter) SerializeRequest(endpoint data.EndpointMetaData) string {
	return ""
}

// this returns the serialized request for a single endpoint
func (i InsomniaDocumenter) SerializeRequests(endpoints []data.EndpointMetaData, collectionName string, outputDir string, separateFiles bool, logger *logrus.Logger) bool {
	// separateFiles is a no-op for insomnia, it outputs a single collection file

	var filePath = path.Join(outputDir, collectionName+i.Extension())
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("InsomniaDocumenter SerializeRequests - Error opening collection file: " + err.Error())
		return false
	}
	defer file.Close()

	var collection data.InsomniaCollection = data.InsomniaCollection{
		Type: "collection.insomnia.rest/5.0",
		Name: "Scratch Pad", // Think this needs to be this way if not logged in
	}
	var collectionRequests = make([]data.InsomniaCollectionItem, 0, len(endpoints))

	var timeStamp int64 = time.Now().Unix()
	var sortKey int = 0 // TODO: sort endpoints
	for _, endpoint := range endpoints {
		if i.Supports(endpoint.TriggerType) {
			endpoint.Route = path.Join("{{host}}", replacePathVars(endpoint.Route)) // TODO: add host to env vars
			collectionRequests = append(collectionRequests, data.InsomniaCollectionItem{
				Url:            endpoint.Route,
				Name:           endpoint.Name,
				Method:         endpoint.Methods[0], // using only the first method for now
				PathParameters: arrayToMapArray(endpoint.PathParameters),
				Meta: data.InsomniaCollectionItemMeta{
					Id:        "req_" + utils.GenerateId(),
					Created:   timeStamp, // TODO: preserve timestamp if updating
					Modified:  timeStamp,
					IsPrivate: false,
					SortKey:   sortKey,
				},
			})
			sortKey++
		}
	}

	collection.Collection = collectionRequests

	err = yaml.NewEncoder(file).Encode(collection)
	if err != nil {
		logger.Error("InsomniaDocumenter SerializeRequests - Error saving insomnia collection: " + err.Error())
		return false
	}

	return true
}

func (i InsomniaDocumenter) Supports(triggerType string) bool {
	return triggerType == data.TriggerType["Http"]
}

func arrayToMapArray(array []string) []map[string]string {
	m := make([]map[string]string, 0, len(array))

	for _, item := range array {
		keyValue := make(map[string]string)
		keyValue["name"] = item
		keyValue["value"] = ""
		m = append(m, keyValue)
	}

	return m
}
