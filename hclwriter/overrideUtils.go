package hclwriter

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

var overriddenPrefix = make(map[string]bool, 0)

type overridePropertyKey struct {
	property     string
	resourceType string
}

type overridePropertyPair struct {
	path       string
	defaultVal interface{}
}

const replaceValue = "to_be_replaced"

var overridePropertiesPaths = map[overridePropertyKey]overridePropertyPair{
	// overridePropertyPairs with their defaultVal set to "to_be_replaced" will have defaults set to values in original JSON

	// awscc_quicksight_data_source
	{"dataSourceId", "dataSource"}: {"dataSourceId", replaceValue},
	{"name", "dataSource"}:         {"name", "QSFormatDataSourceJSON"},
	{"bucket", "dataSource"}:       {"dataSourceParameters.s3Parameters.manifestFileLocation", replaceValue},

	// awscc_quicksight_data_set
	{"dataSetId", "dataSet"}: {"dataSetId", replaceValue},
	{"name", "dataSet"}:      {"name", "QSFormatDataSetJSON"},

	// awscc_quicksight_analysis
	{"analysisId", "analysis"}: {"analysisId", replaceValue},
	{"name", "analysis"}:       {"name", "QSFormatAnalysisJSON"},

	// awscc_quicksight_dashboard
	{"dashboardId", "dashboard"}: {"dashboardId", replaceValue},
	{"name", "dashboard"}:        {"name", "QSFormatDashboardJSON"},

	// awscc_quicksight_theme
	{"themeId", "theme"}: {"themeId", replaceValue},
	{"name", "theme"}:    {"name", "QSFormatThemeJSON"},

	// awscc_quicksight_refresh_schedule
	{"dataSetId", "refreshSchedule"}:          {"dataSetId", replaceValue},
	{"scheduleId", "refreshSchedule"}:         {"schedule.scheduleId", replaceValue},
	{"startAfterDateTime", "refreshSchedule"}: {"schedule.startAfterDateTime", nil},

	// awscc_quicksight_vpc_connection
	{"vpcConnectionId", "vpcConnection"}:  {"vpcConnectionId", replaceValue},
	{"dnsResolvers", "vpcConnection"}:     {"dnsResolvers", replaceValue},
	{"name", "vpcConnection"}:             {"name", "QSFormatVPCConnectionJSON"},
	{"roleArn", "vpcConnection"}:          {"roleArn", replaceValue},
	{"securityGroupIds", "vpcConnection"}: {"securityGroupIds", replaceValue},
	{"subnetIds", "vpcConnection"}:        {"subnetIds", replaceValue},
}

func overrideResourceProperties(body *hclwrite.Body, dataMap map[string]interface{}, paths []string, resourceType, id string) error {
	for _, path := range paths {
		property := path[getLastDotIndex(path)+1:]
		variableId := getVariableId(resourceType, id, property)
		err := overrideResourceProperty(body, dataMap, pathOf(path), 0, variableId)
		if err != nil {
			return fmt.Errorf("error overriding properties: %v", err)
		}
	}

	return nil
}

func overrideResourceProperty(body *hclwrite.Body, newDataMap map[string]interface{}, blocks []string, index int, variableId string) error {
	if index == len(blocks) {
		return nil
	}
	value, ok := newDataMap[blocks[index]]
	if !ok {
		return fmt.Errorf("attribute \"%s\" could not be found in \"%v\"", blocks[index], newDataMap)
	}

	switch value := value.(type) {
	case map[string]interface{}:
		if len(blocks) == 1 {
			return fmt.Errorf("overridePropertiesPaths map contains incorrect path for %s", variableId)
		}
		err := overrideResourceProperty(body, value, blocks, index+1, variableId)
		if err != nil {
			return err
		}
		return nil
	default:
		variableBlock := body.FirstMatchingBlock("variable", []string{variableId})
		if variableBlock != nil {
			variableBlockBody := variableBlock.Body()
			variableDefaultValue := bytes.TrimSpace(variableBlockBody.GetAttribute(defaultValue).Expr().BuildTokens(nil).Bytes())
			replaceDefaultValue := hclwrite.NewExpressionLiteral(cty.StringVal(replaceValue)).BuildTokens(nil).Bytes()
			if reflect.DeepEqual(variableDefaultValue, replaceDefaultValue) {
				switch v := newDataMap[blocks[0]].(type) {
				case string:
					variableBlockBody.SetAttributeValue(defaultValue, cty.StringVal(v))
				case bool:
					variableBlockBody.SetAttributeValue(defaultValue, cty.BoolVal(v))
				case int64:
					variableBlockBody.SetAttributeValue(defaultValue, cty.NumberIntVal(v))
				default:
					variableBlockBody.SetAttributeValue(defaultValue, cty.NilVal)
				}
			}
			newDataMap[blocks[index]] = "var." + variableId
		}

		return nil
	}
}

func getResourceTypeFromArn(arn string) (string, error) {
	var res string
	blocks := splitOn(arn, ":")
	blocks = splitOn(blocks[len(blocks)-1], "/")
	if len(blocks) < 2 {
		return "", fmt.Errorf("invalid arn: %s", arn)
	} else {
		res = blocks[len(blocks)-2]
		if res == "" {
			return "", fmt.Errorf("invalid arn: %s", arn)
		}
	}

	return strings.ReplaceAll(res, "-", ""), nil
}

func getVariableId(resourceType, id, property string) string {
	return resourceType + "-" + id + "-" + property
}

func removeHCLQuotes(hclFile []byte, paths []string, resourceType, id string) []byte {
	var hclFileBytes = hclFile
	for _, path := range paths {
		property := path[getLastDotIndex(path)+1:]
		variableId := getVariableId(resourceType, id, property)
		hclFileBytes = bytes.Replace(hclFileBytes, []byte("\"var."+variableId+"\""), []byte("var."+variableId), 1)
	}

	return hclFileBytes
}

func addPrefix(dataMap map[string]interface{}, k string) map[string]interface{} {
	var arn, id string
	v := dataMap[k].(string)
	lastSlash := strings.LastIndex(v, "/")
	if lastSlash != -1 {
		arn = v[:lastSlash+1]
		id = v[lastSlash+1:]
	} else {
		arn = ""
		id = v
	}
	value := `format(` + arn + `%s` + id + `, var.prefix-for-all-resources)`
	dataMap[k] = value
	if _, ok := overriddenPrefix[value]; !ok {
		overriddenPrefix[value] = true
	}
	return dataMap
}

func removePrefixQuotes(hclFile []byte) []byte {
	var hclFileBytes = hclFile
	for elem := range overriddenPrefix {
		openParentheses := strings.LastIndex(elem, "(")
		comma := strings.LastIndex(elem, ",")
		newElem := elem[:openParentheses+1] + "\"" + elem[openParentheses+1:comma] + "\"" + elem[comma:]
		hclFileBytes = bytes.Replace(hclFileBytes, []byte("\""+elem+"\""), []byte(newElem), -1)
	}

	return hclFileBytes
}
