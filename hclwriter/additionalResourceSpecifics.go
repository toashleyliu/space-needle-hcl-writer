package hclwriter

import (
	"time"
)

const scheduleValue = "schedule"
const startAfterDateTimeValue = "startAfterDateTime"

func dataSetSpecific(dataMap map[string]interface{}) map[string]interface{} {
	// "outputColumns" used in create/update api and CloudFormation
	updateKeyName(dataMap, pathOf("outputColumns"), 0, "", true)
	return dataMap
}

func analysisSpecific(dataMap map[string]interface{}) map[string]interface{} {
	// account for PropertyNamingStrategy.UPPER_CAMEL_CASE not processing correctly
	updateKeyName(dataMap, pathOf("definition.sheets.visuals.kpivisual"), 0, "KPIVisual", false)
	// still determining whether removing tooltipFields can be a permanent fix
	updateKeyName(dataMap, pathOf("definition.sheets.visuals.*.chartConfiguration.tooltip.fieldBasedTooltip.tooltipFields"), 0, "", true)
	return dataMap
}

func dashboardSpecific(dataMap map[string]interface{}) map[string]interface{} {
	// account for PropertyNamingStrategy.UPPER_CAMEL_CASE not processing correctly
	updateKeyName(dataMap, pathOf("definition.sheets.visuals.kpivisual"), 0, "KPIVisual", false)
	// still determining whether removing tooltipFields can be a permanent fix
	updateKeyName(dataMap, pathOf("definition.sheets.visuals.*.chartConfiguration.tooltip.fieldBasedTooltip.tooltipFields"), 0, "", true)
	return dataMap
}

func themeSpecific(dataMap map[string]interface{}) map[string]interface{} {
	// account for PropertyNamingStrategy.UPPER_CAMEL_CASE not processing correctly
	updateKeyName(dataMap, pathOf("configuration.uicolorPalette"), 0, "UIColorPalette", false)
	// "type" isn't part of terraform theme schema
	updateKeyName(dataMap, pathOf("type"), 0, "", true)
	return dataMap
}

func refreshScheduleSpecific(dataMap map[string]interface{}) map[string]interface{} {
	// account for PropertyNamingStrategy.UPPER_CAMEL_CASE not processing correctly
	updateKeyName(dataMap, pathOf("schedule.scheduleFrequency.timezone"), 0, "timeZone", false)
	// update "startAfterDateTime" attribute format from epoch to human-readable date
	switch value := dataMap[scheduleValue].(map[string]interface{})[startAfterDateTimeValue].(type) {
	case float64:
		if int64(value/1000) < time.Now().Unix()+1800 {
			// set "startAfterDateTime" to one hour from now if in the past or <30 mins from now
			value = float64(time.Now().Unix()+3600) * 1000
		}
		dataMap[scheduleValue].(map[string]interface{})[startAfterDateTimeValue] = time.UnixMilli(int64(value)).Format("2006-01-02T15:04:05Z")
	}

	return dataMap
}

func updateKeyName(dataMap interface{}, path []string, index int, newKey string, toDelete bool) {
	if index == len(path)-1 {
		if toDelete {
			delete(dataMap.(map[string]interface{}), path[index])
			return
		} else {
			if _, ok := dataMap.(map[string]interface{})[path[index]]; ok {
				dataMap.(map[string]interface{})[newKey] = dataMap.(map[string]interface{})[path[index]]
				delete(dataMap.(map[string]interface{}), path[index])
				return
			}
		}
	}
	if path[index] != "*" {
		switch value := dataMap.(map[string]interface{})[path[index]].(type) {
		case map[string]interface{}:
			updateKeyName(value, path, index+1, newKey, toDelete)
		case []interface{}:
			for _, elem := range value {
				updateKeyName(elem, path, index+1, newKey, toDelete)
			}
		}
	} else { // * in path accounts for all visuals
		switch value := dataMap.(type) {
		case []interface{}:
			for _, elem := range value {
				updateKeyName(elem, path, index+1, newKey, toDelete)
			}
		case map[string]interface{}:
			for _, elem := range value {
				updateKeyName(elem, path, index+1, newKey, toDelete)
			}
		}
	}
}
