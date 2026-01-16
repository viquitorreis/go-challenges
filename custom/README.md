# Challenges Customizados

## Event Bus

### Aprendizados desse challenge:

✅ Fan-out pattern - um evento distribuído para múltiplos subscribers
✅ Closure bug em goroutines - passar parâmetros vs capturar variáveis do loop
✅ Thread-safety - usar sync.RWMutex com maps compartilhados
✅ Buffered vs unbuffered channels - trade-offs de performance
✅ Ordem vs concorrência - quando goroutines ajudam e quando atrapalham
✅ Trade-offs de sistemas distribuídos - consistency vs availability

Implemente um sistema simples de Event Bus onde:

- Publishers enviam eventos (strings como "user.login", "order.created")
- Subscribers se registram para receber todos os eventos
- Quando um evento é publicado, todos os subscribers recebem concorrentemente
- Cada subscriber processa eventos de forma independente

### Requisitos Funcionais

1. EventBus deve permitir

- Subscribe(name string) <-chan string - registra um subscriber, retorna channel read-only
- Publish(event string) - envia evento para todos subscribers
- Close() - fecha o bus e todos os channels de subscribers

2. Comportamento esperado

- Publicar evento quando não há subscribers -> evento é perdido (ok)
- Subscribers recebem eventos na ordem publicada
- Se um subscriber está lento, não bloqueia outros subscribers
- Fechar o bus -> fecha todos os channels de subscribers

**Constraints**

- Não usar libs externas (só stdlib)
- Implemente uma fila para cada subscriber
- Tempo: ~1 hora
- Código: 80-120 linhas
- Usar sync package conforme necessário

### Func main() para testes:

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

// Exemplo de uso
func main() {
    bus := NewEventBus()
    
    // Subscriber 1: Logger
    logger := bus.Subscribe("logger")
    go func() {
        for event := range logger {
            fmt.Printf("[LOGGER] Received: %s\n", event)
        }
    }()
    
    // Subscriber 2: Counter
    counter := bus.Subscribe("counter")
    eventCount := 0
    go func() {
        for event := range counter {
            eventCount++
            fmt.Printf("[COUNTER] Total events: %d\n", eventCount)
        }
    }()
    
    // Publisher envia eventos
    time.Sleep(100 * time.Millisecond) // dar tempo pros subscribers iniciarem
    
    bus.Publish("user.login")
    bus.Publish("order.created")
    bus.Publish("user.logout")
    
    time.Sleep(500 * time.Millisecond) // dar tempo para processar
    
    bus.Close()
    time.Sleep(100 * time.Millisecond) // dar tempo para fechar
    
    fmt.Println("Event bus closed")
}

// EventBus gerencia publishers e subscribers
type EventBus struct {
    // TODO: adicione campos necessários
    // Dica: você precisa guardar os channels dos subscribers
    // Dica: você precisa de um mutex para thread-safety
}

// NewEventBus cria um novo event bus
func NewEventBus() *EventBus {
    return &EventBus{
        // TODO: inicialize campos
    }
}

// Subscribe registra um novo subscriber e retorna um channel para receber eventos
func (eb *EventBus) Subscribe(name string) <-chan string {
    // TODO: 
    // 1. Criar um channel para este subscriber
    // 2. Guardar o channel internamente
    // 3. Retornar o channel (read-only)
}

// Publish envia um evento para todos os subscribers
func (eb *EventBus) Publish(event string) {
    // TODO:
    // 1. Iterar por todos os subscribers
    // 2. Enviar o evento para cada channel
    // 3. Usar goroutines para não bloquear OU algum mecanismo especifico de criação de subscribers que não bloqueie o envio (Qual? Tradeoffs?)
}

