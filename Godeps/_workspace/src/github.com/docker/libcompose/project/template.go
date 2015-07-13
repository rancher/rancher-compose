package project

import (
	"os"
	"fmt"
	"bufio"
	"regexp"
	"sort"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	templateVarRegex = regexp.MustCompile("\\%{[A-Za-z0-9-_\\.\\+]+}")
)

func ProcessTemplate(datas rawServiceMap, answers map[string]string) error {
	templateVariablesMap := extractFieldDefinitions(datas)
	err := gatherInput(templateVariablesMap, answers)
	if err != nil {
		return err
	}
	for name, data := range datas {
		datas[name] = replace(data, templateVariablesMap).(rawService)
	}
	return nil
}

func extractFieldDefinitions(datas rawServiceMap) (templateVariablesMap map[string]*UserFieldDefinition) {
	templateVariablesMap = make(map[string]*UserFieldDefinition)

	var fieldDefinitionSection map[interface{}]interface{}
	for _, data := range datas {
		if data == nil {
			continue
		}
		templateVariables := extractVariables(data)
		for i := range templateVariables {
			varName := templateVariables[i].Name
			templateVariablesMap[varName] = &templateVariables[i]
		}
		if data["user_data"] != nil {
			// process this after we've collected the full list from examining the template file
			fieldDefinitionSection = data["user_data"].(map[interface{}]interface{})
		}
	}

	if fieldDefinitionSection != nil {
		for key, value := range fieldDefinitionSection {
			fieldDef := templateVariablesMap[key.(string)]
			if fieldDef == nil {
				fieldDef := UserFieldDefinition{}
				fieldDef.Name = key.(string)
				templateVariablesMap[key.(string)] = &fieldDef
			}

			fieldDefMap := value.(map[interface{}]interface{})

			if asString(fieldDefMap["description"]) != "" {
				fieldDef.Description = asString(fieldDefMap["description"])
			}
			if asString(fieldDefMap["type"]) != "" {
				fieldDef.Type = asString(fieldDefMap["type"])
			}
			if asString(fieldDefMap["default"]) != "" {
				fieldDef.Default = asString(fieldDefMap["default"])
			}
		}
	}
	return templateVariablesMap
}

// TODO: Fix to exclude escaped literals like \%{foo}
func extractVariables(line interface{}) []UserFieldDefinition {
	var subVars []UserFieldDefinition
	switch line.(type) {
		case string:
			variables := templateVarRegex.FindAllString(line.(string), -1)
			userFieldDefs := make([]UserFieldDefinition, len(variables))
			for i := range variables {
				fieldDef := UserFieldDefinition{}
				fieldDef.Name = variables[i][2:len(variables[i])-1]
				userFieldDefs[i] = fieldDef
			}
			return userFieldDefs
		case []interface{}:
			for i := range line.([]interface{}) {
				subSubVars := extractVariables(line.([]interface{})[i])
				for j := range subSubVars {
					subVars = append(subVars, subSubVars[j])
				}
			}
		case map[interface{}]interface{}:
			for _, data := range line.(map[interface{}]interface{}) {
				subSubVars := extractVariables(data)
				for i := range subSubVars {
					subVars = append(subVars, subSubVars[i])
				}
			}
		case rawService:
			for _, data := range line.(rawService) {
				subSubVars := extractVariables(data)
				for i := range subSubVars {
					subVars = append(subVars, subSubVars[i])
				}
			}
	}
	return subVars
}

func gatherInput(templateVariablesMap map[string]*UserFieldDefinition, answers map[string]string) error {
	sortedKeys := make([]string, len(templateVariablesMap))
	i := 0
	for k, _ := range templateVariablesMap {
		sortedKeys[i] = k
		i++
	}
	sort.Strings(sortedKeys)

	reader := bufio.NewReader(os.Stdin)
	for i := range sortedKeys {
		key := sortedKeys[i]
		fieldDef := templateVariablesMap[key]
		if val, ok := answers[key]; ok {
			fieldDef.Value = val
		} else {
			if fieldDef.Description == "" {
				fieldDef.Description = "Please enter value for " + key
			}
			fmt.Printf("%s: [%s] ", fieldDef.Description, fieldDef.Default)
			// TODO: Handle other types
			if fieldDef.Type == "password" {
				// TODO: For password, we might want to have them enter it
				// twice to make sure it was entered correctly
				value, err := terminal.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return err
				}
				fieldDef.Value = string(value)
				fmt.Printf("\n")
			} else {
				value, err := reader.ReadString('\n')
				if err != nil {
					return err
				}
				value = value[:len(value) - 1]
				if value == "" {
					fieldDef.Value = fieldDef.Default
				} else {
					fieldDef.Value = value
				}
			}
		}
	}
	return nil
}

// TODO: Handle escaped literals
func replace(line interface{}, values map[string]*UserFieldDefinition) interface{} {
	switch line.(type) {
		case string:
			for key, value := range values {
				re := regexp.MustCompile("\\%{" + key + "}")
				line = re.ReplaceAllString(line.(string), value.Value)
			}
		case []interface{}:
			for i := range line.([]interface{}) {
				line.([]interface{})[i] = replace(line.([]interface{})[i], values)
			}
		case map[interface{}]interface{}:
			for key, data := range line.(map[interface{}]interface{}) {
				line.(map[interface{}]interface{})[key] = replace(data, values)
			}
		case rawService:
			for key, data := range line.(rawService) {
				line.(rawService)[key] = replace(data, values)
			}
	}

	return line
}
