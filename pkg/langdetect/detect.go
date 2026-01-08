// Package langdetect provides programming language detection for code execution.
package langdetect

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-enry/go-enry/v2"
)

// Detector handles programming language detection.
type Detector struct {
	// customMappings for additional file types
	customMappings map[string]string
}

// New creates a new language detector.
func New() *Detector {
	return &Detector{
		customMappings: make(map[string]string),
	}
}

// DetectOptions configures detection behavior.
type DetectOptions struct {
	// Filename hint (e.g., "main.py")
	Filename string

	// UseContent enables content-based detection
	UseContent bool

	// UseShebang enables shebang detection
	UseShebang bool

	// UseHeuristics enables pattern-based heuristics
	UseHeuristics bool
}

// DefaultDetectOptions returns sensible defaults.
func DefaultDetectOptions() *DetectOptions {
	return &DetectOptions{
		UseContent:    true,
		UseShebang:    true,
		UseHeuristics: true,
	}
}

// DetectResult contains detection results.
type DetectResult struct {
	// Language is the detected language name.
	Language string

	// Confidence is the detection confidence (0.0 to 1.0).
	Confidence float64

	// Method indicates how the language was detected.
	Method string
}

// Detect identifies the programming language of code.
func (d *Detector) Detect(code string, opts *DetectOptions) *DetectResult {
	if opts == nil {
		opts = DefaultDetectOptions()
	}

	// Strategy 1: Check filename/extension if provided
	if opts.Filename != "" {
		// Try exact filename match (e.g., Makefile, Dockerfile)
		if lang, safe := enry.GetLanguageByFilename(opts.Filename); safe && lang != "" {
			return &DetectResult{Language: lang, Confidence: 1.0, Method: "filename"}
		}

		// Try extension
		if lang, safe := enry.GetLanguageByExtension(opts.Filename); safe && lang != "" {
			return &DetectResult{Language: lang, Confidence: 0.95, Method: "extension"}
		}
	}

	// Strategy 2: Check shebang
	if opts.UseShebang && strings.HasPrefix(strings.TrimSpace(code), "#!") {
		if lang, safe := enry.GetLanguageByShebang([]byte(code)); safe && lang != "" {
			return &DetectResult{Language: lang, Confidence: 0.95, Method: "shebang"}
		}
	}

	// Strategy 3: Content-based detection using go-enry
	if opts.UseContent {
		filename := opts.Filename
		if filename == "" {
			filename = "code"
		}

		// Get all possible languages
		languages := enry.GetLanguages(filename, []byte(code))
		if len(languages) == 1 {
			return &DetectResult{Language: languages[0], Confidence: 0.9, Method: "content"}
		}
		if len(languages) > 1 {
			// Use classifier to pick best match
			lang := enry.GetLanguage(filename, []byte(code))
			if lang != "" {
				return &DetectResult{Language: lang, Confidence: 0.8, Method: "classifier"}
			}
		}
	}

	// Strategy 4: Heuristic patterns
	if opts.UseHeuristics {
		if result := d.detectByPatterns(code); result != nil {
			return result
		}
	}

	return &DetectResult{Language: "", Confidence: 0, Method: "unknown"}
}

