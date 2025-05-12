package documenters

import (
	"documentApi/data"

	"github.com/sirupsen/logrus"
)

type Documenter interface {
	Extension() string
	Name() string
	SerializeRequest(endpoint data.EndpointMetaData) string                                                                                    // this returns the serialized request for a single endpoint
	SerializeRequests(endpoints []data.EndpointMetaData, collectionName string, outdir string, separateFiles bool, logger *logrus.Logger) bool // this saves all the endpoints to the file
	Supports(string) bool
}
