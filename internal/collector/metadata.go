package collector

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var metricRowPattern = regexp.MustCompile("^\\| `([^`]+)` \\| ([0-9]+) \\| (.+?) \\|")

//go:embed default_metrics.md
var defaultMetrics string

type MetricDefinition struct {
	FieldID int
	Name    string
	Help    string
}

func LoadMetricDefinitions(path string) (map[int]MetricDefinition, error) {
	if path == "" {
		return loadMetricDefinitionsFromReader(strings.NewReader(defaultMetrics))
	}

	file, err := os.Open(path)
	if err != nil {
		return loadMetricDefinitionsFromReader(strings.NewReader(defaultMetrics))
	}
	defer file.Close()

	return loadMetricDefinitionsFromReader(file)
}

func loadMetricDefinitionsFromReader(reader io.Reader) (map[int]MetricDefinition, error) {
	definitions := make(map[int]MetricDefinition)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := metricRowPattern.FindStringSubmatch(line)
		if len(matches) != 4 {
			continue
		}

		fieldID, err := strconv.Atoi(matches[2])
		if err != nil {
			continue
		}

		definitions[fieldID] = MetricDefinition{
			FieldID: fieldID,
			Name:    matches[1],
			Help:    sanitizeHelp(matches[3]),
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan metrics metadata: %w", err)
	}
	if len(definitions) == 0 {
		return nil, fmt.Errorf("no metric definitions found")
	}

	return definitions, nil
}

func sanitizeHelp(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "|")
	raw = strings.ReplaceAll(raw, "`", "")
	return raw
}
