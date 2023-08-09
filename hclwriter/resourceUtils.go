package hclwriter

import (
	"fmt"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type resourcePath struct {
	path     string
	resource string
}

var dependencyPaths = map[string][]resourcePath{
	"dataSource": {
		resourcePath{"vpcConnectionProperties.vpcConnectionArn", "vpcConnection"},
	},
	"dataSet": {
		resourcePath{"physicalTableMap.*.s3Source.dataSourceArn", "dataSource"},
		resourcePath{"logicalTableMap.*.source.dataSetArn", "dataSet"},
	},
	"analysis": {
		resourcePath{"dataSetArns", "dataSet"},
		resourcePath{"themeArn", "theme"},
	},
	"dashboard": {
		resourcePath{"sourceEntity.sourceTemplate.dataSetReferences.dataSetArn", "dataSet"},
		resourcePath{"definition.dataSetIdentifierDeclarations.dataSetArn", "dataSet"},
		resourcePath{"themeArn", "theme"},
	},
	"refreshSchedule": {
		resourcePath{"dataSetId", "dataSet"},
	},
}

// map of paths where key <resource>Id is either named differently or nested
var idPaths = map[string]string{
	"refreshSchedule": "schedule.scheduleId",
	"VPCConnection":   "vpcConnectionId",
}

func getResourceType(dataMap map[string]interface{}, resourceTypeLower string) string {
	var resourceType string
	if resourceTypeLower == "" {
		resourceType = dataMap["resourceType"].(string)
	} else {
		resourceType = resourceTypeLower
	}
	switch resourceType {
	case "datasource":
		resourceType = "dataSource"
	case "dataset":
		resourceType = "dataSet"
	case "refreshschedule":
		resourceType = "refreshSchedule"
	}

	return resourceType
}

func getResourceId(dataMap map[string]interface{}, resourceType string, prefixForAllResources bool) (string, map[string]interface{}, error) {
	if v, ok := idPaths[resourceType]; ok {
		return doGetResourceId(dataMap, pathOf(v), 0, prefixForAllResources)
	} else {
		res := dataMap[resourceType+"Id"].(string)
		if prefixForAllResources {
			dataMap = addPrefix(dataMap, resourceType+"Id")
		}
		return res, dataMap, nil
	}
}

func doGetResourceId(dataMap map[string]interface{}, blocks []string, index int, prefixForAllResources bool) (string, map[string]interface{}, error) {
	var err error
	var id string
	if _, ok := dataMap[blocks[index]]; ok {
		switch value := dataMap[blocks[index]].(type) {
		case map[string]interface{}:
			id, dataMap[blocks[index]], err = doGetResourceId(value, blocks, index+1, prefixForAllResources)
			return id, dataMap, err
		case string:
			if index != len(blocks)-1 {
				return "", dataMap, fmt.Errorf("resource id \"%s\" occured before reaching end of path", value)
			} else {
				if prefixForAllResources {
					dataMap = addPrefix(dataMap, blocks[index])
				}
				return value, dataMap, nil
			}
		}
	}

	return "", nil, fmt.Errorf("failed to get resource id with path %v", blocks)
}

func getResourceDependencyArns(dataMap map[string]interface{}, resourceType string, prefixForAllResources bool) []resourcePath {
	paths := dependencyPaths[resourceType]
	dependencies := []resourcePath{}
	for _, path := range paths {
		blocks := pathOf(path.path)
		arns := getDependencyArnsList(dataMap, blocks, 0, prefixForAllResources)
		for _, arn := range arns {
			dependencies = append(dependencies, resourcePath{arn, path.resource})
		}
	}
	return dependencies
}

func getDependencyArnsList(dataMap interface{}, blocks []string, index int, prefixForAllResources bool) []string {
	if len(blocks) == 0 {
		return nil
	}

	var arns []string

	block := blocks[index]
	if block != "*" {
		value, ok := dataMap.(map[string]interface{})[block]
		if !ok {
			return nil
		}
		arns = getDependencyArns(dataMap, value, blocks, index+1, prefixForAllResources)
	} else { // * in path accounts for user assigned names
		switch v := dataMap.(type) {
		case []interface{}:
			for _, m := range v {
				nestedArns := getDependencyArnsList(m, blocks, index+1, prefixForAllResources)
				arns = append(arns, nestedArns...)
			}
		case map[string]interface{}:
			for _, value := range v {
				nestedArns := getDependencyArns(dataMap, value, blocks, index+1, prefixForAllResources)
				arns = append(arns, nestedArns...)
			}
		}
	}

	return arns
}

func getDependencyArns(dataMap interface{}, value interface{}, blocks []string, index int, prefixForAllResources bool) []string {
	var arns []string

	if nestedData, ok := value.(map[string]interface{}); ok {
		arns = getDependencyArnsList(nestedData, blocks, index, prefixForAllResources)
	} else if nestedData, ok := value.([]interface{}); ok {
		for _, m := range nestedData {
			nestedArns := getDependencyArns(dataMap, m, blocks, index, prefixForAllResources)
			arns = append(arns, nestedArns...)
		}
	} else if arn, ok := value.(string); ok {
		arns = append(arns, arn)
		if prefixForAllResources {
			dataMap = addPrefix(dataMap.(map[string]interface{}), blocks[index-1])
		}
	}

	return arns
}

func setAttributeMap(block *hclwrite.Body, values map[string]interface{}) {
	for k, v := range values {
		k = convertKey(k)

		switch v := v.(type) {
		case bool:
			block.SetAttributeValue(k, cty.BoolVal(v))
		case string:
			block.SetAttributeValue(k, cty.StringVal(v))
		case float64:
			block.SetAttributeValue(k, cty.NumberIntVal(int64(v)))
		case map[string]any:
			nestedBlock := block.AppendNewBlock(k+" =", nil)
			setAttributeMap(nestedBlock.Body(), v)
		case []any:
			setAttributeList(block, k, v)
		default:
			fmt.Printf("unrecognized parameter value type: %T\n", v)
		}
	}
}

func setAttributeList(block *hclwrite.Body, k string, values []interface{}) {
	// open list bracket
	openBracket := hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenOBrack,
			Bytes: []byte(`[`),
		},
		{
			Type: hclsyntax.TokenNewline,
		},
	}
	block.SetAttributeRaw(k, openBracket)
	for _, v := range values {
		switch value := v.(type) {
		case map[string]interface{}:
			body := block.AppendNewBlock("", nil)
			setAttributeMap(body.Body(), value)
		case string:
			block.AppendUnstructuredTokens(hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(`"` + value + `"`),
				},
			})
		default:
			fmt.Printf("unrecognized parameter value type: %T\n", v)
		}
		// add commas between the list values
		block.AppendUnstructuredTokens(hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenComma,
				Bytes: []byte(`,`),
			},
		})
		block.AppendNewline()
	}
	// close list bracket
	block.AppendNewline()
	block.AppendUnstructuredTokens(hclwrite.Tokens{
		{
			Type:  hclsyntax.TokenOBrack,
			Bytes: []byte(`]`),
		},
	})
	block.AppendNewline()
}
