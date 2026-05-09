package client

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var enumValuePattern = regexp.MustCompile(`\[(\-?\d+(?:\.\d+)?)\]`)
var p101RecordPattern = regexp.MustCompile(`^([0-9]+)~[^;]*;[^@]*@(.+)$`)

type Sample struct {
	FieldID int
	Name    string
	Value   float64
	Unit    string
}

// ParseSamples parses the p05_equip_sample.cgi response body into field-id keyed samples.
func ParseSamples(raw string) (map[int]Sample, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("empty response body")
	}

	if strings.HasPrefix(trimmed, "Y|") {
		samples, err := parseP101Samples(trimmed)
		if err != nil {
			return nil, err
		}
		if len(samples) > 0 {
			return samples, nil
		}
	}

	records := strings.Split(trimmed, ";")
	samples := make(map[int]Sample, len(records))

	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" {
			continue
		}

		sample, ok := parseRecord(record)
		if !ok {
			continue
		}
		samples[sample.FieldID] = sample
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("no samples parsed from response: %s", previewResponse(trimmed))
	}

	return samples, nil
}

func parseP101Samples(raw string) (map[int]Sample, error) {
	records := strings.Split(strings.TrimPrefix(raw, "Y|"), "|")
	samples := make(map[int]Sample, len(records))

	for _, record := range records {
		record = strings.TrimSpace(record)
		if record == "" || strings.Contains(record, ".gif") {
			continue
		}

		matches := p101RecordPattern.FindStringSubmatch(record)
		if len(matches) != 3 {
			continue
		}

		fieldID, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		value, err := parseValue(strings.TrimSpace(matches[2]))
		if err != nil {
			continue
		}

		samples[fieldID] = Sample{
			FieldID: fieldID,
			Value:   value,
		}
	}

	if len(samples) == 0 {
		return nil, fmt.Errorf("no p101 samples parsed from response: %s", previewResponse(raw))
	}

	return samples, nil
}

func parseRecord(record string) (Sample, bool) {
	fields := strings.Split(record, ",")
	if len(fields) < 4 {
		return Sample{}, false
	}

	// The first p05 record includes device metadata before the real field id:
	// 3021,AC_1,ENP_AC_SRVII[COM]^2,Return air temperature measurement,28.6,℃
	if len(fields) >= 6 && strings.Contains(fields[2], "^") {
		fieldID, err := parseFieldID(fields[2])
		if err != nil {
			return Sample{}, false
		}
		value, err := parseValue(strings.TrimSpace(fields[4]))
		if err != nil {
			return Sample{}, false
		}
		return Sample{
			FieldID: fieldID,
			Name:    strings.TrimSpace(fields[3]),
			Value:   value,
			Unit:    strings.TrimSpace(fields[5]),
		}, true
	}

	fieldID, err := parseFieldID(fields[0])
	if err != nil {
		return Sample{}, false
	}
	value, err := parseValue(strings.TrimSpace(fields[2]))
	if err != nil {
		return Sample{}, false
	}

	unit := ""
	if len(fields) > 3 {
		unit = strings.TrimSpace(fields[3])
	}

	return Sample{
		FieldID: fieldID,
		Name:    strings.TrimSpace(fields[1]),
		Value:   value,
		Unit:    unit,
	}, true
}

func parseFieldID(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if idx := strings.LastIndex(raw, "^"); idx >= 0 {
		raw = raw[idx+1:]
	}
	return strconv.Atoi(raw)
}

func parseValue(raw string) (float64, error) {
	if raw == "" {
		return 0, fmt.Errorf("empty value")
	}

	if value, err := strconv.ParseFloat(raw, 64); err == nil {
		return value, nil
	}

	matches := enumValuePattern.FindStringSubmatch(raw)
	if len(matches) == 2 {
		return strconv.ParseFloat(matches[1], 64)
	}

	return 0, fmt.Errorf("unsupported sample value %q", raw)
}

func previewResponse(raw string) string {
	raw = strings.ReplaceAll(raw, "\n", " ")
	raw = strings.ReplaceAll(raw, "\r", " ")
	raw = strings.TrimSpace(raw)
	if len(raw) > 240 {
		raw = raw[:240] + "..."
	}
	return strconv.Quote(raw)
}
