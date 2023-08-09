package hclwriter

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var supportedResources = map[string]bool{
	"awscc_quicksight_data_source":      true,
	"awscc_quicksight_data_set":         true,
	"awscc_quicksight_analysis":         true,
	"awscc_quicksight_dashboard":        true,
	"awscc_quicksight_theme":            true,
	"awscc_quicksight_refresh_schedule": true,
	"awscc_quicksight_vpc_connection":   true,
}

type Args struct {
	AccountId             string     // aws account id (required)
	Region                string     // aws account region (required)
	Name                  string     // terraform file name (optional)
	SrcPath               string     // path containing jsons (required)
	DestPath              string     // path to put terraforms (required)
	OverrideProperties    arrayFlags // array of properties to override (optional)
	PrefixForAllResources bool       // prefex for all resource ids (optional)
}

type arrayFlags []string

func ReadCmdLineArgs() (*Args, bool, error) {
	var overrideProperties arrayFlags

	flag.Usage = func() {
		fmt.Printf("Usage: %s --aws-account-id <account-id> --region <region> --src <src> --dest <dest> [--override <\"propertyName, resourceArn\"] [--prefix-for-all-resources] [--terraform-file-name <terraform-file-name>]\n>", os.Args[0])
	}

	accountId := flag.String("aws-account-id", "", "The ID of the AWS account that owns the resource (Required)")
	region := flag.String("region", "", "The AWS account region (Required)")
	terraformFileName := flag.String("terraform-file-name", "convertedFile.tf", "Converted Terraform file name (Optional)")
	srcPath := flag.String("src", "", "Path containing JSONs (Required)")
	destPath := flag.String("dest", "", "Path to put converted Terraform file (Required)")
	prefix := flag.Bool("prefix-for-all-resources", false, "Prefix to add to all resource IDs (Optional)")
	help := flag.Bool("help", false, "Display usage information")

	flag.Var(&overrideProperties, "override", "Pairs of properties to override and their resource arn (Optional)")

	flag.Parse()

	if *help {
		flag.Usage()
		return nil, true, nil
	}

	if *accountId == "" {
		fmt.Println("aws account id is required")
		flag.Usage()
		flag.PrintDefaults()
		return nil, false, fmt.Errorf("invalid parameter entered")
	}

	if *region == "" {
		fmt.Println("aws account region is required")
		flag.Usage()
		flag.PrintDefaults()
		return nil, false, fmt.Errorf("invalid parameter entered")
	}

	if *srcPath == "" {
		fmt.Println("source path is required")
		flag.Usage()
		flag.PrintDefaults()
		return nil, false, fmt.Errorf("invalid parameter entered")
	}

	if *destPath == "" {
		fmt.Println("dest path is required")
		flag.Usage()
		flag.PrintDefaults()
		return nil, false, fmt.Errorf("invalid parameter entered")
	}

	if *prefix {
		overrideProperties = append(overrideProperties, "prefix-for-all-resources")
	}

	for _, argument := range overrideProperties {
		argument = strings.ReplaceAll(argument, " ", "")
		regex, err := regexp.Compile(`^\w+,[\w\-:/0-9]+`)
		if err != nil {
			return nil, false, fmt.Errorf("error compiling regex: %v", err)
		}

		if !regex.MatchString(argument) && argument != "prefix-for-all-resources" {
			return nil, false, fmt.Errorf("invalid parameter override: %s", argument)
		} else {
			continue
		}
	}

	return &Args{
		AccountId:             *accountId,
		Region:                *region,
		Name:                  *terraformFileName,
		SrcPath:               *srcPath,
		DestPath:              *destPath,
		OverrideProperties:    overrideProperties,
		PrefixForAllResources: *prefix,
	}, false, nil
}

func ProcessFiles(accountId, srcPath, terraformFileName, destPath string, overrideParameters arrayFlags, prefixForAllResources bool) error {
	err := os.MkdirAll(destPath, os.ModePerm)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(srcPath)
	if err != nil {
		return err
	}

	schemaJson, _ := getTFProviderSchema()
	schemaMap, _ := createSchemaMap(schemaJson)

	return convertFiles(files, srcPath, destPath, terraformFileName, accountId, schemaMap, overrideParameters, prefixForAllResources)
}

func getTFProviderSchema() ([]byte, error) {
	cmd := exec.Command("terraform", "providers", "schema", "-json")
	fmt.Println("Executing command:", cmd.String())

	var res bytes.Buffer
	cmd.Stdout = &res

	if err := cmd.Start(); err != nil {
		fmt.Println("Error running terraform providers schema -json: ", err)
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		fmt.Println("Error waiting for terraform providers schema -json: ", err)
		return nil, err
	}

	return res.Bytes(), nil
}

func createSchemaMap(schemaJson []byte) (map[string]interface{}, error) {
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schemaJson, &schemaMap); err != nil {
		fmt.Println("Error parsing schema json: ", err)
		return nil, err
	}
	keys := make(map[string]interface{})
	extractKeys(schemaMap, keys)
	return keys, nil
}

func extractKeys(data map[string]interface{}, keys map[string]interface{}) {
	for key := range data {
		if _, ok := supportedResources[key]; ok {
			nestedMap := data[key].(map[string]interface{})
			keys[key] = make(map[string]interface{})
			extractKeys(nestedMap, keys[key].(map[string]interface{}))
		} else {
			keys[key] = true
			if nestedMap, ok := data[key].(map[string]interface{}); ok {
				extractKeys(nestedMap, keys)
			}
		}
	}
}
