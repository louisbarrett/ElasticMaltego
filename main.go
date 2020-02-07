package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/olivere/elastic"
	"github.com/sha1sum/aws_signing_client"
)

var (
	esURL          = os.Getenv("ES_URL")
	searchResults  = []*elastic.SearchHit{}
	indexFlag      = flag.String("index", "", "ElasticSearch index to query")
	urlFlag        = flag.String("ES", "", "ElasticSearch URL")
	maltegoInput   = flag.String("type", "", "Maltego source entity i.e maltego.IPV4Address")
	queryFlag      = flag.String("query", "", "Elasticsearch term query ")
	queryFieldFlag = flag.String("field", "", "Elasticsearch field name for termquery")
	listFlag       = flag.Bool("list", false, "list ElasticSearch indexes")
	debugFlag      = flag.Bool("debug", false, "debug output")
	limitFlag      = flag.Int("limit", 1000, "Result limit for query")
	mapFlag        = flag.String("map", "", "ElasticSearch to Maltego entity mapping i.e data.ip:maltego.IPv4Address")
	parsedJSON     *gabs.Container
	AWSOktaPath    = os.Getenv("AWSOktaPath")

	entityTemplate = `
	<Entity Type="TYPE">
	<Value>DATA</Value>
	<Weight>100</Weight>
	</Entity>
	`
	maltegoMessageTemplate = `
	<MaltegoMessage>
	<MaltegoTransformResponseMessage>
	<Entities>
		MALTEGO_ENTITY
	</Entities>
	</MaltegoTransformResponseMessage>
	</MaltegoMessage>
	`
	maltegoEntities   string
	transformNameBlob string
)

type queryTransform struct {
	field       string
	maltegoType string
	limit       int
}

