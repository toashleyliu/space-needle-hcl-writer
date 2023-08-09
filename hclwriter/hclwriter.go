package hclwriter

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

func convertFiles(files []fs.DirEntry, srcPath, destPath, hclFileName, awsAccountId string, keys map[string]interface{}, overrideParameters arrayFlags, prefixForAllResources bool) error {
	hclFile := hclwrite.NewEmptyFile()
	var overridePropertiesMap map[string][]string
	var err error

	// if user inputs override parameters, create variable blocks at top of converted file
	if len(overrideParameters) != 0 {
		var variablesFile *hclwrite.File
		*hclFile.Body(), variablesFile, overridePropertiesMap, err = createVariableBlocks(overrideParameters, destPath, hclFile.Body())
		if err != nil {
			return fmt.Errorf("failed to create override variable blocks")
		}
		defer createTerraformFile(destPath+"/variables.tfvars", variablesFile)
	}

	// create terraform block
	tfBlock := createTerraformBlock(hclwrite.NewBlock("terraform", nil))
	hclFile.Body().AppendBlock(tfBlock)

	// create provider block
	provBlock := createProvidersBlock(hclwrite.NewBlock("provider", []string{"awscc"}))
	hclFile.Body().AppendBlock(provBlock)

	// create resource blocks for all .json files in src path
	for _, f := range files {
		if f.IsDir() {
			dir, _ := os.ReadDir(srcPath + "/" + f.Name())
			for _, sd := range dir {
				if sd.IsDir() || !strings.HasSuffix(sd.Name(), ".json") {
					fmt.Printf("ignoring directory/non-JSON file: %s/%s/%s\n", srcPath, f.Name(), sd.Name())
					continue
				} else {
					overridePropertiesMap, err = convertFile(sd, hclFile, srcPath+"/"+f.Name(), awsAccountId, keys, overridePropertiesMap, prefixForAllResources)
					if err != nil {
						return err
					}
				}
			}
		} else if !strings.HasSuffix(f.Name(), ".json") {
			fmt.Printf("ignoring directory/non-JSON file: %s/%s\n", srcPath, f.Name())
			continue
		} else {
			overridePropertiesMap, err = convertFile(f, hclFile, srcPath, awsAccountId, keys, overridePropertiesMap, prefixForAllResources)
			if err != nil {
				return err
			}
		}
	}
	if prefixForAllResources {
		hclFileBytes := removePrefixQuotes(hclFile.Bytes())
		hclFile, _ = hclwrite.ParseConfig(hclFileBytes, "", hcl.Pos{Line: 1, Column: 1})
	}
	// create converted.tf file in dest path
	fmt.Printf("writing terraform file: %s/%s\n", destPath, hclFileName)
	err = createTerraformFile(destPath+"/"+hclFileName, hclFile)
	if err != nil {
		return err
	}

	err = FormatFile(destPath, hclFileName)
	if err != nil {
		return err
	}
	return nil
}

