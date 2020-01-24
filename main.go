package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"

	"github.com/Jeffail/gabs"
	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/olivere/elastic"
	"github.com/sha1sum/aws_signing_client"
)

var (
	esURL         = os.Getenv("ES_URL")
	searchResults = []*elastic.SearchHit{}
	esc           *elastic.Client
	indexFlag     = flag.String("index", "", "ElasticSearch index to query")
	queryFlag     = flag.String("query", "", "Elasticsearch lucene query")
	listFlag      = flag.Bool("list", false, "list ElasticSearch indexes")
	limitFlag     = flag.Int("limit", 1000, "Result limit for query")
)

func executeQuery() {
	// Check to see if ElasticSearch URL is set
	if esURL == "" {
		log.Fatal("Please set ES_URL environment variable")
	} else {
		// ElasticSearch client initialization
		creds := credentials.NewEnvCredentials()
		signer := v4.NewSigner(creds)
		awsClient, err := aws_signing_client.New(signer, nil, "es", "us-west-2")
		if err != nil {
			fmt.Println("Unable to create AWS client", err)
		}
		esc, err = elastic.NewClient(elastic.SetURL(esURL), elastic.SetScheme("https"), elastic.SetHttpClient(awsClient), elastic.SetSniff(false), elastic.SetHealthcheck(false))
		if err != nil {
			log.Fatal("ES client creation failed")
		}
	}

	// Initialize flags

	if *listFlag {
		IndexNames, err := (esc.IndexNames())
		if err != nil {
			log.Fatal("An error occured ", err)
		}
		sort.Strings(IndexNames)
		for i := range IndexNames {
			SystemIndex, _ := regexp.Match("^\\.", []byte(IndexNames[i]))
			if SystemIndex == false {
				fmt.Println(IndexNames[i])
			}
		}
	}

	if *indexFlag != "" {
		if *queryFlag != "" {
			QueryInput := elastic.NewQueryStringQuery(*queryFlag)
			// Issue elasticsearch query
			Results, err := esc.Search().Index(*indexFlag).Size(*limitFlag).Query(QueryInput).Do(context.Background())
			if err != nil {
				log.Fatal("ES Query failed", err)
			} else {
				// Process the query hits (security events)
				searchResults = Results.Hits.Hits
				for i := range Results.Hits.Hits {
					ResultBytes := (Results.Hits.Hits[i].Source)
					MarshallOutput, err := ResultBytes.MarshalJSON()
					if err != nil {
						log.Fatal("Error encountered marshalling JSON", err)
					}
					ParsedJSON, err := gabs.ParseJSON(MarshallOutput)
					if err != nil {
						log.Fatal("An error occured parsing the JSON", err)
					}
					_ = string(ParsedJSON.String())
					// fmt.Println(string(ParsedJSON.String()))
				}
			}
		} else {
			flag.Usage()
		}
	} else {
		flag.Usage()
	}
}

func main() {
	flag.Parse()

	executeQuery()
	// fmt.Println(searchResults[0].Source)
	for i := range searchResults {
		ResultBytes := (searchResults[i].Source)
		MarshallOutput, err := ResultBytes.MarshalJSON()
		if err != nil {
			log.Fatal("Error encountered marshalling JSON", err)
		}
		ParsedJSON, err := gabs.ParseJSON(MarshallOutput)
		if err != nil {
			log.Fatal("An error occured parsing the JSON", err)
		}
		fmt.Println(string(ParsedJSON.String()))
		// fmt.Println(string(ParsedJSON.String()))

	}

}
