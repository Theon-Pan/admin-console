//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/mojocn/base64Captcha"
)

// outputDir is the directory where captcha images are saved,
// relative to the project root (run from the repo root).
const outputDir = "examples/captcha/imgs"

func main() {
	// Ensure output directory exists.
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory %q: %v", outputDir, err)
	}

	// Configure the math driver.
	driver := &base64Captcha.DriverMath{
		Height:          50,
		Width:           200,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowSlimeLine | base64Captcha.OptionShowSineLine,
	}

	// Generate a math question (e.g., "3+7=?").
	id, question, answer := driver.GenerateIdQuestionAnswer()

	// Draw the captcha image.
	item, err := driver.DrawCaptcha(question)
	if err != nil {
		log.Fatalf("Failed to draw captcha: %v", err)
	}

	// Save as PNG under examples/captcha/imgs/math_<id>.png
	filename := fmt.Sprintf("math_%s.png", id)
	filePath := filepath.Join(outputDir, filename)

	f, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("Failed to create file %q: %v", filePath, err)
	}
	defer f.Close()

	if _, err := item.WriteTo(f); err != nil {
		log.Fatalf("Failed to write captcha image: %v", err)
	}

	fmt.Printf("Question : %s\n", question)
	fmt.Printf("Answer   : %s\n", answer)
	fmt.Printf("Saved to : %s\n", filePath)
}
