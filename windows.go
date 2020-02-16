// +build windows

package main

const PATH_SEPARATOR = "\\"

var (
	maltegoTransformBasepath = "C:" + PATH_SEPARATOR + "temp" + PATH_SEPARATOR + "ElasticMaltego"
	maltegoTransformLocal    = maltegoTransformBasepath + PATH_SEPARATOR + "TransformRepositories" + PATH_SEPARATOR + "Local" + PATH_SEPARATOR
	maltegoLocalServers      = maltegoTransformBasepath + PATH_SEPARATOR + "Servers" + PATH_SEPARATOR
)