// detectByPatterns uses regex patterns for common language constructs.
func (d *Detector) detectByPatterns(code string) *DetectResult {
	patterns := map[string][]string{
		"Python": {
			`(?m)^import\s+\w+`,
			`(?m)^from\s+\w+\s+import`,
			`(?m)^def\s+\w+\s*\(`,
			`(?m)^class\s+\w+.*:`,
			`(?m)^\s*print\s*\(`,
			`exec\s*\(`,
			`range\s*\(`,
			`len\s*\(`,
			`str\s*\(`,
			`int\s*\(`,
			`list\s*\(`,
			`dict\s*\(`,
			`\.append\s*\(`,
			`\.join\s*\(`,
			`time\.sleep`,
			`for\s+\w+\s+in\s+`,
			`if\s+__name__\s*==`,
			`lambda\s+\w*:`,
			`\[\s*\w+\s+for\s+\w+\s+in`,
		},
		"Go": {
			`(?m)^package\s+\w+`,
			`\bpackage\s+main\b`,
			`(?m)^import\s*\(`,
			`(?m)^func\s+\w*\s*\(`,
			`\bfunc\s+main\s*\(`,
			`(?m)^type\s+\w+\s+(struct|interface)`,
			`:=`,
			`fmt\.Print`,
			`fmt\.Sprintf`,
			`fmt\.Errorf`,
			`errors\.New`,
			`make\s*\(\s*(map|chan|\[\])`,
			`go\s+func\s*\(`,
			`<-\s*\w+`,
			`defer\s+`,
			`panic\s*\(`,
			`recover\s*\(`,
			`range\s+\w+`,
			`\[\]byte`,
			`\[\]string`,
			`map\[string\]`,
			`interface\{\}`,
			`struct\s*\{`,
		},
		"JavaScript": {
			`(?m)^const\s+\w+\s*=`,
			`(?m)^let\s+\w+\s*=`,
			`(?m)^var\s+\w+\s*=`,
			`(?m)^function\s+\w+\s*\(`,
			`=>\s*[{\(]`,
			`console\.log\s*\(`,
			`console\.error\s*\(`,
			`console\.warn\s*\(`,
			`require\s*\(`,
			`module\.exports`,
			`exports\.`,
			`document\.`,
			`window\.`,
			`async\s+function`,
			`await\s+`,
			`\.then\s*\(`,
			`\.catch\s*\(`,
			`JSON\.parse`,
			`JSON\.stringify`,
			`Array\.`,
			`Object\.`,
			`Promise\.`,
			`new\s+Promise`,
			`setTimeout\s*\(`,
			`setInterval\s*\(`,
		},
		"TypeScript": {
			`(?m)^interface\s+\w+`,
			`(?m)^type\s+\w+\s*=`,
			`:\s*(string|number|boolean|any)\b`,
			`<[A-Z]\w*>`,
		},
		"Rust": {
			`(?m)^fn\s+\w+`,
			`\bfn\s+main\s*\(`,
			`(?m)^use\s+\w+`,
			`(?m)^mod\s+\w+`,
			`(?m)^struct\s+\w+`,
			`(?m)^impl\s+`,
			`(?m)^let\s+mut\s+`,
			`(?m)^pub\s+(fn|struct|enum|mod)`,
			`println!\s*\(`,
			`print!\s*\(`,
			`eprintln!\s*\(`,
			`format!\s*\(`,
			`vec!\s*\[`,
			`panic!\s*\(`,
			`->\s*(i32|i64|u32|u64|f32|f64|bool|String|&str|\(\))`,
			`&mut\s+\w+`,
			`&str`,
			`::new\s*\(`,
			`\.unwrap\s*\(`,
			`\.expect\s*\(`,
			`Option<`,
			`Result<`,
			`Some\s*\(`,
			`None\b`,
			`Ok\s*\(`,
			`Err\s*\(`,
		},
		"Java": {
			`(?m)^public\s+class\s+\w+`,
			`(?m)^import\s+java\.`,
			`(?m)^package\s+\w+(\.\w+)*;`,
			`System\.out\.print`,
			`public\s+static\s+void\s+main`,
		},
		"Ruby": {
			`(?m)^require\s+['"]`,
			`(?m)^def\s+\w+`,
			`(?m)^class\s+\w+`,
			`(?m)^module\s+\w+`,
			`\.each\s+do\s*\|`,
			`puts\s+`,
		},
		"PHP": {
			`(?m)^<\?php`,
			`\$\w+\s*=`,
			`(?m)^function\s+\w+\s*\(`,
			`echo\s+`,
			`->\w+\(`,
		},
		"C": {
			`(?m)^#include\s*<`,
			`(?m)^int\s+main\s*\(`,
			`printf\s*\(`,
			`(?m)^(void|int|char|float|double)\s+\w+\s*\(`,
		},
		"C++": {
			`(?m)^#include\s*<iostream>`,
			`std::`,
			`cout\s*<<`,
			`(?m)^class\s+\w+\s*[:{]`,
			`(?m)^namespace\s+\w+`,
		},
		"C#": {
			`(?m)^using\s+System`,
			`(?m)^namespace\s+\w+`,
			`(?m)^class\s+\w+`,
			`Console\.(Write|Read)`,
			`(?m)^public\s+(class|interface|enum)`,
		},
		"Shell": {
			`(?m)^#!/bin/(ba)?sh`,
			`(?m)^\s*if\s+\[\s+`,
			`(?m)^\s*for\s+\w+\s+in\s+`,
			`\$\{?\w+\}?`,
			`(?m)^\s*echo\s+`,
		},
		"SQL": {
			`(?mi)^SELECT\s+`,
			`(?mi)^INSERT\s+INTO`,
			`(?mi)^UPDATE\s+\w+\s+SET`,
			`(?mi)^CREATE\s+TABLE`,
			`(?mi)^DROP\s+TABLE`,
		},
		"R": {
			`(?m)^library\s*\(`,
			`<-\s*`,
			`(?m)^function\s*\(`,
			`data\.frame\s*\(`,
			`ggplot\s*\(`,
		},
		"Kotlin": {
			`(?m)^fun\s+\w+`,
			`(?m)^val\s+\w+`,
			`(?m)^var\s+\w+`,
			`(?m)^class\s+\w+`,
			`println\s*\(`,
		},
		"Swift": {
			`(?m)^import\s+(Foundation|UIKit|SwiftUI|Cocoa|Darwin)`,
			`(?m)^func\s+\w+\s*\([^)]*\)\s*(->|{)`,
			`(?m)^let\s+\w+\s*:\s*\w+`,
			`(?m)^var\s+\w+\s*:\s*\w+`,
			`(?m)^class\s+\w+\s*:\s*\w+`,
			`(?m)^struct\s+\w+`,
			`(?m)^enum\s+\w+`,
			`(?m)^protocol\s+\w+`,
			`guard\s+let`,
			`if\s+let`,
			`@IBOutlet`,
			`@IBAction`,
			`override\s+func`,
		},
		"Scala": {
			`(?m)^object\s+\w+`,
			`(?m)^def\s+\w+`,
			`(?m)^val\s+\w+`,
			`(?m)^var\s+\w+`,
			`println\s*\(`,
		},
	}

	scores := make(map[string]int)

	for lang, regexes := range patterns {
		for _, pattern := range regexes {
			if matched, _ := regexp.MatchString(pattern, code); matched {
				scores[lang]++
			}
		}
	}

	// Find language with highest score
	var bestLang string
	var bestScore int
	for lang, score := range scores {
		if score > bestScore {
			bestLang = lang
			bestScore = score
		}
	}

	if bestScore >= 1 {
		confidence := float64(bestScore) / 5.0
		if confidence > 0.8 {
			confidence = 0.8
		}
		if confidence < 0.2 {
			confidence = 0.2
		}
		return &DetectResult{Language: bestLang, Confidence: confidence, Method: "heuristic"}
	}

	return nil
}

