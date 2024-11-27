package parse_ims_metadata_txt

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

func GetAllMetadata(filePath string) (map[string]interface{}, error) {
	// (code omitted for brevity: import statements, etc.)

	content, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	parsedContent, err := parseContent(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %w", err)
	}

	cleanedContent := cleanIDStrings(parsedContent)
	cleanedContent = removeBrackets(cleanedContent)             // Update cleanedContent with the clean version from removeBrackets
	cleanedContent = fixFeatureValueListStrings(cleanedContent) // New function call

	//fmt.Printf("cleanedContent: %v\n", cleanedContent)

	return cleanedContent, nil
}

func MakeYaml(filePath string, outputPath string) error {
	cleanedContent, err := GetAllMetadata(filePath)
	if err != nil {
		return fmt.Errorf("failed to get all metadata: %w", err)
	}

	err = writeYAML(outputPath, cleanedContent)
	if err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	fmt.Println("YAML file created successfully.")
	return nil
}

// Functions to retrieve values from cleanedContent
// Extract values (used for debugging or further processing)
/*
	protocol_name := cleanedContent["Protocol Name"]
	Height := cleanedContent["Height"]
	fmt.Printf("Height: %v\n", Height)
	Width := cleanedContent["Width"]
	fmt.Printf("Width: %v\n", Width)
	NumberOfChannels := cleanedContent["NumberOfChannels"]
	fmt.Printf("NumberOfChannels: %v\n", NumberOfChannels)
	NumberOfTimePoints := cleanedContent["NumberOfTimePoints"]
	fmt.Printf("NumberOfTimePoints: %v\n", NumberOfTimePoints)
	NumberOfZPoints := cleanedContent["NumberOfZPoints"]
	fmt.Printf("NumberOfZPoints: %v\n", NumberOfZPoints)
	// Replace Wizard with updated logic
	fmt.Printf("protocol_name: %v\n", protocol_name)
*/
func GetProtocolName(filePath string) (string, error) {
	cleanedContent, err := GetAllMetadata(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to get all metadata: %w", err)
	}

	protocolName, ok := cleanedContent["Protocol Name"]
	if !ok {
		return "", fmt.Errorf("protocol name not found in metadata")
	}

	protocolNameStr, ok := protocolName.(string)
	if !ok {
		return "", fmt.Errorf("protocol name is not a string")
	}

	return protocolNameStr, nil
}

func GetHeight(filePath string) (interface{}, error) {
	cleanedContent, err := GetAllMetadata(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get all metadata: %w", err)
	}

	height, ok := cleanedContent["Height"]
	if !ok {
		return 0, fmt.Errorf("height not found in metadata")
	}

	return height, nil
}

func GetWidth(filePath string) (interface{}, error) {
	cleanedContent, err := GetAllMetadata(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get all metadata: %w", err)
	}

	width, ok := cleanedContent["Width"]
	if !ok {
		return 0, fmt.Errorf("width not found in metadata")
	}

	return width, nil
}

func GetNumberOfChannels(filePath string) (interface{}, error) {
	cleanedContent, err := GetAllMetadata(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get all metadata: %w", err)
	}

	numChannels, ok := cleanedContent["NumberOfChannels"]
	if !ok {
		return 0, fmt.Errorf("number of channels not found in metadata")
	}

	return numChannels, nil
}

func GetNumberOfTimePoints(filePath string) (interface{}, error) {
	cleanedContent, err := GetAllMetadata(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get all metadata: %w", err)
	}

	numTimePoints, ok := cleanedContent["NumberOfTimePoints"]
	if !ok {
		return 0, fmt.Errorf("number of time points not found in metadata")
	}

	return numTimePoints, nil
}

func GetNumberOfZPoints(filePath string) (interface{}, error) {
	cleanedContent, err := GetAllMetadata(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get all metadata: %w", err)
	}

	numZPoints, ok := cleanedContent["NumberOfZPoints"]
	if !ok {
		return 0, fmt.Errorf("number of Z points not found in metadata")
	}

	return numZPoints, nil
}

func fixFeatureValueListStrings(data map[string]interface{}) map[string]interface{} {
	fixedData := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case []interface{}:
			var fixedList []interface{}
			for _, item := range v {
				if str, ok := item.(string); ok {
					split := strings.SplitN(str, ", Value=", 2)
					if len(split) == 2 {
						cleanedKey := strings.TrimSpace(split[0])
						cleanedValue := strings.TrimSpace(split[1])
						fixedList = append(fixedList, map[string]interface{}{cleanedKey: cleanedValue})
					} else {
						fixedList = append(fixedList, item)
					}
				} else {
					fixedList = append(fixedList, item)
				}
			}
			fixedData[key] = fixedList
		case map[string]interface{}:
			fixedData[key] = fixFeatureValueListStrings(v)
		default:
			fixedData[key] = value
		}
	}

	return fixedData
}

