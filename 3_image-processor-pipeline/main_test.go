package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPipeline_BasicFlow(t *testing.T) {
	inputDir := "./assets/image_processor_pipeline/input"
	outputDir := "./test_output"
	defer os.RemoveAll(outputDir)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	pipeline := NewPipeline(inputDir, outputDir)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pipeline.Run(ctx); err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "*_processed.jpg"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Error("expected processed images, got 0")
	}

	t.Logf("Processed %d images successfully", len(files))
}

func TestPipeline_EmptyDirectory(t *testing.T) {
	inputDir := "./test_empty"
	outputDir := "./test_output_empty"
	defer cleanup(inputDir, outputDir)

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	pipeline := NewPipeline(inputDir, outputDir)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Não deve dar erro com diretório vazio
	if err := pipeline.Run(ctx); err != nil {
		t.Fatalf("Pipeline failed on empty dir: %v", err)
	}

	// Não deve criar nenhum arquivo
	files, _ := filepath.Glob(filepath.Join(outputDir, "*"))
	if len(files) != 0 {
		t.Errorf("expected 0 files in output, got %d", len(files))
	}
}

func TestPipeline_ContextCancellation(t *testing.T) {
	inputDir := "./test_cancel"
	outputDir := "./test_output_cancel"
	defer cleanup(inputDir, outputDir)

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Criar várias imagens
	for i := 0; i < 10; i++ {
		path := filepath.Join(inputDir, fmt.Sprintf("test%d.jpg", i))
		createTestImage(t, path, color.RGBA{uint8(i * 25), 0, 0, 255})
	}

	pipeline := NewPipeline(inputDir, outputDir)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Pipeline deve parar quando context cancelar
	err := pipeline.Run(ctx)

	// Pode dar erro de context canceled ou não (depende do timing)
	// O importante é que não trave para sempre
	if err != nil && err != context.DeadlineExceeded {
		t.Logf("Pipeline stopped with: %v", err)
	}
}

func TestPipeline_MultipleImages(t *testing.T) {
	inputDir := "./test_multiple"
	outputDir := "./test_output_multiple"
	defer cleanup(inputDir, outputDir)

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Criar 5 imagens
	numImages := 5
	for i := 0; i < numImages; i++ {
		path := filepath.Join(inputDir, fmt.Sprintf("img%d.jpg", i))
		createTestImage(t, path, color.RGBA{uint8(i * 50), uint8(i * 30), uint8(i * 20), 255})
	}

	pipeline := NewPipeline(inputDir, outputDir)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	if err := pipeline.Run(ctx); err != nil {
		t.Fatalf("Pipeline failed: %v", err)
	}
	elapsed := time.Since(start)

	// Verificar todas as imagens foram processadas
	files, err := filepath.Glob(filepath.Join(outputDir, "*_processed.jpg"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != numImages {
		t.Errorf("expected %d processed images, got %d", numImages, len(files))
	}

	t.Logf("Processed %d images in %v (avg: %v per image)",
		numImages, elapsed, elapsed/time.Duration(numImages))
}

// Helper functions

func createTestImage(t *testing.T, path string, c color.RGBA) {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, c)
		}
	}

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	if err := jpeg.Encode(file, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
}

func loadImage(t *testing.T, path string) image.Image {
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatal(err)
	}

	return img
}

func isGrayscale(img image.Image) bool {
	bounds := img.Bounds()

	// Verificar alguns pixels aleatórios
	checkPoints := []image.Point{
		{bounds.Min.X, bounds.Min.Y},
		{bounds.Max.X - 1, bounds.Max.Y - 1},
		{(bounds.Min.X + bounds.Max.X) / 2, (bounds.Min.Y + bounds.Max.Y) / 2},
	}

	for _, p := range checkPoints {
		r, g, b, _ := img.At(p.X, p.Y).RGBA()
		// Em grayscale, R == G == B
		if r != g || g != b {
			return false
		}
	}

	return true
}

func cleanup(dirs ...string) {
	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}
