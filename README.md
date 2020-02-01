# ElasticMaltego
Create Maltego transforms and import packages (mtz) from  ElasticSearch term queries

# Why 

Maltego has a huge number of really useful transforms for working with OSINT data, however there is not much in the way of transform for
PSINT data (housed within the  SIEM). The purpose of this project is to provide an easy process by which data housed in ElasticSearch databases
can be aggregated and delivered to Maltego as a set of entities.

# How

This tool is able to either output maltego entities directly using a query, or produce a Maltego transform mtz export for inclusion in your 
Maltego UI.

# Usage 

`./ElasticMaltego -h`

```
Usage of ./ElasticMaltego
  -ES string
        ElasticSearch URL
  -debug
        debug output
  -field string
        Elasticsearch field name for termquery
  -index string
        ElasticSearch index to query
  -limit int
        Result limit for query (default 1000)
  -list
        list ElasticSearch indexes
  -map string
        ElasticSearch to Maltego entity mapping i.e data.ip:maltego.IPv4Address
  -query string
        Elasticsearch term query
  -type string
        Maltego source entity i.e maltego.IPV4Address
```

## Transform package output
`./ElasticMaltego --field data.ip -m maltego.IPV4Address   `

```
Generates valid maltego mtz files for transform import with the structure below:


├── Servers
│   └── Local.tas
├── TransformRepositories
│   └── Local
│       ├── <transformGroup>.<indexName>-<ByMaltegoEntity>.transform
│       └── <transformGroup>.<indexName>-<ByMaltegoEntity>.transformsettings
└── version.properties

```