// AddMapping adds a custom file extension to language mapping.
func (d *Detector) AddMapping(extension, language string) {
	d.customMappings[extension] = language
}

// DetectFromFilename detects language from filename only.
func (d *Detector) DetectFromFilename(filename string) *DetectResult {
	ext := filepath.Ext(filename)

	// Check custom mappings first
	if lang, ok := d.customMappings[ext]; ok {
		return &DetectResult{Language: lang, Confidence: 1.0, Method: "custom"}
	}

	// Use enry
	if lang, safe := enry.GetLanguageByFilename(filename); safe && lang != "" {
		return &DetectResult{Language: lang, Confidence: 1.0, Method: "filename"}
	}

	if lang, safe := enry.GetLanguageByExtension(filename); safe {
		return &DetectResult{Language: lang, Confidence: 0.95, Method: "extension"}
	}

	return &DetectResult{Language: "", Confidence: 0, Method: "unknown"}
}

// Quick performs fast detection with less accuracy.
func Quick(code string) string {
	d := New()
	result := d.Detect(code, &DetectOptions{
		UseContent:    true,
		UseShebang:    true,
		UseHeuristics: false,
	})
	return result.Language
}

// Full performs comprehensive detection.
func Full(code string, filename string) *DetectResult {
	d := New()
	return d.Detect(code, &DetectOptions{
		Filename:      filename,
		UseContent:    true,
		UseShebang:    true,
		UseHeuristics: true,
	})
}
