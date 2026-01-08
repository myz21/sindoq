// Package main demonstrates language detection with sindoq.
package main

import (
	"fmt"

	"github.com/happyhackingspace/sindoq/pkg/langdetect"
)

func main() {
	fmt.Println("=== Language Detection Demo ===")
	fmt.Println()

	// Various code samples
	samples := []struct {
		name string
		code string
	}{
		{
			name: "Python",
			code: `import json

def process_data(items):
    return [item.upper() for item in items]

data = process_data(["hello", "world"])
print(json.dumps(data))`,
		},
		{
			name: "JavaScript",
			code: `const fs = require('fs');

function processData(items) {
    return items.map(item => item.toUpperCase());
}

console.log(processData(['hello', 'world']));`,
		},
		{
			name: "Go",
			code: `package main

import "fmt"

func main() {
    items := []string{"hello", "world"}
    for _, item := range items {
        fmt.Println(item)
    }
}`,
		},
		{
			name: "Rust",
			code: `fn main() {
    let items = vec!["hello", "world"];
    for item in items {
        println!("{}", item.to_uppercase());
    }
}`,
		},
		{
			name: "Java",
			code: `public class Main {
    public static void main(String[] args) {
        String[] items = {"hello", "world"};
        for (String item : items) {
            System.out.println(item.toUpperCase());
        }
    }
}`,
		},
		{
			name: "Shell Script",
			code: `#!/bin/bash
items=("hello" "world")
for item in "${items[@]}"; do
    echo "${item^^}"
done`,
		},
	}

	detector := langdetect.New()
	opts := langdetect.DefaultDetectOptions()

	for _, sample := range samples {
		result := detector.Detect(sample.code, opts)
		fmt.Printf("Sample: %s\n", sample.name)
		fmt.Printf("  Detected: %s\n", result.Language)
		fmt.Printf("  Confidence: %.2f\n", result.Confidence)
		fmt.Printf("  Method: %s\n", result.Method)

		// Get runtime info
		if info, ok := langdetect.GetRuntimeInfo(result.Language); ok {
			fmt.Printf("  Runtime: %s\n", info.Runtime)
			fmt.Printf("  Docker Image: %s\n", info.DockerImage)
		}
		fmt.Println()
	}

	// Demonstrate filename detection
	fmt.Println("=== Filename Detection ===")
	fmt.Println()
	filenames := []string{
		"main.py",
		"app.js",
		"server.go",
		"Main.java",
		"lib.rs",
		"script.sh",
		"index.ts",
	}

	for _, filename := range filenames {
		result := detector.DetectFromFilename(filename)
		fmt.Printf("%s -> %s (confidence: %.2f)\n",
			filename, result.Language, result.Confidence)
	}

	// List all supported languages
	fmt.Println()
	fmt.Println("=== Supported Languages ===")
	fmt.Println()
	for _, lang := range langdetect.SupportedLanguages() {
		info, _ := langdetect.GetRuntimeInfo(lang)
		fmt.Printf("  %s (%s): %s\n", lang, info.FileExt, info.DockerImage)
	}
}