// Close fecha o event bus e todos os channels de subscribers
func (eb *EventBus) Close() {
    // TODO:
    // 1. Fechar todos os channels
    // 2. Limpar subscribers
}
```

## Log Aggregator (Fan-in)

### O Desafio

Você tem múltiplos serviços (API, Database, Cache) gerando logs concorrentemente. Precisa de um agregador central que coleta todos esses logs e garante que nenhum se perde.

Aprendizados desse challenge:

✅ Fan-in pattern (N -> 1) - múltiplos producers enviando para um único channel intermediário (bridge) que é consumido por uma goroutine
✅ for range em channels - o padrão idiomático para processar todos os valores até o channel fechar, bloqueia automaticamente quando não há dados
✅ Select com default cria busy-waiting - evitar usar default quando você quer bloquear esperando dados, channels já dormem eficientemente
✅ break dentro de select - só quebra o select, não o for externo (use return para sair da goroutine inteira)
✅ WaitGroup regra crítica: Add antes de Wait - nunca chamar Wait() antes que todos os Add() tenham sido executados, causa race condition no contador interno
✅ WaitGroup para coordenação - Add(1) antes de criar goroutine, defer wg.Done() na goroutine, Wait() para bloquear até todas terminarem
✅ Onde criar a goroutine auxiliar importa - se criar no Start() e ela chamar Wait() imediatamente, vai retornar antes dos Register() adicionarem ao WaitGroup; melhor chamar Wait() no Stop()
✅ Variable shadowing com := - usar bridge := make(chan) cria variável local que esconde la.bridge, deixando o campo do struct como nil; sempre usar la.bridge = make(chan) para inicializar campos
✅ Inicializar bridge no lugar certo - criar no Start() ao invés do construtor, senão a goroutine auxiliar pode fechar o channel antes de qualquer Register() acontecer
✅ Done channel para sincronizar término - usar chan struct{} que a goroutine consumidora fecha quando termina, permitindo Stop() esperar sem race condition em la.logs
✅ defer dentro de loop acumula - não executa a cada iteração, só no final da função (causa deadlock com mutex em loops)
✅ Single writer pattern - quando só uma goroutine escreve em uma estrutura, não precisa de mutex (sem race condition)
✅ Ownership de channels - quem cria o channel deve fechá-lo, não tente fechar channels de terceiros (causa panic)
✅ Graceful shutdown em camadas - producers fecham channels -> leitoras terminam e Done() -> Stop() fecha bridge -> consumidora termina e fecha done channel -> Stop() retorna logs
✅ Channel buffering - buffer no bridge permite leitoras continuarem enviando mesmo se consumidora estiver ocupada (trade-off latência vs throughput)
✅ Send on closed channel panic - tentar enviar para channel fechado causa panic; garantir que bridge só é fechado quando todas as goroutines leitoras já terminaram

### Os 3 Problemas Principais

**1. Fan-in (N -> 1)**: Como ler de múltiplos channels ao mesmo tempo?
- Não dá pra fazer `for` em cada channel separado (seria sequencial)
- Precisa de algo que espera em TODOS simultaneamente

**2. Graceful Shutdown**: Como saber quando TODOS os producers terminaram?
- Um producer fecha seu channel... mas ainda tem outros rodando
- Você só pode parar quando o ÚLTIMO fechar

**3. Coordenação**: Como avisar o aggregator que pode parar?
- Quem vai fechar o channel central?
- Como sincronizar isso com os producers?

### De forma simples:

**Fan-in**

- Vários produtores.
- Um consimudor.

**Produtores:** criam logs, api, database, etc
**Consumidor:** aggregator.

Aggregator deve registrar os channels para ler depois. Estrutura simples.

**O que é Fan-in de verdade?**

Fan-in significa que você precisa estar lendo de TODOS os channels ao mesmo tempo.
Quando qualquer um deles tiver um log pronto, processa. É concorrente, não sequencial.

### Requisitos Funcionais

1. **LogAggregator deve permitir:**
   - `Register(logChan <-chan LogEntry)` - registra um channel de producer para agregar
   - `Start()` - inicia o processamento de logs de todos os producers
   - `Stop() []LogEntry` - para gracefully e retorna todos os logs coletados

2. **Comportamento esperado:**
   - Múltiplos producers podem enviar logs simultaneamente
   - Nenhum log pode ser perdido (todos devem ser coletados)
   - Se um producer termina antes dos outros, seus logs já devem estar agregados
   - `Stop()` só retorna quando TODOS os producers terminaram e todos os logs foram processados
   - Producers mais lentos não bloqueiam producers mais rápidos

### Constraints

- Não usar libs externas (só stdlib)
- Tempo: ~1 hora
- Código: 80-160 linhas
- Usar `sync` package conforme necessário
- Implementar graceful shutdown sem perder logs

### Aprendizados desse challenge:

### Template

```go

