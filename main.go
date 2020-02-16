package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/olivere/elastic"
	"github.com/sha1sum/aws_signing_client"
)

var (
	esURL           = os.Getenv("ES_URL")
	searchResults   = []*elastic.SearchHit{}
	indexFlag       = flag.String("index", "", "ElasticSearch index to query")
	urlFlag         = flag.String("ES", "", "ElasticSearch URL")
	queryFlag       = flag.String("query", "", "Elasticsearch term query ")
	queryFieldFlag  = flag.String("field", "", "Elasticsearch field name for term query")
	listFlag        = flag.Bool("list", false, "List ElasticSearch indexes on ES_URL")
	debugFlag       = flag.Bool("debug", false, "Enable debug output")
	limitFlag       = flag.Int("limit", 1000, "Result limit for query")
	weeksFlag       = flag.Int("weeks", 4, "Number of weeks to search in each log")
	mapFlag         = flag.String("map", "", "ElasticSearch to Maltego entity mapping i.e data.ip:maltego.IPv4Address")
	transformConfig = flag.String("config", "", "Path to an elastic to maltego csv file")
	parsedJSON      *gabs.Container
	//AWSOktaPath the location of aws-okta such as /usr/bin/aws-okta
	AWSOktaPath = os.Getenv("AWS_OKTA_PATH")

	//transformGroup the prefix for the transforms
	transformGroup = "sirt"

	// filesCreated - Contains a list of files generated as a part of this package
	filesCreated []string

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

func createMaltegoTransform(maltegoInputType string, elasticIndexName string, queryField string, entityMap string) {

	OK := os.MkdirAll(maltegoTransformLocal, os.ModePerm)
	// Create the folder for Local.tas
	OK = os.MkdirAll(maltegoLocalServers, os.ModePerm)
	if OK != nil {
		fmt.Println("An error occured", OK)
	}

	for _, maltegoSourceEntity := range strings.Split(maltegoInputType, ",") {

		transformAppPath := os.Args[0]
		// Simplify index name for transform name
		transformDisplayName := strings.Split(elasticIndexName, "-")[0]
		transformDisplayName = transformDisplayName + "-" + "By" + strings.Split(maltegoSourceEntity, ".")[1]
		transformName := transformGroup + "." + transformDisplayName
		// Define .transform template
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
		// Write transform meta data
		transformXML = strings.ReplaceAll(transformXML, "TRANSFORM_GROUP", transformGroup)
		transformXML = strings.ReplaceAll(transformXML, "TRANSFORM_DISPLAY_NAME", transformDisplayName)
		transformXML = strings.ReplaceAll(transformXML, "MALTEGO_SOURCE_ENTITY", maltegoSourceEntity)

		// Define .transformsettings template
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
		// Write transformsettings data
		transformSettings = strings.ReplaceAll(transformSettings, "TRANSFORM_APP_PATH", transformAppPath)
		transformSettings = strings.ReplaceAll(transformSettings, "ES_FLAG", esURL)
		transformSettings = strings.ReplaceAll(transformSettings, "INDEX_FLAG", elasticIndexName)
		transformSettings = strings.ReplaceAll(transformSettings, "FIELD_FLAG", queryField)
		transformSettings = strings.ReplaceAll(transformSettings, "TRANSFORM_MAPPINGS", entityMap)
		transformSettings = strings.ReplaceAll(transformSettings, "AWS_OKTA", AWSOktaPath)

		if *debugFlag {
			fmt.Println(transformSettings, transformXML)
		}
		// Create .transform file from bytes
		transformFile, err := os.Create(maltegoTransformLocal + transformName + ".transform")
		filesCreated = append(filesCreated, (maltegoTransformLocal + transformName + ".transform"))

		// Create .transformSettings file from bytes
		transformSettingsFile, err := os.Create(maltegoTransformLocal + transformName + ".transformsettings")
		filesCreated = append(filesCreated, (maltegoTransformLocal + transformName + ".transformsettings"))

		if err != nil {
			log.Fatal("File creation failed", err)
		}

		transformFile.Write([]byte(transformXML))
		transformSettingsFile.Write([]byte(transformSettings))

	}

}

func generateTASFile() {
	// Update local.tas entry to include this transform
	localTASLocation, err := ioutil.ReadDir(maltegoTransformLocal)
	if err != nil {
		fmt.Print("Failed to access directory")
		return
	}

	var transformNameBlob string
	for _, transform := range localTASLocation {

		transformName := transform.Name()
		result, _ := regexp.Match("(settings$)", []byte(transformName))
		if result {

			transformName = strings.Replace(transformName, ".transformsettings", "", 1)

			fmt.Println(transformName)
			transformNameBlob = transformNameBlob + (strings.ReplaceAll(`
			<Transform name="TRANSFORM_DISPLAY_NAME"/>`, "TRANSFORM_DISPLAY_NAME", transformName))
		}
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
	// Add local.tas to the files list
	filesCreated = append(filesCreated, (maltegoLocalServers + "local.tas"))
}

func createZIPFile() {
	// Create the file in which to write the zip payload - os.Create
	zipBuffer := new(bytes.Buffer)
	zipWriter := zip.NewWriter(zipBuffer)

	for i := range filesCreated {
		zipPath := strings.Replace(filesCreated[i], (maltegoTransformBasepath + PATH_SEPARATOR), "", 1)
		// zip files prefer forward slashes
		zipPath = strings.ReplaceAll(zipPath, "\\", "/")
		zipContent, err := zipWriter.Create(zipPath)
		if err != nil {
			log.Fatal(err)
		}
		// get contents of the written file
		originalPath := filesCreated[i]
		originalBytes, err := ioutil.ReadFile(originalPath)
		if err != nil {
			log.Fatal(err)
		}
		zipContent.Write(originalBytes)
	}
	zipWriter.Flush()
	zipWriter.Close()
	ioutil.WriteFile("ElasticMaltego.mtz", zipBuffer.Bytes(), os.FileMode(0700))
}

func runESQuery(query string, index string, maltegoEntitys []queryTransform) {
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
	Now := time.Now()
	for Weeks := 1; Weeks < *weeksFlag; Weeks++ {

		Now = Now.AddDate(0, 0, 7*(Weeks*-1))
		YearInt, WeekInt := Now.ISOWeek()
		Year := strconv.Itoa(YearInt)
		WeekToRotate := fmt.Sprintf("%02d", WeekInt)

		// Query Client creation
		QueryIndex := index + "-" + Year + "-" + "w" + WeekToRotate
		// fmt.Println(QueryIndex)
		esc, err := elastic.NewClient(elastic.SetURL(esURL), elastic.SetScheme("https"), elastic.SetHttpClient(awsClient), elastic.SetSniff(false), elastic.SetHealthcheck(false))
		if err != nil {
			log.Fatal("ES client creation failed")
		}
		Client := esc.Search().Index(QueryIndex).Size(*limitFlag)

		// Parse Aggregations
		transformAggregation := elastic.NewFilterAggregation().Filter(QueryInput)
		for _, entityTransform := range maltegoEntitys {
			if *debugFlag {
				fmt.Println(entityTransform.field, "-->", entityTransform.maltegoType)
			}
			// Create aggregation filter from query input
			transformAggregation = transformAggregation.SubAggregation(entityTransform.maltegoType, elastic.NewTermsAggregation().Field(entityTransform.field).MinDocCount(0).Size(2000))

			Client = esc.Search().Index(QueryIndex).Size(*limitFlag).Aggregation("top", transformAggregation).Size(1)
			if *debugFlag {
				fmt.Println(transformAggregation.Source())
			}
			Results, err := Client.Do(context.Background())
			if err != nil {
				fmt.Println("ES Query failed ", err)
				return
			}
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

		}

		// Query Client Finalized
		maltegoMessage := strings.Replace(maltegoMessageTemplate, "MALTEGO_ENTITY", maltegoEntities, -1)
		fmt.Println(maltegoMessage)
	}
}

func main() {
	flag.Parse()
	// check for AWS_OKTA path variable set

	// list indexes
	if *listFlag {
		// ElasticSearch client initialization
		creds := credentials.NewEnvCredentials()
		signer := v4.NewSigner(creds)
		awsClient, err := aws_signing_client.New(signer, nil, "es", "us-west-2")
		esc, err := elastic.NewClient(elastic.SetURL(esURL), elastic.SetScheme("https"), elastic.SetHttpClient(awsClient), elastic.SetSniff(false), elastic.SetHealthcheck(false))
		if err != nil {
			log.Fatal("ES client creation failed", err)
		} else {

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

	// is the --config flag set? If so generate a transform .mtz file from the entity map csv
	if *transformConfig != "" {
		if AWSOktaPath == "" {
			log.Fatal("Please set the path to your aws-okta binary in the AWS_OKTA_PATH environment variable")
		}
		transformCSVFile, err := os.Open(*transformConfig)
		if err != nil {
			log.Fatal(err)
		}
		transformCSVData := csv.NewReader(transformCSVFile)
		transformCSVData.Read()
		for {

			CSVRow, err := transformCSVData.Read()
			if err != nil {
				generateTASFile()
				createZIPFile()
				return
			}
			createMaltegoTransform(CSVRow[2], CSVRow[0], CSVRow[1], CSVRow[3])
		}
		// create a transform for reach row in the spreadsheet

	}

	// Parse mapFlag into queryTransforms -- create transform output from elastic search query
	if *mapFlag != "" && *transformConfig == "" {
		maltegoTransforms := []queryTransform{}

		userTransforms := (strings.Split(*mapFlag, ","))
		for _, transformMap := range userTransforms {
			maltegoTransforms = append(maltegoTransforms, queryTransform{field: (strings.Split(transformMap, ":")[0]), maltegoType: (strings.Split(transformMap, ":")[1])})
		}
		runESQuery(string(*queryFieldFlag)+"\""+string(*queryFlag)+"\"", *indexFlag, maltegoTransforms)
		// If no valid flags are set show the usage doc
	} else {
		flag.Usage()
	}

}
