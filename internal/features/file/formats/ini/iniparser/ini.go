/*
Package parsers provides optimized INI file parsing with structure preservation.

This parser maintains the exact order and structure of INI files by returning
a slice of INILine objects rather than a map. This preserves:
- Line order
- Comments and their positions
- Empty lines
- Formatting and spacing

The parser processes each line sequentially and stores all information
in INILine objects, making serialization a simple iteration.

Configuration:
- useSpacing: Controls whether new keys are written with spaces around the delimiter
  (e.g., "key = value" vs "key=value"). Default is true. Existing keys preserve
  their original formatting regardless of this setting.
- commentChars: Characters to recognize as comment prefixes. Default is "#;".
  Each character in the slice will be treated as a valid comment prefix.
- delimiter: Key-value delimiter character. Default is '='.
*/

package iniparser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
)

// RelaxedINIParser preserves comments, formatting, and structure
type RelaxedINIParser struct {
	lines        []INILine
	useSpacing   bool // Add spaces around separator for new keys (default: true)
	commentChars []byte
	delimiter    byte
}

// INILine represents a single line in an INI file with all its components
type INILine struct {
	Section       string
	Key           string
	Value         string
	Indent        string
	Delimiter     string
	Suffix        string
	CommentPrefix string
	Original      []byte
	IsEmpty       bool
	IsSection     bool
	_             [6]byte
	_             [48]byte
}

func NewRelaxedINIParser() *RelaxedINIParser {
	return &RelaxedINIParser{
		lines:        make([]INILine, 0),
		useSpacing:   true,
		commentChars: []byte{'#', ';'},
		delimiter:    '=',
	}
}

func (p *RelaxedINIParser) Parse(data []byte) ([]INILine, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	p.lines = make([]INILine, 0)

	currentSection := ""
	for scanner.Scan() {
		parsedLine := p.parseLine(scanner.Bytes(), currentSection)
		if parsedLine.IsSection {
			currentSection = parsedLine.Section
		}
		p.lines = append(p.lines, parsedLine)
	}

	return p.lines, scanner.Err()
}

func (p *RelaxedINIParser) parseLine(lineBytes []byte, section string) INILine {
	iniLine := INILine{
		Original: lineBytes,
		Section:  section,
	}

	if len(lineBytes) == 0 {
		iniLine.IsEmpty = true
		return iniLine
	}

	i := 0
	n := len(lineBytes)

	// Step 1: Extract leading space prefix
	spaceStart := i
	for i < n && (lineBytes[i] == ' ' || lineBytes[i] == '\t') {
		i++
	}
	if i > spaceStart {
		iniLine.Indent = string(lineBytes[spaceStart:i])
	}

	// Check if line is empty (only spaces)
	if i >= n {
		iniLine.IsEmpty = true
		return iniLine
	}

	// Step 2: Check for comment character
	isComment := false
	if slices.Contains(p.commentChars, lineBytes[i]) {
		commentStart := i
		isComment = true
		i++
		// Capture all spaces after comment char as part of CommentPrefix
		for i < n && (lineBytes[i] == ' ' || lineBytes[i] == '\t') {
			i++
		}
		iniLine.CommentPrefix = string(lineBytes[commentStart:i])
	}

	if i >= n {
		return iniLine
	}

	// If this is a comment line, treat the rest as plain text (don't parse sections/keys)
	if isComment {
		iniLine.Key = string(lineBytes[i:])
		return iniLine
	}

	// Step 3: Check for section [...]
	if lineBytes[i] == '[' {
		iniLine.IsSection = true
		sectionStart := i + 1
		i++
		// Find closing ]
		for i < n && lineBytes[i] != ']' {
			i++
		}
		if i < n && lineBytes[i] == ']' {
			iniLine.Section = string(lineBytes[sectionStart:i])
			i++ // skip ]
			// Capture any suffix after ]
			if i < n {
				iniLine.Suffix = string(lineBytes[i:])
			}
		} else {
			// No closing ], treat as malformed section
			iniLine.Section = string(lineBytes[sectionStart:])
		}
		return iniLine
	}

	// Step 4: Extract key (until delimiter or end of line)
	keyStart := i
	for i < n && lineBytes[i] != p.delimiter {
		i++
	}

	// If we found delimiter, separate key from value
	if i < n && lineBytes[i] == p.delimiter {
		// Extract key (may have trailing spaces)
		keyEnd := i

		// Trim trailing spaces from key to find actual key end
		keyActualEnd := keyEnd
		for keyActualEnd > keyStart && (lineBytes[keyActualEnd-1] == ' ' || lineBytes[keyActualEnd-1] == '\t') {
			keyActualEnd--
		}
		iniLine.Key = string(lineBytes[keyStart:keyActualEnd])

		// Capture delimiter with surrounding spaces (spaces before delimiter + delimiter + spaces after)
		delimStart := keyActualEnd // Start from end of actual key
		i++                        // move past delimiter

		// Find where spaces after delimiter end (start of value)
		for i < n && (lineBytes[i] == ' ' || lineBytes[i] == '\t') {
			i++
		}

		// Delimiter captures: spaces-before + delimiter + spaces-after
		iniLine.Delimiter = string(lineBytes[delimStart:i])

		// Step 5: Extract value (rest of line, may have trailing spaces)
		if i < n {
			iniLine.Value = string(lineBytes[i:])
		}
	} else {
		// No delimiter found, entire remaining part is the key (bare key/flag)
		iniLine.Key = string(lineBytes[keyStart:])
	}

	return iniLine
}