func main() {
	fmt.Println("=== Log Aggregator Challenge ===")
	fmt.Println("Fan-in Pattern: N producers -> 1 aggregator")

	agg := NewLogAggregator()

	// simulando 3 serviços diferentes gerando logs
	services := []string{"api", "database", "cache"}

	for _, service := range services {
		logChan := make(chan LogEntry, 10)
		agg.Register(logChan)

		// Cada serviço roda em sua própria goroutine
		go func(name string, ch chan LogEntry) {
			for i := 0; i < 5; i++ {
				ch <- LogEntry{
					Timestamp: time.Now(),
					Source:    name,
					Level:     "INFO",
					Message:   fmt.Sprintf("Log %d from %s", i+1, name),
				}
				time.Sleep(time.Millisecond * 10) // simulando trabalho
			}
			fmt.Printf("[%s] Finished producing logs\n", name)
			close(ch) // PRECIOSAMOS fechar o channel depois de terminar
		}(service, logChan)
	}

	agg.Start()

	// aguardar logs serem gerados
	time.Sleep(time.Millisecond * 200)

	fmt.Println("\nStopping aggregator...")
	logs := agg.Stop()

	fmt.Printf("\n=== Collected %d logs ===\n", len(logs))

	// agrupa por source
	bySource := make(map[string]int)
	for _, log := range logs {
		bySource[log.Source]++
		// fmt.Printf("[%s] %s: %s\n", log.Source, log.Level, log.Message)
	}

	fmt.Println("\nLogs per service:")
	for service, count := range bySource {
		fmt.Printf("  %s: %d logs\n", service, count)
	}

	fmt.Println("\nRun 'go test -v' to verify your implementation")
	fmt.Println("Run 'go test -race' to check for race conditions")
}


// LogEntry representa um log de qualquer serviço
type LogEntry struct {
	Timestamp time.Time
	Source    string // "api", "database", "cache", etc
	Level     string // "DEBUG", "INFO", "WARN", "ERROR"
	Message   string
}

// LogProducer simula um serviço gerando logs
type LogProducer struct {
	id       string
	logsChan chan LogEntry
}

// LogAggregator é o coração do Fan-in pattern
type LogAggregator struct {
	// TODO: adicione campos necessários para:
	// - Armazenar channels dos producers
	// - Coordenar o shutdown de todos producers
	// - Coletar logs processados
}

func NewLogAggregator() *LogAggregator {
	return &LogAggregator{
		// TODO: inicialize seus campos
	}
}

// Register adiciona um novo producer ao aggregator
func (la *LogAggregator) Register(logChan <-chan LogEntry) {
	// TODO: armazene o channel do producer
	// Hint: você vai precisar ler de TODOS esses channels depois
}

// Start inicia o processamento de logs
func (la *LogAggregator) Start() {
	// TODO: inicie uma goroutine que:
	// 1. Lê de TODOS os channels registrados (Fan-in!)
	// 2. Processa cada log recebido
	// 3. Para quando todos os producers terminarem

	// Hint: como ler de múltiplos channels ao mesmo tempo?
	// Hint: como saber quando TODOS os channels foram fechados?
}

// Stop para o aggregator e retorna todos os logs coletados
func (la *LogAggregator) Stop() []LogEntry {
	// TODO: implemente graceful shutdown
	// - Espere todos os producers terminarem
	// - Retorne os logs coletados

	// Hint: você precisa sinalizar que quer parar E esperar
	// que o processamento realmente termine

	return nil // TODO: retorne os logs
}

