package documenters

import (
	"documentApi/data"

	"github.com/sirupsen/logrus"
)

type Documenter interface {
	Extension() string
	Name() string
	SerializeRequest(endpoint data.EndpointMetaData) (string, error)                                                                                                   // this returns the serialized request for a single endpoint (maybe this should not be part of the interface)
	SerializeRequests(endpoints []data.EndpointMetaData, collectionName string, outdir string, separateFiles bool, vars map[string]string, logger *logrus.Logger) bool // this saves all the endpoints to the file
	Supports(string) bool
}