func convertFile(f fs.DirEntry, hclFile *hclwrite.File, srcPath, awsAccountId string, keys map[string]interface{}, overridePropertiesMap map[string][]string, prefixForAllResources bool) (map[string][]string, error) {

	fmt.Printf("reading resource file: %s/%s\n", srcPath, f.Name())
	jsonFile, err := readJson((path.Join(srcPath, f.Name())))
	if err != nil {
		return overridePropertiesMap, err
	}

	dataMap, err := createDataMap(jsonFile)
	if err != nil {
		return overridePropertiesMap, err
	}

	resourceType := getResourceType(dataMap, "")
	id, dataMap, err := getResourceId(dataMap, resourceType, prefixForAllResources)
	if err != nil {
		return overridePropertiesMap, err
	}

	// override property values in data map before being converted
	if overridePropertiesMap != nil {
		clonedDataMap := cloneMap(dataMap)
		err := overrideResourceProperties(hclFile.Body(), clonedDataMap, overridePropertiesMap[resourceType+id], resourceType, id)
		if err != nil {
			return overridePropertiesMap, err
		}
		dataMap = clonedDataMap
	}

	// convert and create resource block
	resourceBlock := hclwrite.NewBlock("resource", []string{"awscc_quicksight_" + convertKey(resourceType), strings.ToLower(resourceType + "-" + id)})
	resBlock := createResourceBlock(resourceBlock, dataMap, awsAccountId, resourceType, keys, prefixForAllResources)
	hclFile.Body().AppendBlock(resBlock)

	// remove the quotes around var.OverridePropertyID in converted file
	if overridePropertiesMap != nil {
		hclFileBytes := removeHCLQuotes(hclFile.Bytes(), overridePropertiesMap[resourceType+id], resourceType, id)
		hclwrite.ParseConfig(hclFileBytes, "", hcl.Pos{Line: 1, Column: 1})
	}

	return overridePropertiesMap, nil
}

func createVariableBlocks(overrideProperties arrayFlags, destPath string, body *hclwrite.Body) (hclwrite.Body, *hclwrite.File, map[string][]string, error) {
	// create variables.tfvars file
	variablesFile := hclwrite.NewEmptyFile()

	var overridePropertiesMap = make(map[string][]string)
	var defaultVal interface{}
	var variableId string

	for _, op := range overrideProperties {
		if op == "prefix-for-all-resources" {
			defaultVal = ""
			variableId = "prefix-for-all-resources"
		} else {
			i := strings.Index(op, ",")
			resourceType, err := getResourceTypeFromArn(string(op[i+1:]))
			resourceType = getResourceType(nil, resourceType)
			if err != nil {
				return *body, nil, nil, err
			}
			property := op[0:i]
			id := getId(op[i+1:])

			path := overridePropertiesPaths[overridePropertyKey{property, resourceType}].path
			defaultVal = overridePropertiesPaths[overridePropertyKey{property, resourceType}].defaultVal
			variableId = getVariableId(resourceType, id, property)

			overridePropertiesMap[resourceType+id] = append(overridePropertiesMap[resourceType+id], path)
		}

		// add variables block to converted file
		variableBlock := createVariableBlock(hclwrite.NewBlock("variable", []string{variableId}), defaultVal)
		body.AppendBlock(variableBlock)

		// add variable prompts to variables.tfvars file
		variablesFile.Body().SetAttributeRaw(variableId, nil)
	}

	return *body, variablesFile, overridePropertiesMap, nil
}

func createVariableBlock(variableBlock *hclwrite.Block, defaultVal interface{}) *hclwrite.Block {
	variableBlockBody := variableBlock.Body()
	switch v := defaultVal.(type) {
	case string:
		variableBlockBody.SetAttributeValue(defaultValue, cty.StringVal(v))
	case bool:
		variableBlockBody.SetAttributeValue(defaultValue, cty.BoolVal(v))
	case int64:
		variableBlockBody.SetAttributeValue(defaultValue, cty.NumberIntVal(v))
	case nil:
		variableBlockBody.SetAttributeValue(defaultValue, cty.NilVal)
	}

	return variableBlock
}

func createTerraformBlock(tfBlock *hclwrite.Block) *hclwrite.Block {
	tfBlockBody := tfBlock.Body()

	reqProvBlock := tfBlockBody.AppendNewBlock("required_providers", nil)
	reqProvBlockBody := reqProvBlock.Body()

	reqProvBlockBody.SetAttributeValue("awscc", cty.ObjectVal(map[string]cty.Value{
		"source":  cty.StringVal("hashicorp/awscc"),
		"version": cty.StringVal("0.52.0"),
	}))

	return tfBlock
}

func createProvidersBlock(provBlock *hclwrite.Block) *hclwrite.Block {
	provBlockBody := provBlock.Body()

	provBlockBody.SetAttributeValue("region", cty.StringVal("us-east-1"))

	return provBlock
}

