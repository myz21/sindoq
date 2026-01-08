package langdetect

import (
	"testing"
)

func TestDetector_DetectPython(t *testing.T) {
	d := New()
	// Use DefaultDetectOptions which enables heuristics for pattern detection
	opts := DefaultDetectOptions()

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "def function with print",
			code: `def hello():
    print("Hello, World!")
    return "Hello"`,
			expected: "Python",
		},
		{
			name: "import and class",
			code: `import os
import sys

class MyClass:
    def __init__(self):
        pass`,
			expected: "Python",
		},
		{
			name: "full script",
			code: `#!/usr/bin/env python3
import json

def main():
    data = {"key": "value"}
    print(json.dumps(data))

if __name__ == "__main__":
    main()`,
			expected: "Python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.Detect(tt.code, opts)
			if result.Language != tt.expected {
				t.Errorf("Detect() = %v, want %v (method: %s, confidence: %f)", result.Language, tt.expected, result.Method, result.Confidence)
			}
		})
	}
}

func TestDetector_DetectJavaScript(t *testing.T) {
	d := New()
	opts := DefaultDetectOptions()

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "console and function",
			code: `function hello() {
    console.log("Hello, World!");
    return "Hello";
}`,
			expected: "JavaScript",
		},
		{
			name: "const and require",
			code: `const fs = require('fs');
const path = require('path');

const data = fs.readFileSync('file.txt');
console.log(data);`,
			expected: "JavaScript",
		},
		{
			name: "arrow functions",
			code: `const hello = () => {
    const message = "Hello";
    console.log(message);
    return message;
};`,
			expected: "JavaScript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.Detect(tt.code, opts)
			if result.Language != tt.expected {
				t.Errorf("Detect() = %v, want %v (method: %s, confidence: %f)", result.Language, tt.expected, result.Method, result.Confidence)
			}
		})
	}
}

func TestDetector_DetectGo(t *testing.T) {
	d := New()
	opts := DefaultDetectOptions()

	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "package main",
			code: `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}`,
			expected: "Go",
		},
		{
			name: "struct and func",
			code: `package mypackage

type User struct {
	Name string
	Age  int
}

func (u *User) Greet() string {
	return "Hello, " + u.Name
}`,
			expected: "Go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.Detect(tt.code, opts)
			if result.Language != tt.expected {
				t.Errorf("Detect() = %v, want %v (method: %s, confidence: %f)", result.Language, tt.expected, result.Method, result.Confidence)
			}
		})
	}
}

func TestDetector_DetectWithOptions(t *testing.T) {
	d := New()
	tests := []struct {
		name     string
		code     string
		opts     *DetectOptions
		expected string
	}{
		{
			name:     "detect by filename",
			code:     "some code",
			opts:     &DetectOptions{Filename: "test.py"},
			expected: "Python",
		},
		{
			name:     "detect js by filename",
			code:     "some code",
			opts:     &DetectOptions{Filename: "index.js"},
			expected: "JavaScript",
		},
		{
			name:     "detect by shebang",
			code:     "#!/usr/bin/env python3\nprint('hello')",
			opts:     &DetectOptions{UseShebang: true},
			expected: "Python",
		},
		{
			name:     "detect bash by shebang",
			code:     "#!/bin/bash\necho hello",
			opts:     &DetectOptions{UseShebang: true},
			expected: "Shell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.Detect(tt.code, tt.opts)
			if result.Language != tt.expected {
				t.Errorf("Detect() = %v, want %v (method: %s)", result.Language, tt.expected, result.Method)
			}
		})
	}
}

func TestDetector_DetectFromFilename(t *testing.T) {
	d := New()
	// Note: go-enry may not recognize all extensions equally.
	// These are the ones that reliably work.
	tests := []struct {
		filename string
		expected string
	}{
		{"main.py", "Python"},
		{"app.js", "JavaScript"},
		{"main.go", "Go"},
		{"App.java", "Java"},
		{"script.rb", "Ruby"},
		{"main.c", "C"},
		{"main.cpp", "C++"},
		{"script.sh", "Shell"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := d.DetectFromFilename(tt.filename)
			if result.Language != tt.expected {
				t.Errorf("DetectFromFilename(%q) = %v, want %v", tt.filename, result.Language, tt.expected)
			}
		})
	}
}

func TestQuick(t *testing.T) {
	// Quick uses content detection without heuristics, so needs longer/clearer code
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{
			name: "python with shebang",
			code: `#!/usr/bin/env python3
print("Hello")`,
			expected: "Python",
		},
		{
			name: "bash with shebang",
			code: `#!/bin/bash
echo "Hello"`,
			expected: "Shell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Quick(tt.code)
			if result != tt.expected {
				t.Errorf("Quick() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFull(t *testing.T) {
	result := Full(`print("Hello")`, "test.py")
	if result.Language != "Python" {
		t.Errorf("Full() = %v, want Python", result.Language)
	}
	if result.Confidence < 0.9 {
		t.Errorf("Full() confidence = %v, want >= 0.9", result.Confidence)
	}
}

func TestSupportedLanguages(t *testing.T) {
	langs := SupportedLanguages()
	if len(langs) == 0 {
		t.Error("SupportedLanguages() returned empty slice")
	}

	// Check for common languages
	expected := []string{"Python", "JavaScript", "Go", "Java", "Rust"}
	for _, lang := range expected {
		found := false
		for _, l := range langs {
			if l == lang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("SupportedLanguages() missing %s", lang)
		}
	}
}

func TestDefaultDetectOptions(t *testing.T) {
	opts := DefaultDetectOptions()
	if opts == nil {
		t.Fatal("DefaultDetectOptions() returned nil")
	}
	if !opts.UseContent {
		t.Error("DefaultDetectOptions().UseContent should be true")
	}
	if !opts.UseShebang {
		t.Error("DefaultDetectOptions().UseShebang should be true")
	}
	if !opts.UseHeuristics {
		t.Error("DefaultDetectOptions().UseHeuristics should be true")
	}
}