func createMaltegoTransform(maltegoSourceEntities string) {
	// TRANSFORM_APP_PATH
	// MALTEGO_SOURCE_ENTITY
	// FIELD_FLAG
	// INDEX_FLAG
	// ES_FLAG
	// TRANSFORM_GROUP
	// TRANSFORM_DISPLAY_NAME

	// Validate maltego statements
	// maltegoTemporary := (strings.Split(maltegoSourceEntities, ","))

	// Create Maltego folder structure

	// Define temporary path and Maltego zip Folder name
	// maltegoTransformBasepath := "/tmp/SirtMaltego/"
	maltegoTransformBasepath := "c:\\Temp\\"
	maltegoTransformLocal := maltegoTransformBasepath + "TransformRepositories/Local/"
	// Generate TAS file including all transforms
	// maltegoLocalServers := maltegoTransformBasepath + "/Servers/"
	maltegoLocalServers := maltegoTransformBasepath + "\\Servers\\"

	OK := os.MkdirAll(maltegoTransformLocal, os.ModePerm)
	// Create teh folder for Local.tas
	OK = os.MkdirAll(maltegoLocalServers, os.ModePerm)
	if OK != nil {
		fmt.Println("An error occured", OK)
	}

	for _, maltegoSourceEntity := range strings.Split(maltegoSourceEntities, ",") {

		transformAppPath := os.Args[0]
		fmt.Println(transformAppPath)
		var transformGroup = "sirt"
		transformDisplayName := strings.Split(*indexFlag, "-")[0]
		transformDisplayName = transformDisplayName + "-" + "By" + strings.Split(maltegoSourceEntity, ".")[1]
		transformName := transformGroup + "." + transformDisplayName
		transformXML := `
	<MaltegoTransform name="TRANSFORM_GROUP.TRANSFORM_DISPLAY_NAME" displayName="TRANSFORM_DISPLAY_NAME" abstract="false" template="false" visibility="public" description="Generated using magic" author="Louis Barrett" requireDisplayInfo="false">
   <TransformAdapter>com.paterva.maltego.transform.protocol.v2api.LocalTransformAdapterV2</TransformAdapter>
   <Properties>
      <Fields>
         <Property name="transform.local.command" type="string" nullable="false" hidden="false" readonly="false" description="The command to execute for this transform" popup="false" abstract="false" visibility="public" auth="false" displayName="Command line">
            <SampleValue></SampleValue>
         </Property>
         <Property name="transform.local.parameters" type="string" nullable="true" hidden="false" readonly="false" description="The parameters to pass to the transform command" popup="false" abstract="false" visibility="public" auth="false" displayName="Command parameters">
            <SampleValue></SampleValue>
         </Property>
         <Property name="transform.local.working-directory" type="string" nullable="true" hidden="false" readonly="false" description="The working directory used when invoking the executable" popup="false" abstract="false" visibility="public" auth="false" displayName="Working directory">
            <DefaultValue>/</DefaultValue>
            <SampleValue></SampleValue>
         </Property>
         <Property name="transform.local.debug" type="boolean" nullable="true" hidden="false" readonly="false" description="When this is set, the transform&apos;s text output will be printed to the output window" popup="false" abstract="false" visibility="public" auth="false" displayName="Show debug info">
            <SampleValue>false</SampleValue>
         </Property>
      </Fields>
   </Properties>
   <InputConstraints>
      <Entity type="MALTEGO_SOURCE_ENTITY" min="1" max="1"/>
   </InputConstraints>
   <OutputEntities/>
   <StealthLevel>0</StealthLevel>
</MaltegoTransform>`

		transformXML = strings.ReplaceAll(transformXML, "TRANSFORM_GROUP", transformGroup)
		transformXML = strings.ReplaceAll(transformXML, "TRANSFORM_DISPLAY_NAME", transformDisplayName)
		transformXML = strings.ReplaceAll(transformXML, "MALTEGO_SOURCE_ENTITY", maltegoSourceEntity)

		transformSettings := `
   <TransformSettings enabled="true" disclaimerAccepted="false" showHelp="true" runWithAll="true" favorite="true">
   <Properties>
      <Property name="transform.local.command" type="string" popup="false">AWS_OKTA</Property>
      <Property name="transform.local.parameters" type="string" popup="false">exec security-write  -- TRANSFORM_APP_PATH --ES ES_FLAG -index INDEX_FLAG -field FIELD_FLAG -map TRANSFORM_MAPPINGS -query</Property>
      <Property name="transform.local.working-directory" type="string" popup="false">/</Property>
      <Property name="transform.local.debug" type="boolean" popup="false">false</Property>
   </Properties>
</TransformSettings>
`

		transformSettings = strings.ReplaceAll(transformSettings, "TRANSFORM_APP_PATH", transformAppPath)
		transformSettings = strings.ReplaceAll(transformSettings, "ES_FLAG", esURL)
		transformSettings = strings.ReplaceAll(transformSettings, "INDEX_FLAG", *indexFlag)
		transformSettings = strings.ReplaceAll(transformSettings, "FIELD_FLAG", *queryFieldFlag)
		transformSettings = strings.ReplaceAll(transformSettings, "TRANSFORM_MAPPINGS", *mapFlag)
		transformSettings = strings.ReplaceAll(transformSettings, "AWS_OKTA", AWSOktaPath)

		fmt.Println(transformSettings, transformXML)

		// Create .transform file from bytes
		transformFile, err := os.Create(maltegoTransformLocal + transformName + ".transform")
		// Create .transformSettings file from bytes
		transformSettingsFile, err := os.Create(maltegoTransformLocal + transformName + ".transformsettings")
		if err != nil {
			log.Fatal("File creation failed", err)
		}

		transformFile.Write([]byte(transformXML))
		transformSettingsFile.Write([]byte(transformSettings))
		// Update local.tas entry to include this transform
		transformNameBlob = transformNameBlob + (strings.ReplaceAll(`<Transform name="TRANSFORM_DISPLAY_NAME"/>
		`, "TRANSFORM_DISPLAY_NAME", transformName))

	}

	localTAS := `
	<MaltegoServer name="Local" enabled="true" description="Local transforms hosted on this machine" url="http://localhost">
	   <LastSync>2020-01-29 14:21:09.761 PST</LastSync>
	   <Protocol version="0.0"/>
	   <Authentication type="none"/>
	   <Transforms> 
		  TRANSFORM_DISPLAY_NAME
	   </Transforms>
	   <Seeds/>
	</MaltegoServer>
`
	TASFile, err := os.Create(maltegoLocalServers + "local.tas")
	if err != nil {
		log.Fatal("File creation failed", err)
	}
	localTAS = strings.ReplaceAll(localTAS, "TRANSFORM_DISPLAY_NAME", transformNameBlob)
	TASFile.Write([]byte(localTAS))
	// Write ZIP files to ..\TransformRepositories\Local\<displayName>.
}

