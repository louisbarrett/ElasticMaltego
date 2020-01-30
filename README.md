# ElasticMaltego
Create Maltego Transforms from ElasticSearch queries

#Why 

Maltego has a huge number of really useful transforms for working with OSINT data, however there is not much in the way of transform for
PSINT data (housed with the  SIEM). The purpose of this project is to provide an easy process by which data housed in ElasticSearch databases
can be aggregated and delivered to Maltego as a set of entities.

#How

This tool is able to either output maltego entities directly using a query, or produce a Maltego transform mtz export for inclusion in your 
Maltego UI.

#Usage 

```
Usage of ./ElasticMaltego:
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
  -m string
        Maltego source entity i.e maltego.IPV4Address
  -query string
        Elasticsearch term query
```

## Transform package output
```
 ./ElasticMaltego -m gateway-api -field data.ip -m maltego.IPV4Address                                                                                                                                                                      master * ] 6:59 PM ./ElasticMaltego

<TransformSettings enabled="true" disclaimerAccepted="false" showHelp="true" runWithAll="true" favorite="true">
   <Properties>
      <Property name="transform.local.command" type="string" popup="false">/usr/local/bin/aws-okta</Property>
      <Property name="transform.local.parameters" type="string" popup="false">exec security-write  -- ./ElasticMaltego --ES  -index  -field data.ip -query</Property>
      <Property name="transform.local.working-directory" type="string" popup="false">/</Property>
      <Property name="transform.local.debug" type="boolean" popup="false">true</Property>
   </Properties>
</TransformSettings>

        <MaltegoTransform name="sirt.-ByIPV4Address" displayName="-ByIPV4Address" abstract="false" template="false" visibility="public" description="Generated using magic" author="Louis Barrett" requireDisplayInfo="false">
   <TransformAdapter>com.paterva.maltego.transform.protocol.v2api.LocalTransformAdapterV2</TransformAdapter>
   <Properties>
      <Fields>
         <Property name="transform.local.command" type="string" nullable="false" hidden="false" readonly="false" description="The command to execute for this transform" popup="false" abstract="false" visibility="public" auth="false" displayName="Command line">
            <SampleValue></SampleValue>
         </Property>
         <Property name="transform.local.parameters" type="string" nullable="true" hidden="false" readonly="false" description="The parameters to pass to the transform command" popup="false" abstract="false" visibility="public" auth="false" displayName="Command parameters">                  <SampleValue></SampleValue>
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
      <Entity type="maltego.IPV4Address" min="1" max="1"/>
   </InputConstraints>
   <OutputEntities/>
   <StealthLevel>0</StealthLevel>
</MaltegoTransform>
```

