package engine

import (
	"encoding/json"
	"muu-alpha/backend/pi"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v3"
)

type Task struct {
	ID               string          `json:"id"`
	Entry            string          `json:"entry"`
	PipelineOverride json.RawMessage `json:"pipeline_override"`
	StartedAt        time.Time       `json:"started_at"`
	FinishedAt       time.Time       `json:"finished_at"`
	Status           maa.Status      `json:"status"`
}

// GetTaskList gets the list of selected tasks, merging all PipelineOverride
func GetTaskList() []*Task {
	piService := pi.PI()
	config := piService.GetConfig()
	v2Loaded := piService.V2Loaded()

	tasks := make([]*Task, 0)

	if config == nil || v2Loaded == nil || v2Loaded.Interface == nil {
		return tasks
	}

	iface := v2Loaded.Interface

	// Build V2Task mapping for quick lookup
	taskMap := make(map[string]*pi.V2Task)
	for i := range iface.Task {
		taskMap[iface.Task[i].Name] = &iface.Task[i]
	}

	// Iterate over selected tasks in the configuration
	for _, configTask := range config.Task {
		if !configTask.Checked {
			continue
		}

		v2Task, exists := taskMap[configTask.Name]
		if !exists {
			continue
		}

		// Merge PipelineOverride
		mergedOverride := mergePipelineOverrides(v2Task, configTask.Option, iface.Option)

		// Serialize merged PipelineOverride
		var overrideJSON json.RawMessage
		if mergedOverride != nil {
			if data, err := json.Marshal(mergedOverride); err == nil {
				overrideJSON = data
			}
		}

		tasks = append(tasks, &Task{
			ID:               configTask.ID,
			Entry:            v2Task.Entry,
			PipelineOverride: overrideJSON,
		})
	}

	return tasks
}

// mergePipelineOverrides merges PipelineOverride for the task and its options
func mergePipelineOverrides(v2Task *pi.V2Task, configOptions []pi.ConfigTaskOption, optionDefs map[string]pi.V2Option) map[string]map[string]interface{} {
	merged := make(map[string]map[string]interface{})

	// 1. First merge PipelineOverride for the task itself
	if len(v2Task.PipelineOverride) > 0 {
		var taskOverride map[string]map[string]interface{}
		if err := json.Unmarshal(v2Task.PipelineOverride, &taskOverride); err == nil {
			mergeOverride(merged, taskOverride)
		}
	}

	// 2. Build configuration option value mapping
	optionValues := make(map[string]string)
	for _, opt := range configOptions {
		optionValues[opt.Name] = opt.Value
	}

	// 3. Recursively merge PipelineOverride for options
	collectOptionOverrides(merged, v2Task.Option, optionValues, optionDefs)

	if len(merged) == 0 {
		return nil
	}

	return merged
}

// collectOptionOverrides recursively collects PipelineOverride for options
func collectOptionOverrides(merged map[string]map[string]interface{}, optionNames []string, optionValues map[string]string, optionDefs map[string]pi.V2Option) {
	for _, optName := range optionNames {
		optDef, exists := optionDefs[optName]
		if !exists {
			continue
		}

		optType := optDef.GetType()

		switch optType {
		case "select", "switch":
			// Get current selected value
			selectedValue := optionValues[optName]
			if selectedValue == "" {
				continue
			}

			// Find selected case
			for _, optCase := range optDef.Cases {
				if optCase.Name == selectedValue {
					// Merge PipelineOverride for this case
					if len(optCase.PipelineOverride) > 0 {
						var caseOverride map[string]map[string]interface{}
						if err := json.Unmarshal(optCase.PipelineOverride, &caseOverride); err == nil {
							mergeOverride(merged, caseOverride)
						}
					}

					// Recursively process nested options
					if len(optCase.Option) > 0 {
						collectOptionOverrides(merged, optCase.Option, optionValues, optionDefs)
					}
					break
				}
			}

		case "input":
			// For input type, process option-level PipelineOverride (with variable replacement)
			if len(optDef.PipelineOverride) > 0 {
				// Collect input values
				inputValues := make(map[string]inputValue)
				for _, input := range optDef.Inputs {
					key := optName + "." + input.Name
					value := optionValues[key]
					if value == "" {
						value = input.Default
					}
					inputValues[input.Name] = inputValue{
						value:        value,
						pipelineType: input.PipelineType,
					}
				}

				// Replace variables and merge
				var inputOverride map[string]map[string]interface{}
				if err := json.Unmarshal(optDef.PipelineOverride, &inputOverride); err == nil {
					// Replace variables
					processedOverride := replaceVariables(inputOverride, inputValues)
					mergeOverride(merged, processedOverride)
				}
			}
		}
	}
}

// inputValue stores value and pipelineType for input
type inputValue struct {
	value        string
	pipelineType string
}

// replaceVariables replaces variable placeholders in PipelineOverride
func replaceVariables(override map[string]map[string]interface{}, inputValues map[string]inputValue) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	for taskName, taskProps := range override {
		result[taskName] = make(map[string]interface{})
		for key, value := range taskProps {
			result[taskName][key] = replaceValue(value, inputValues)
		}
	}

	return result
}

// replaceValue recursively replaces variables in the value
func replaceValue(value interface{}, inputValues map[string]inputValue) interface{} {
	switch v := value.(type) {
	case string:
		// Check if it is a complete variable reference "{varName}"
		re := regexp.MustCompile(`^\{(\w+)\}$`)
		if matches := re.FindStringSubmatch(v); len(matches) == 2 {
			varName := matches[1]
			if input, exists := inputValues[varName]; exists {
				// Convert type based on pipeline_type
				return convertValue(input.value, input.pipelineType)
			}
		}

		// Replace variables in the string
		result := v
		for name, input := range inputValues {
			placeholder := "{" + name + "}"
			result = strings.ReplaceAll(result, placeholder, input.value)
		}
		return result

	case []interface{}:
		newArr := make([]interface{}, len(v))
		for i, item := range v {
			newArr[i] = replaceValue(item, inputValues)
		}
		return newArr

	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for k, val := range v {
			newMap[k] = replaceValue(val, inputValues)
		}
		return newMap

	default:
		return value
	}
}

// convertValue converts value based on pipelineType
func convertValue(value string, pipelineType string) interface{} {
	switch pipelineType {
	case "int":
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
		return value
	case "bool":
		return value == "true" || value == "True" || value == "1"
	default:
		return value
	}
}

// mergeOverride merges two PipelineOverrides, with the new one overriding top-level keys of the old one
func mergeOverride(base, override map[string]map[string]interface{}) {
	for taskName, taskProps := range override {
		if _, exists := base[taskName]; !exists {
			base[taskName] = make(map[string]interface{})
		}
		for key, value := range taskProps {
			base[taskName][key] = value
		}
	}
}
