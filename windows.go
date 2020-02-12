// +build windows

package main

const PATH_SEPARATOR = '\\'

var (
	maltegoTransformBasepath = "c:\\Temp\\ElasticMaltego\\"
	maltegoTransformLocal    = maltegoTransformBasepath + "TransformRepositories\\Local\\"
	maltegoLocalServers      = maltegoTransformBasepath + "\\Servers\\"
)
