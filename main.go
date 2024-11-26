package parse_ims_metadata_txt

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var bom = []byte{0xef, 0xbb, 0xbf}

func ParseImsMetadatatxt(filePath string) map[string]interface{} {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
	defer file.Close()

	// Open UTF-8 BOM files correctly
	reader := bufio.NewReader(file)
	bom, _ := reader.Peek(3)
	if bytes.Equal(bom, []byte{0xEF, 0xBB, 0xBF}) {
		reader.Discard(3)
	}

	scanner := bufio.NewScanner(reader)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	metadata := make(map[string]interface{})
	for i, line := range lines {
		if strings.Contains(line, "=") {
			var key, value string
			if strings.Contains(line, "DisplayName=") {
				pairs := strings.Split(line, ", ")
				key = strings.Split(pairs[0], "=")[1] + " |" + strconv.Itoa(i)
				value = strings.Split(pairs[1], "=")[1]
				value = strings.Split(value, "}")[0]
			} else {
				parts := strings.SplitN(line, "=", 2)
				key = strings.TrimSpace(parts[0]) + " |" + strconv.Itoa(i)
				value = strings.TrimSpace(parts[1])
				value = strings.ReplaceAll(value, " ", "")
				value = strings.ReplaceAll(value, "\"", "")
			}

			path := []string{}
			tabCountLine := strings.Count(line, "\t")
			nextPathTabs := tabCountLine - 1
			for j := i - 1; j >= 0; j-- {
				tabCountPrevLine := strings.Count(lines[j], "\t")
				if strings.Contains(lines[j], "[") && tabCountPrevLine == nextPathTabs {
					newpath := strings.TrimSpace(lines[j])
					path = append([]string{newpath + " |" + fmt.Sprint(j)}, path...)
					nextPathTabs--
				}
			}

			currentDict := metadata
			for _, p := range path {
				p = strings.ReplaceAll(p, "[", "")
				p = strings.ReplaceAll(p, "]", "")
				if _, ok := currentDict[p]; !ok {
					currentDict[p] = make(map[string]interface{})
				}
				currentDict = currentDict[p].(map[string]interface{})
			}
			currentDict[key] = value
		}
	}
	metadata = fixDictKeys(metadata)
	return metadata
}

func fixDictKeys(d map[string]interface{}) map[string]interface{} {
	keyCounts := make(map[string]int)
	keyTotals := make(map[string]int)
	newDict := make(map[string]interface{})

	for key := range d {
		baseKey := strings.Split(key, " |")[0]
		keyTotals[baseKey]++
	}

	for key, value := range d {
		baseKey := strings.Split(key, " |")[0]
		keyCounts[baseKey]++

		var newKey string
		if keyTotals[baseKey] > 1 {
			newKey = fmt.Sprintf("%s {%d}", baseKey, keyCounts[baseKey])
		} else {
			newKey = baseKey
		}

		switch v := value.(type) {
		case map[string]interface{}:
			newDict[newKey] = fixDictKeys(v)
		default:
			newDict[newKey] = value
		}
	}

	return newDict
}

func processDirectory(dirPath string) {
	files, _ := os.ReadDir(dirPath)

	delPath := filepath.Join(dirPath, "del")
	os.MkdirAll(delPath, os.ModePerm)

	var protocolFiles, movedFiles, metadataFiles int
	toMove := make(map[string]bool)

	for _, file := range files {
		if strings.HasSuffix(file.Name(), "_metadata.txt") {
			metadataFiles++

			metadata := ParseImsMetadatatxt(filepath.Join(dirPath, file.Name()))

			// Write the YAML to a file (Removed for now)

			// Convert the map to YAML
			// y, err := yaml.Marshal(metadata)
			// if err != nil {
			// 	fmt.Printf("error: %v\n", err)
			// }

			// yaml_file_name := filepath.Join(dirPath, strings.TrimSuffix(file.Name(), ".txt")+".yaml")
			// fmt.Printf("yaml_file_name: %v\n", yaml_file_name)

			// err = os.WriteFile(yaml_file_name, y, 0644)
			// if err != nil {
			// 	fmt.Printf("error: %v\n", err)
			// }

			// Check if This image was aquired with
			// Aquire button -> Protocol Specification
			// Or by live/snap -> No Protocol Specification key
			if _, ok := metadata["Protocol Specification"]; ok {
				fmt.Printf("Keeping file: %v\n", file)
				protocolFiles++
			} else {
				base := strings.TrimSuffix(file.Name(), "_metadata.txt")
				toMove[base] = true
			}

			// This is how it was done before the parser was made
			// content, _ := os.ReadFile(filepath.Join(dirPath, file.Name()))

			// // Check for and remove BOM
			// content = bytes.TrimPrefix(content, bom)

			// if strings.Contains(string(content), "Protocol Name=") {
			// 	fmt.Printf("Keeping file: %v\n", file)
			// 	protocolFiles++
			// } else {
			// 	base := strings.TrimSuffix(file.Name(), "_metadata.txt")
			// 	toMove[base] = true
			// }

			// Extract essential metadata ans save to _processing_settings.yaml file

		}
	}

	for _, file := range files {
		base := strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))
		if toMove[base] {
			oldPath := filepath.Join(dirPath, file.Name())
			newPath := filepath.Join(delPath, file.Name())
			os.Rename(oldPath, newPath)
			fmt.Printf("Moved file: %s\n", file.Name())
			movedFiles++

			// Move the corresponding _metadata.txt file
			metadataFile := base + "_metadata.txt"
			oldMetadataPath := filepath.Join(dirPath, metadataFile)
			newMetadataPath := filepath.Join(delPath, metadataFile)
			os.Rename(oldMetadataPath, newMetadataPath)
		}
	}

	fmt.Printf("Found %d _metadata.txt files.\n", metadataFiles)
	fmt.Printf("Found %d files with 'Protocol Name='.\n", protocolFiles)
	fmt.Printf("Moved %d files.\n", movedFiles)
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter dragonfly directory path: ")
		fmt.Print("OR ENTER to exit (Not type any characters): ")

		dirPath, _ := reader.ReadString('\n')
		dirPath = strings.TrimSpace(dirPath) // Remove newline character

		if dirPath == "" {
			fmt.Println("No directory path provided. Exiting.")
			return
		}

		// If dirpath does not exist, continue to the next iteration
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			fmt.Println("Directory path does not exist. Skipping.")
			continue
		}

		// Check if the input is enclosed in quotes
		if (strings.HasPrefix(dirPath, "\"") && strings.HasSuffix(dirPath, "\"")) ||
			(strings.HasPrefix(dirPath, "'") && strings.HasSuffix(dirPath, "'")) {
			// Remove the quotes from the input
			dirPath = dirPath[1 : len(dirPath)-1]
		}

		processDirectory(dirPath)
	}
}
