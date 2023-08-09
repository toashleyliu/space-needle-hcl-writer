package hclwriter

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const defaultValue = "default"

func (i *arrayFlags) String() string {
	return strings.Join(*i, " ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func readJson(jsonFileName string) ([]byte, error) {
	jsonFile, err := os.Open(jsonFileName)
	if err != nil {
		return nil, fmt.Errorf("error opening json file: %v", err)
	}
	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, fmt.Errorf("error reading json file: %v", err)
	}
	return jsonData, nil
}

func createDataMap(jsonData []byte) (map[string]interface{}, error) {
	dataMap := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &dataMap); err != nil {
		return nil, fmt.Errorf("error parsing json file: %v", err)
	}
	return dataMap, nil
}

func convertKey(key string) string {
	var s strings.Builder
	s.WriteByte(key[0])
	if len(key) > 1 {
		for i := 1; i < len(key)-1; i++ {
			if key[i] >= 'A' && key[i] <= 'Z' {
				if (key[i+1] >= 'A' && key[i+1] <= 'Z') &&
					(key[i-1] >= 'A' && key[i-1] <= 'Z') {
					s.WriteByte(key[i])
				} else {
					s.WriteRune('_')
					s.WriteByte(key[i])
				}
			} else {
				s.WriteByte(key[i])
			}
		}
		s.WriteByte(key[len(key)-1])
	}

	return strings.ToLower(s.String())
}

func getId(resourceTypeArn string) string {
	lastSlashIndex := strings.LastIndex(resourceTypeArn, "/")
	return resourceTypeArn[lastSlashIndex+1:]
}

func cloneMap(dataMap map[string]interface{}) map[string]interface{} {
	clonedDataMap := make(map[string]interface{})
	for k, v := range dataMap {
		clonedDataMap[k] = cloneItem(v)
	}
	return clonedDataMap
}

func cloneItem(v interface{}) interface{} {
	switch value := v.(type) {
	case map[string]interface{}:
		return cloneMap(value)
	case []interface{}:
		nestedList := make([]interface{}, len(value))
		for i, x := range value {
			nestedList[i] = cloneItem(x)
		}
		return nestedList
	default:
		return value
	}
}

func getLastDotIndex(path string) int {
	return strings.LastIndex(path, ".")
}

func splitOn(str, char string) []string {
	return strings.Split(str, char)
}

func pathOf(path string) []string {
	return strings.Split(path, ".")
}