func runESQuery(query string, index string, maltegoEntitys []queryTransform) *gabs.Container {
	jsonResults := gabs.New()

	// Check to see if ElasticSearch URL is set
	if esURL == "" {
		if *urlFlag != "" {
			esURL = *urlFlag
		} else {
			log.Fatal("Please set ES_URL environment variable")
		}
	}

	jsonResults.Array("data")
	// ElasticSearch client initialization
	creds := credentials.NewEnvCredentials()
	signer := v4.NewSigner(creds)
	awsClient, err := aws_signing_client.New(signer, nil, "es", "us-west-2")
	if err != nil {
		fmt.Println("Unable to create AWS client", err)
	}

	// Parse Query input
	QueryInput := elastic.NewQueryStringQuery(query)

	// Query Client creation
	esc, err := elastic.NewClient(elastic.SetURL(esURL), elastic.SetScheme("https"), elastic.SetHttpClient(awsClient), elastic.SetSniff(false), elastic.SetHealthcheck(false))
	if err != nil {
		log.Fatal("ES client creation failed")
	}
	Client := esc.Search().Index(index).Size(*limitFlag)

	// Parse Aggregations
	transformAggregation := elastic.NewFilterAggregation().Filter(QueryInput)
	for _, entityTransform := range maltegoEntitys {
		if *debugFlag {
			fmt.Println(entityTransform.field, "-->", entityTransform.maltegoType)
		}
		// Create aggregation filter from query input
		transformAggregation = transformAggregation.SubAggregation(entityTransform.maltegoType, elastic.NewTermsAggregation().Field(entityTransform.field).MinDocCount(0).Size(2000))

		Client = esc.Search().Index(index).Size(*limitFlag).Aggregation("top", transformAggregation).Size(1)
		if *debugFlag {
			fmt.Println(transformAggregation.Source())
		}
		Results, err := Client.Do(context.Background())
		if err != nil {
			log.Fatal("ES Query failed ", err)
		} else {
			// Process the query aggs

			data, ok := Results.Aggregations.Filters("top")
			groupedResults, ok := data.Aggregations.Terms(entityTransform.maltegoType)
			if !ok {
				log.Fatal("Error Retrieving results ", err)
			}
			for _, k := range groupedResults.Buckets {
				if k.DocCount > 0 {
					entityObject := strings.Replace(entityTemplate, "TYPE", entityTransform.maltegoType, 1)
					entityObject = strings.Replace(entityObject, "DATA", k.Key.(string), 1)
					maltegoEntities = maltegoEntities + entityObject

				}

			}

		} // This is a test
	}

	// Query Client Finalized
	maltegoMessage := strings.Replace(maltegoMessageTemplate, "MALTEGO_ENTITY", maltegoEntities, -1)
	fmt.Println(maltegoMessage)

	return jsonResults
}

func main() {
	flag.Parse()
	// parse user input
	// * fuzzy index look up - Display # next to index
	// * weeks - number of indexes to shard through
	// * data_model - data model file to include

	// ElasticMaltegoObject - q, output[field aggregation <---> maltego.TYPE mapping,...]
	// maltego outputs
	// field aggregation <---> maltego.TYPE mapping

	// query - Size 0 for maximum speed - data.ip:12.206.218.178

	// list indexes
	if *listFlag {
		// ElasticSearch client initialization
		creds := credentials.NewEnvCredentials()
		signer := v4.NewSigner(creds)
		awsClient, err := aws_signing_client.New(signer, nil, "es", "us-west-2")
		esc, err := elastic.NewClient(elastic.SetURL(esURL), elastic.SetScheme("https"), elastic.SetHttpClient(awsClient), elastic.SetSniff(false), elastic.SetHealthcheck(false))
		if err != nil {
			fmt.Println("ES client creation failed", err)
			os.Exit(1)
		} else {
			_, err := esc.IndexExists("cloudtrail-2019-w22").Do(context.Background())
			if err != nil {
				fmt.Println(err)
			}

			IndexNames, _ := (esc.IndexNames())
			sort.Strings(IndexNames)
			for i := range IndexNames {
				SystemIndex, _ := regexp.Match("^\\.", []byte(IndexNames[i]))
				if SystemIndex == false {
					fmt.Println(IndexNames[i])
				}
			}
			return
		}
	}

	// Chunk queries across many weeks
	if *maltegoInput != "" {
		createMaltegoTransform(*maltegoInput)
	} else {
		// GroupByUserID := queryTransform{field: "data.viewer.userId.keyword", maltegoType: "segment.UserId", limit: 1000}
		// GroupByUserEmail := queryTransform{field: "data.viewer.userEmail.keyword", maltegoType: "maltego.EmailAddress", limit: 1000}
		// GroupByUserWorkspaceID := queryTransform{field: "data.viewer.userName.keyword", maltegoType: "maltego.Person", limit: 1000}
		// GroupByUserAgent := queryTransform{field: "data.userAgent.keyword", maltegoType: "maltego.Phrase", limit: 1000}

		maltegoTransforms := []queryTransform{}

		// Parse flagMap into queryTransforms
		if *mapFlag != "" && *maltegoInput == "" {
			userTransforms := (strings.Split(*mapFlag, ","))
			for _, transformMap := range userTransforms {
				maltegoTransforms = append(maltegoTransforms, queryTransform{field: (strings.Split(transformMap, ":")[0]), maltegoType: (strings.Split(transformMap, ":")[1])})
			}
			runESQuery(string(*queryFieldFlag)+"\""+string(*queryFlag)+"\"", *indexFlag, maltegoTransforms)

		} else {
			flag.Usage()
		}

	}
}
