// +build !windows

package main

const PATH_SEPARATOR = '/'

var (
	maltegoTransformBasepath = "/tmp/ElasticMaltego"
	maltegoTransformLocal    = maltegoTransformBasepath + "/TransformRepositories/Local/"
	maltegoLocalServers      = maltegoTransformBasepath + "/Servers/"
)