func (p *RelaxedINIParser) Serialize(lines []INILine, writer io.Writer) error {
	for _, line := range lines {
		p.writeLine(writer, &line)
	}
	return nil
}

func (p *RelaxedINIParser) writeLine(writer io.Writer, line *INILine) {
	if line.IsEmpty || (line.CommentPrefix != "" && line.Key == "") {
		fmt.Fprintln(writer, string(line.Original))
		return
	}

	var builder strings.Builder
	builder.WriteString(line.Indent)
	builder.WriteString(line.CommentPrefix)

	if line.IsSection {
		builder.WriteByte('[')
		builder.WriteString(line.Section)
		builder.WriteByte(']')
		builder.WriteString(line.Suffix)
		fmt.Fprintln(writer, builder.String())
		return
	}

	if line.Key != "" {
		builder.WriteString(line.Key)
		if line.Delimiter != "" {
			builder.WriteString(line.Delimiter)
			builder.WriteString(line.Value)
		} else if line.Value != "" {
			if p.useSpacing {
				builder.WriteString(" ")
				builder.WriteByte(p.delimiter)
				builder.WriteString(" ")
			} else {
				builder.WriteByte(p.delimiter)
			}
			builder.WriteString(line.Value)
		}
	}

	builder.WriteString(line.Suffix)
	fmt.Fprintln(writer, builder.String())
}

func (p *RelaxedINIParser) UpdateValue(lines []INILine, section, key, newValue string) []INILine {
	currentSection := ""
	for i := range lines {
		line := &lines[i]
		if line.IsSection {
			currentSection = line.Section
		} else if currentSection == section && line.Key == key {
			line.Value = newValue
			break
		}
	}
	return lines
}

func (p *RelaxedINIParser) AddKey(lines []INILine, section, key, value string) []INILine {
	insertIndex := len(lines)

	if section == "" {
		for i, line := range lines {
			if line.IsEmpty || line.IsSection || (line.CommentPrefix != "" && line.Key != "") {
				insertIndex = i
				break
			}
		}
	} else {
		inSection := false
		lastKeyIndex := -1
		for i, line := range lines {
			if line.IsSection {
				if line.Section == section {
					inSection = true
				} else if inSection {
					break
				}
			} else if inSection && line.Key != "" && line.CommentPrefix == "" {
				lastKeyIndex = i
			}
		}
		if lastKeyIndex != -1 {
			insertIndex = lastKeyIndex + 1
		}
	}

	newLine := INILine{Section: section, Key: key, Value: value}
	lines = append(lines, INILine{})
	copy(lines[insertIndex+1:], lines[insertIndex:])
	lines[insertIndex] = newLine
	return lines
}

func GetValue(lines []INILine, section, key string) (string, bool) {
	currentSection := ""
	for _, line := range lines {
		if line.IsSection {
			currentSection = line.Section
		} else if currentSection == section && line.Key == key {
			return line.Value, true
		}
	}
	return "", false
}

func KeyExists(lines []INILine, section, key string) bool {
	currentSection := ""
	for _, line := range lines {
		if line.IsSection {
			currentSection = line.Section
		} else if currentSection == section && line.Key == key {
			return true
		}
	}
	return false
}