func createResourceBlock(resBlock *hclwrite.Block, dataMap map[string]interface{}, awsAccountId, resourceType string, keys map[string]interface{}, prefixForAllResources bool) *hclwrite.Block {
	delete(dataMap, "resourceType")

	// ~~~~~~~~~~ add additional resource specifics here ~~~~~~~~~~

	if resourceType == "dataSet" {
		dataMap = dataSetSpecific(dataMap)
	} else if resourceType == "analysis" {
		dataMap = analysisSpecific(dataMap)
	} else if resourceType == "dashboard" {
		dataMap = dashboardSpecific(dataMap)
	} else if resourceType == "theme" {
		dataMap = themeSpecific(dataMap)
	} else if resourceType == "refreshSchedule" {
		dataMap = refreshScheduleSpecific(dataMap)
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	resBlockBody := resBlock.Body()

	resBlockBody.SetAttributeValue("aws_account_id", cty.StringVal(awsAccountId))
	addDependencies(resBlock.Body(), dataMap, resourceType, prefixForAllResources)
	createBlock(dataMap, keys, resBlockBody, resourceType)

	return resBlock
}

func createBlock(dataMap map[string]interface{}, keys map[string]interface{}, body *hclwrite.Body, resourceType string) {
	for k, v := range dataMap {
		resourceName := "awscc_quicksight_" + convertKey(resourceType)
		if keys[resourceName].(map[string]interface{})[convertKey(k)] == true { // compares each resource to the resource's individual schema
			k = convertKey(k)
		} else { // adds quotes in case of special characters in user assigned names
			k = "\"" + k + "\""
		}

		if nestedMap, ok := v.(map[string]interface{}); ok { // if nested map
			block := body.AppendNewBlock(k+" =", nil)
			createBlock(nestedMap, keys, block.Body(), resourceType)
		} else if nestedList, ok := v.([]interface{}); ok { // if nested list
			setAttributeList(body, k, nestedList)
		} else {
			switch v := v.(type) {
			case bool:
				body.SetAttributeValue(k, cty.BoolVal(v))
			case string:
				body.SetAttributeValue(k, cty.StringVal(v))
			case float64:
				body.SetAttributeValue(k, cty.NumberIntVal(int64(v)))
			default:
				fmt.Printf("unrecognized parameter value type: %T\n", v)
			}
		}
	}
}

func addDependencies(body *hclwrite.Body, dataMap map[string]interface{}, resourceType string, prefixForAllResources bool) {
	arns := getResourceDependencyArns(dataMap, resourceType, prefixForAllResources)
	if len(arns) != 0 {
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenOBrack,
				Bytes: []byte(`[`),
			},
		}
		for i := 0; i < len(arns); i++ {
			id := getId(arns[i].path)
			tokens = append(tokens, &hclwrite.Token{
				Type: hclsyntax.TokenIdent,

				Bytes: []byte(`awscc_quicksight_` + convertKey(arns[i].resource) + `.` + strings.ToLower(arns[i].resource) + `-` + id),
			})
			if i != len(arns)-1 {
				tokens = append(tokens, &hclwrite.Token{
					Type:  hclsyntax.TokenComma,
					Bytes: []byte(`,`),
				})
			}
		}
		tokens = append(tokens, &hclwrite.Token{
			Type:  hclsyntax.TokenOBrack,
			Bytes: []byte(`]`),
		})
		body.SetAttributeRaw("depends_on", tokens)
	}
}

func createTerraformFile(name string, hclFile *hclwrite.File) error {
	tfFile, err := os.Create(name)
	if err != nil {
		return fmt.Errorf("error creating terraform file: %v", err)
	}
	defer tfFile.Close()
	_, err = tfFile.Write(hclFile.Bytes())
	if err != nil {
		return fmt.Errorf("error writing to terraform file: %v", err)
	}
	return nil
}
