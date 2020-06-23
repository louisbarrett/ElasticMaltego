# ElasticMaltego
Create Maltego transforms and import packages (mtz) from  ElasticSearch term queries

# Why 

Maltego has a huge number of really useful transforms for working with OSINT data, however there is not much in the way of transforms for
PSINT data (housed within the  SIEM). The purpose of this project is to provide an easy process by which data housed in ElasticSearch databases
can be aggregated and delivered to Maltego as a set of entities.

# How

This tool is able to either output maltego entities directly using an elasticsearch query, or produce a Maltego transform mtz pacckage for import into your 
Maltego UI. 

Transform package is created in the local temp directory by default

# Usage 

`./ElasticMaltego -h`

```
  -ES string
        ElasticSearch URL
  -config string
        Path to an elastic to maltego csv file
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
```

## Creating a maltego Transform package

`./ElasticMaltego --awsokta=false --config ~/Projects/ElasticMaltego/Elastic.csv --weeks 1 --mtz ~`

```
Generates valid maltego mtz files for transform import with the structure below:


├── Servers
│   └── Local.tas
├── TransformRepositories
│   └── Local
│       ├── <transformGroup>.<indexName>-<ByMaltegoEntity>.transform
│       └── <transformGroup>.<indexName>-<ByMaltegoEntity>.transformsettings
|      . . . 
└── version.properties

```

## Sample Elastic.csv file 
|BaseIndex|	Field	|InputType|	FieldEntityMap|
| ------------- | ------------- |---------|------|
| cloudtrail | sourceIPAddress.keyword  |maltego.IPv4Address|sourceIPAddress.keyword:maltego.IPv4Address,userAgent.keyword:maltego.Phrase,userIdentity.principalId.keyword:maltego.Person,userIdentity.arn.keyword:maltego.Person|
| cloudtrail  | userAgent.keyword  |maltego.Phrase|sourceIPAddress.keyword:maltego.IPv4Address,userAgent.keyword:maltego.Phrase,userIdentity.principalId.keyword:maltego.Person,userIdentity.arn.keyword:maltego.Person|
| cloudtrail  |  userIdentity.sessionContext.sessionIssuer.userName|maltego.Phrase|sourceIPAddress.keyword:maltego.IPv4Address,userAgent.keyword:maltego.Phrase,userIdentity.principalId.keyword:maltego.Person,userIdentity.arn.keyword:maltego.Person|

## Creating a single Transform
`./ElasticMaltego -index $INDEX -map data.ip.keyword:maltego.IPv4Address -field data.ip.keyword -query $QUERY`

Returns maltego XML when used as a local transform
```
        <MaltegoMessage>
        <MaltegoTransformResponseMessage>
        <Entities>

        <Entity Type="maltego.EmailAddress">
        <Value>re@dacted.com</Value>
        <Weight>100</Weight>
        </Entity>

        <Entity Type="maltego.Person">
        <Value>Louis Barrett</Value>
        <Weight>100</Weight>
        </Entity>

        <Entity Type="maltego.IPv4Address">
        <Value>123.45.67.89</Value>
        <Weight>100</Weight>
        </Entity>

        <Entity Type="maltego.Phrase">
        <Value>Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/79.0.3945.117 Safari/537.36</Value>
        <Weight>100</Weight>
        </Entity>

        </Entities>
        </MaltegoTransformResponseMessage>
        </MaltegoMessage>
```
