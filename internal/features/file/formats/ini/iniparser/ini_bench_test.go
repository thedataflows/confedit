package iniparser

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// generateLargeINIContent creates a large INI file for benchmarking
func generateLargeINIContent(sections, keysPerSection int) string {
	var builder strings.Builder
	builder.Grow(sections * keysPerSection * 50) // Pre-allocate roughly

	_, _ = builder.WriteString("# Large INI file for benchmarking\n")
	_, _ = builder.WriteString("# Generated automatically\n\n")

	for s := range sections {
		_, _ = builder.WriteString(fmt.Sprintf("[section_%d]\n", s))
		_, _ = builder.WriteString(fmt.Sprintf("# Section %d comment\n", s))

		for k := range keysPerSection {
			_, _ = builder.WriteString(fmt.Sprintf("key_%d_%d = value_%d_%d # inline comment\n", s, k, s, k))
		}
		_, _ = builder.WriteString("\n")
	}

	return builder.String()
}

// Benchmark parsing performance
func BenchmarkINIParser_Parse_Small(b *testing.B) {
	content := generateLargeINIContent(5, 10) // 5 sections, 10 keys each
	data := []byte(content)
	parser := NewRelaxedINIParser()

	b.ReportAllocs()

	for b.Loop() {
		_, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkINIParser_Parse_Medium(b *testing.B) {
	content := generateLargeINIContent(20, 25) // 20 sections, 25 keys each
	data := []byte(content)
	parser := NewRelaxedINIParser()

	b.ReportAllocs()

	for b.Loop() {
		_, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkINIParser_Parse_Large(b *testing.B) {
	content := generateLargeINIContent(100, 50) // 100 sections, 50 keys each
	data := []byte(content)
	parser := NewRelaxedINIParser()

	b.ReportAllocs()

	for b.Loop() {
		_, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark serialization performance
func BenchmarkINIParser_Serialize_Small(b *testing.B) {
	content := generateLargeINIContent(5, 10)
	data := []byte(content)
	parser := NewRelaxedINIParser()

	parsed, err := parser.Parse(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		var buf bytes.Buffer
		err := parser.Serialize(parsed, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkINIParser_Serialize_Medium(b *testing.B) {
	content := generateLargeINIContent(20, 25)
	data := []byte(content)
	parser := NewRelaxedINIParser()

	parsed, err := parser.Parse(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		var buf bytes.Buffer
		err := parser.Serialize(parsed, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkINIParser_Serialize_Large(b *testing.B) {
	content := generateLargeINIContent(100, 50)
	data := []byte(content)
	parser := NewRelaxedINIParser()

	parsed, err := parser.Parse(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		var buf bytes.Buffer
		err := parser.Serialize(parsed, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark complete parse-modify-serialize cycle
func BenchmarkINIParser_ParseModifySerialize_Small(b *testing.B) {
	content := generateLargeINIContent(5, 10)
	data := []byte(content)

	b.ReportAllocs()

	for b.Loop() {
		parser := NewRelaxedINIParser()

		// Parse
		parsed, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}

		// Modify some values
		modifiedKeys := 0
		for i := range parsed {
			line := &parsed[i]
			if line.Key != "" && line.Value != "" && line.CommentPrefix == "" {
				line.Value = "modified_" + line.Value
				modifiedKeys++
				if modifiedKeys >= 1 { // Only modify one key for quick benchmark
					break
				}
			}
		}

		// Serialize
		var buf bytes.Buffer
		err = parser.Serialize(parsed, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkINIParser_ParseModifySerialize_Medium(b *testing.B) {
	content := generateLargeINIContent(20, 25)
	data := []byte(content)

	b.ReportAllocs()

	for b.Loop() {
		parser := NewRelaxedINIParser()

		// Parse
		parsed, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}

		// Modify some values
		modifiedKeys := 0
		for i := range parsed {
			line := &parsed[i]
			if line.Key != "" && line.Value != "" && line.CommentPrefix == "" {
				line.Value = "modified_" + line.Value
				modifiedKeys++
				if modifiedKeys >= 5 { // Modify up to 5 keys for comprehensive benchmark
					break
				}
			}
		}

		// Serialize
		var buf bytes.Buffer
		err = parser.Serialize(parsed, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark memory usage with repeated parsing of the same content
func BenchmarkINIParser_MemoryReuse(b *testing.B) {
	content := generateLargeINIContent(50, 20)
	data := []byte(content)
	parser := NewRelaxedINIParser()

	b.ReportAllocs()

	for b.Loop() {
		// This should reuse the internal line slice capacity
		_, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark RelaxedINIParser directly for comparison
func BenchmarkRelaxedINIParser_Parse_Medium(b *testing.B) {
	content := generateLargeINIContent(20, 25)
	data := []byte(content)
	parser := NewRelaxedINIParser()

	b.ReportAllocs()

	for b.Loop() {
		_, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRelaxedINIParser_Serialize_Medium(b *testing.B) {
	content := generateLargeINIContent(20, 25)
	data := []byte(content)
	parser := NewRelaxedINIParser()

	parsed, err := parser.Parse(data)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	for b.Loop() {
		var buf bytes.Buffer
		err := parser.Serialize(parsed, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark against real-world files
func BenchmarkINIParser_RealWorld_SystemdService(b *testing.B) {
	content := loadTestFileForBench(b, "systemd_service_file.ini")
	data := []byte(content)
	parser := NewRelaxedINIParser()

	b.ReportAllocs()

	for b.Loop() {
		parsed, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}

		var buf bytes.Buffer
		err = parser.Serialize(parsed, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkINIParser_RealWorld_PacmanConf(b *testing.B) {
	content := loadTestFileForBench(b, "pacman.conf")
	data := []byte(content)
	parser := NewRelaxedINIParser()

	b.ReportAllocs()

	for b.Loop() {
		parsed, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}

		var buf bytes.Buffer
		err = parser.Serialize(parsed, &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function for benchmark tests
func loadTestFileForBench(b *testing.B, filename string) string {
	b.Helper()
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("load test file %s: %v", filename, err)
	}
	return string(data)
}