func removeBrackets(data map[string]interface{}) map[string]interface{} {
	cleanedData := make(map[string]interface{}) // Create a new map to store cleaned data

	for key, value := range data {
		cleanedKey := trimBrackets(key) // Clean key brackets

		switch v := value.(type) {
		case string:
			cleanedData[cleanedKey] = trimBrackets(v) // Clean string values
		case map[string]interface{}:
			cleanedData[cleanedKey] = removeBrackets(v) // Recursively clean nested maps
		case []interface{}:
			cleanedList := make([]interface{}, len(v))
			for i, item := range v {
				switch item := item.(type) {
				case string:
					cleanedList[i] = trimBrackets(item)
				case map[string]interface{}:
					cleanedList[i] = removeBrackets(item) // Recursively clean nested maps in slices
				default:
					cleanedList[i] = item
				}
			}
			cleanedData[cleanedKey] = cleanedList // Update the cleaned list
		default:
			cleanedData[cleanedKey] = value // Preserve other types of values as-is
		}
	}

	return cleanedData // Return the cleaned map
}

func trimBrackets(s string) string {
	s = strings.TrimSpace(s) // Let's trim space to handle cases with spaces outside brackets
	if strings.HasPrefix(s, "[") || strings.HasSuffix(s, "]") {
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")
	}
	if strings.HasPrefix(s, "{") || strings.HasSuffix(s, "}") {
		s = strings.TrimPrefix(s, "{")
		s = strings.TrimSuffix(s, "}")
	}
	if strings.HasPrefix(s, "(") || strings.HasSuffix(s, ")") {
		s = strings.TrimPrefix(s, "(")
		s = strings.TrimSuffix(s, ")")
	}
	return s
}
func readFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		content.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return content.String(), nil
}

// parseContent parses the file content and puts every key-value pair in a list.
func parseContent(content string) (map[string]interface{}, error) {
	lines := strings.Split(content, "\n")
	root := make(map[string]interface{})
	stack := []map[string]interface{}{root}
	indentStack := []int{0}

	for _, line := range lines {
		if line == "" {
			continue
		}

		indentLevel, trimmedLine := countIndent(line)
		for len(indentStack) > 0 && indentStack[len(indentStack)-1] >= indentLevel {
			stack = stack[:len(stack)-1]
			indentStack = indentStack[:len(indentStack)-1]
		}

		if len(stack) == 0 {
			stack = append(stack, root)
			indentStack = append(indentStack, 0)
		}

		current := stack[len(stack)-1]

		if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
			sectionName := trimmedLine
			newSection := make(map[string]interface{})
			if _, exists := current[sectionName]; !exists {
				current[sectionName] = []interface{}{}
			}
			current[sectionName] = append(current[sectionName].([]interface{}), newSection)
			stack = append(stack, newSection)
			indentStack = append(indentStack, indentLevel)
		} else {
			key, value := parseKeyValue(trimmedLine)
			if _, exists := current[key]; !exists {
				current[key] = []interface{}{}
			}
			current[key] = append(current[key].([]interface{}), value)
		}
	}

	return root, nil
}

func countIndent(line string) (int, string) {
	indentLevel := 0
	for i := 0; i < len(line); i++ {
		if line[i] != ' ' && line[i] != '\t' {
			break
		}
		indentLevel++
	}

	return indentLevel, strings.TrimSpace(line)
}

func parseKeyValue(line string) (string, string) {
	parts := strings.SplitN(line, "=", 2)
	key := strings.TrimSpace(parts[0])
	value := ""
	if len(parts) > 1 {
		value = strings.TrimSpace(parts[1])
	}
	return key, value
}

var numberSuffixRegex = regexp.MustCompile(`^(.*?)(\d+)$`)

func cleanIDStrings(data map[string]interface{}) map[string]interface{} {
	cleanedData := make(map[string]interface{})
	for key, value := range data {
		// Strip the number suffix for generic processing
		baseKey := stripNumberSuffix(key)

		// Clean nested maps
		switch v := value.(type) {
		case []interface{}:
			if len(v) == 1 {
				// If there's just one item in the list, store it directly
				cleanedData[baseKey] = cleanValue(v[0])
			} else {
				// Otherwise, process each item in the list
				var cleanedList []interface{}
				for _, item := range v {
					cleanedList = append(cleanedList, cleanValue(item))
				}
				cleanedData[baseKey] = cleanedList
			}
		default:
			cleanedData[baseKey] = cleanValue(v)
		}
	}

	// Simplify lists that contain only one map entry
	for key, value := range cleanedData {
		cleanedData[key] = simplifyList(value)
	}

	return cleanedData
}

func stripNumberSuffix(key string) string {
	matches := numberSuffixRegex.FindStringSubmatch(key)
	if len(matches) != 3 {
		return key
	}
	return matches[1]
}

func cleanValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		return cleanIDStrings(v)
	default:
		return value
	}
}

func simplifyList(value interface{}) interface{} {
	switch v := value.(type) {
	case []interface{}:
		if len(v) == 1 {
			return v[0]
		}
		for i, item := range v {
			v[i] = cleanValue(item)
		}
	}
	return value
}

func writeYAML(outputPath string, data map[string]interface{}) error {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, bytes, 0644)
}