```

## Image Pipeline

Você vai construir um pipeline de 4 stages onde cada stage processa imagens e passa para o próximo, tudo rodando concorrentemente. É como uma linha de montagem, enquanto o Stage 1 está listando novos arquivos, o Stage 2 já está carregando imagens anteriores, o Stage 3 processando (converter em gray scale), e o Stage 4 salvando.

**O que fazer**: stages encadeados onde cada um procesa e passa para o próximo.

**Importante**: No Fan-In tinhamos multiplos produtores enviando para UM channel central. Aqui temos uma **corrente de channels**:

```
Generator → [channel1] → Loader → [channel2] → Processor → [channel3] → Saver
   (lista)              (carrega)              (processa)              (salva)
   arquivos             imagem                 grayscale               disco
```

### Os 4 Stages

**Stage 1** - Generator: Lista todos os arquivos .jpg e .png do diretório de input

- Saída: channel de ImageJob com apenas o Path preenchido

**Stage 2** - Loader: Carrega cada imagem do disco para memória

- Entrada: jobs com path
- Saída: jobs com Image carregado

**Stage 3** - Processor: Converte imagens para grayscale

- Entrada: jobs com imagem colorida
- Saída: jobs com imagem em tons de cinza

**Stage 4** - Saver: Salva imagens processadas no disco

- Entrada: jobs com imagem processada
- Saída: nenhuma (é o fim do pipeline)

### Os 3 Desafios Principais

#### 1. Conectar os Stages com Channels

- Cada stage lê de um channel e escreve no próximo
- Você precisa criar os channels intermediários e conectar tudo no Run()

#### 2. Graceful Shutdown

- Cada stage precisa fechar seu output channel quando o input channel fechar
- Como saber quando o pipeline inteiro terminou? (WaitGroup para cada stage?)

#### 3. Context Cancellation

- Se o context cancelar (timeout ou erro), todos os stages devem parar imediatamente
- Use select com ctx.Done() em pontos estratégicos

### Como começar

1. Implementar o ```generator```
	- Use ```filepath.Glob()``` para listar os arquivos

```go
files, _ := filepath.Glob(filepath.Join(inputDir, "*.jpg"))
```

2. Implementar o ```loader```
	- Use ```image.Decode()```

```go
file, _ := os.Open(job.Path)
img, _, _ := image.Decode(file)
```

3. Implemente o ```processor```
	- Loop pelos pixels e converta para grayscale:

```
grayImg := image.NewGray(img.Bounds())
// loop pelos pixels, calcular média RGB, setar no grayImg
```

4. Implementar o ```saver```. Use ```jpeg.Encode()```

jpeg.Encode(file, job.Image, &jpeg.Options{Quality: 90})

5. Conecte tudo no ```Run()``` criando channels e lançando goroutines

### Dicas Importantes

- Fechar channels: Cada stage deve fazer defer close(outputChan) no início
- WaitGroup: Você vai precisar de um WaitGroup para cada stage ou um para todos? Pensa no fluxo
- Buffered channels: Use buffer nos channels intermediários (ex: 10) para evitar bloqueios
- Context: Em cada loop, faça select { case <-ctx.Done(): return } para respeitar cancellation

### Para Testar

```
# Rodar testes
go test -v

# Verificar race conditions
go test -race
```

### Main e assinaturas:

```go
package main

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"
)

func main() {
	fmt.Println("=== Image Processing Pipeline ===")
	fmt.Println("Pipeline Pattern: Generator → Loader → Processor → Saver\n")

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
	fmt.Println("\nRun 'go test -v' to verify your implementation")
}

// ImageJob representa uma imagem no pipeline
type ImageJob struct {
	Path   string      // caminho do arquivo original
	Image  image.Image // imagem carregada (nil nos primeiros stages)
	Error  error       // erro se algo deu errado
	StageNum int       // apenas para debug
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
	// Generator → fileChan → Loader → imageChan → Processor → processedChan → Saver
	
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
```