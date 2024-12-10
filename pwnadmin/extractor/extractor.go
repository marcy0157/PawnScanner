package extractor

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// ExtractEmailsFromFile estrae email valide da un file di testo.
func ExtractEmailsFromFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var emails []string
	scanner := bufio.NewScanner(file)
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	emailValidationRegex := regexp.MustCompile(`^[a-zA-Z0-9.%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := emailRegex.FindAllString(line, -1)
		for _, email := range matches {
			sanitized := strings.TrimSpace(strings.TrimRight(email, ":;"))
			if emailValidationRegex.MatchString(sanitized) {
				emails = append(emails, sanitized)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return emails, nil
}
