package main

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	fmt.Println("=== Image Processing Pipeline ===")
	fmt.Println("Pipeline Pattern: Generator -> Loader -> Processor -> Saver")

	// Criar diretórios se não existirem
	inputDir := "./assets/image_processor_pipeline/input"
	outputDir := "./assets/image_processor_pipeline/output"

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		fmt.Printf("Erro criando input dir: %v\n", err)
		return
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Erro criando output dir: %v\n", err)
		return
	}

	// Criar pipeline
	pipeline := NewPipeline(inputDir, outputDir)

	// Context com timeout de 30 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Executar pipeline
	fmt.Printf("Processando imagens de %s...\n", inputDir)
	start := time.Now()

	if err := pipeline.Run(ctx); err != nil {
		fmt.Printf("Erro no pipeline: %v\n", err)
		return
	}

	elapsed := time.Since(start)
	fmt.Printf("\n✓ Pipeline concluído em %v\n", elapsed)
	fmt.Printf("✓ Imagens salvas em %s\n", outputDir)
	fmt.Println("\nRun 'go test -v' to verify your implementation")
}

// ImageJob representa uma imagem no pipeline
type ImageJob struct {
	Path     string      // caminho do arquivo original
	Image    image.Image // imagem carregada (nil nos primeiros stages)
	Error    error       // erro se algo deu errado
	StageNum int         // apenas para debug
}

// Pipeline gerencia os 4 stages de processamento
type Pipeline struct {
	inputDir  string
	outputDir string

	wg sync.WaitGroup
}

func NewPipeline(inputDir, outputDir string) *Pipeline {
	return &Pipeline{
		inputDir:  inputDir,
		outputDir: outputDir,
	}
}

func (p *Pipeline) Run(ctx context.Context) error {
	fileChan := make(chan ImageJob, 10)      // Generator -> Loader
	imageChan := make(chan ImageJob, 10)     // Loader -> Processor
	processedChan := make(chan ImageJob, 10) // Processor -> Saver

	var wg sync.WaitGroup

	wg.Go(func() {
		// defer wg.Done()
		go p.generator(context.Background(), fileChan) // primeiro escrevemos no filechan
	})

	wg.Go(func() {
		go p.loader(ctx, fileChan, imageChan)
	})

	wg.Go(func() {
		go p.processor(ctx, imageChan, processedChan)
	})

	wg.Go(func() {
		p.saver(ctx, processedChan)
	})

	wg.Wait()

	return nil
}
func (p *Pipeline) generator(ctx context.Context, outputChan chan<- ImageJob) {

	defer close(outputChan)

	patterns := []string{"*.jpeg", "*.jpg", "*.png"}

	for _, ext := range patterns {
		pattern := filepath.Join(p.inputDir, ext)
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		for _, filePath := range files {
			select {
			case <-ctx.Done():
				return
			case outputChan <- ImageJob{Path: filePath, StageNum: 1}:
			}
		}
	}
}

// Stage 2: Loader - carrega imagens do disco
func (p *Pipeline) loader(ctx context.Context, inputChan <-chan ImageJob, outputChan chan<- ImageJob) {
	defer close(outputChan)

	for job := range inputChan {
		select {
		case <-ctx.Done():
			return
		default:
			f, err := os.Open(job.Path)
			if err != nil {
				job.Error = err
				outputChan <- job
				fmt.Println("err opening file:", err)
				continue
			}

			imageDecoded, _, err := image.Decode(f)
			f.Close()
			if err != nil {
				job.Error = err
			}

			outputChan <- ImageJob{
				Path:     job.Path,
				Image:    imageDecoded,
				StageNum: 2,
			}
		}
	}
}

// Stage 3: Processor - processa as imagens (grayscale)
func (p *Pipeline) processor(ctx context.Context, inputChan <-chan ImageJob, outputChan chan<- ImageJob) {
	defer close(outputChan)

	for job := range inputChan {
		select {
		case <-ctx.Done():
			return
		default:
			if job.Error != nil || job.Image == nil {
				outputChan <- job
				continue
			}

			bounds := job.Image.Bounds()
			grayImg := image.NewGray(bounds)

			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					r, g, b, _ := job.Image.At(x, y).RGBA()
					// converter para grayscale usando luminancia
					gray := uint8((r*299 + g*587 + b*114) / 1000 >> 8)
					grayImg.SetGray(x, y, color.Gray{Y: gray})
				}
			}

			job.Image = grayImg
			job.StageNum = 3
			fmt.Printf("[Processor] Processed: %s\n", filepath.Base(job.Path))
			outputChan <- job
		}
	}
}

// Stage 4: Saver - salva imagens processadas
func (p *Pipeline) saver(ctx context.Context, inputChan <-chan ImageJob) {
	for job := range inputChan {
		select {
		case <-ctx.Done():
			return
		default:
			if job.Error != nil || job.Image == nil {
				fmt.Printf("[Saver] Skipping %s due to error: %v\n", filepath.Base(job.Path), job.Error)
				continue
			}

			baseName := filepath.Base(job.Path)
			ext := filepath.Ext(baseName)
			nameWithoutExt := strings.TrimSuffix(baseName, ext)
			outPath := filepath.Join(p.outputDir, nameWithoutExt+"_processed.jpg")

			outFile, err := os.Create(outPath)
			if err != nil {
				fmt.Printf("[Saver] Error creating %s: %v\n", outPath, err)
				continue
			}

			err = jpeg.Encode(outFile, job.Image, &jpeg.Options{Quality: 90})
			outFile.Close()

			if err != nil {
				fmt.Printf("[Saver] Error encoding %s: %v\n", outPath, err)
			} else {
				fmt.Printf("[Saver] Saved: %s\n", filepath.Base(outPath))
			}
		}
	}
}
