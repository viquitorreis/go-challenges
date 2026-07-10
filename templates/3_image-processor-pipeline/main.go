package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"time"
)

func main() {
	fmt.Println("=== Image Processing Pipeline ===")
	fmt.Println("Pipeline Pattern: Generator -> Loader -> Processor -> Saver")

	// Criar diretórios se não existirem
	inputDir := "./input_images"
	outputDir := "./output_images"

	if err := os.MkdirAll(inputDir, 0755); err != nil {
		fmt.Printf("Erro criando input dir: %v\n", err)
		return
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Erro criando output dir: %v\n", err)
		return
	}

	// Criar algumas imagens de teste se não existirem
	createTestImages(inputDir)

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
	fmt.Println("Run 'go test -v' to verify your implementation")
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
	// TODO: adicione campos necessários para:
	// - Channels entre stages
	// - Context para cancellation
	// - WaitGroup para coordenação
	// - Caminho dos diretórios input/output
}

func NewPipeline(inputDir, outputDir string) *Pipeline {
	return &Pipeline{
		// TODO: inicialize campos
	}
}

// Run executa o pipeline completo
func (p *Pipeline) Run(ctx context.Context) error {
	// TODO: conecte os 4 stages usando channels
	// Generator -> fileChan -> Loader -> imageChan -> Processor -> processedChan -> Saver

	// Dica: cada stage deve ser uma goroutine
	// Dica: use WaitGroup para saber quando tudo terminou
	// Dica: passe o context para poder cancelar em caso de erro

	return nil
}

// Stage 1: Generator - lista arquivos de imagens no diretório
func (p *Pipeline) generator(ctx context.Context, outputChan chan<- ImageJob) {
	// TODO:
	// 1. Listar arquivos .jpg e .png no inputDir usando filepath.Glob ou filepath.Walk
	// 2. Para cada arquivo, criar um ImageJob com o Path
	// 3. Enviar para o outputChan
	// 4. Fechar o outputChan quando terminar (importante!)
	// 5. Respeitar o context - se ctx.Done(), parar imediatamente

	defer close(outputChan)

	// Hint: filepath.Glob("dir/*.jpg") ou filepath.Walk
	// Hint: select { case <-ctx.Done(): return; case outputChan <- job: }
}

// Stage 2: Loader - carrega imagens do disco
func (p *Pipeline) loader(ctx context.Context, inputChan <-chan ImageJob, outputChan chan<- ImageJob) {
	// TODO:
	// 1. Receber jobs do inputChan
	// 2. Para cada job, abrir o arquivo (os.Open)
	// 3. Decodificar a imagem (image.Decode ou jpeg.Decode/png.Decode)
	// 4. Colocar a imagem no job.Image
	// 5. Se der erro, colocar em job.Error
	// 6. Enviar job para outputChan
	// 7. Fechar outputChan quando inputChan fechar

	defer close(outputChan)

	// Hint: import "image/jpeg" e "image/png"
	// Hint: image.Decode detecta o formato automaticamente
	// Hint: não esqueça de fechar o arquivo: defer file.Close()
}

// Stage 3: Processor - processa as imagens (grayscale)
func (p *Pipeline) processor(ctx context.Context, inputChan <-chan ImageJob, outputChan chan<- ImageJob) {
	// TODO:
	// 1. Receber jobs com imagens carregadas
	// 2. Converter para grayscale (percorrer pixels e calcular média RGB)
	// 3. Ou redimensionar (usar bounds e criar nova imagem menor)
	// 4. Atualizar job.Image com a imagem processada
	// 5. Enviar para outputChan

	defer close(outputChan)

	// Hint: bounds := img.Bounds()
	// Hint: new image: image.NewGray(bounds) ou image.NewRGBA(bounds)
	// Hint: loop: for y := bounds.Min.Y; y < bounds.Max.Y; y++ { for x := ... }
}

// Stage 4: Saver - salva imagens processadas
func (p *Pipeline) saver(ctx context.Context, inputChan <-chan ImageJob) {
	// TODO:
	// 1. Receber jobs com imagens processadas
	// 2. Criar arquivo no outputDir (mesmo nome, adicionar sufixo "_processed")
	// 3. Encode da imagem no formato JPEG
	// 4. Fechar arquivo
	// 5. Imprimir sucesso ou erro

	// Hint: novoNome := strings.TrimSuffix(basename, ext) + "_processed.jpg"
	// Hint: jpeg.Encode(file, img, &jpeg.Options{Quality: 90})
	// Não precisa fechar channel aqui - é o último stage
}
