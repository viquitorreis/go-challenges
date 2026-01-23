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

	fmt.Println("Stopping aggregator...")
	logs := agg.Stop()

	fmt.Printf("\n=== Collected %d logs ===\n", len(logs))

	// agrupa por source
	bySource := make(map[string]int)
	for _, log := range logs {
		bySource[log.Source]++
		// fmt.Printf("[%s] %s: %s\n", log.Source, log.Level, log.Message)
	}

	fmt.Println("Logs per service:")
	for service, count := range bySource {
		fmt.Printf("  %s: %d logs\n", service, count)
	}

	fmt.Println("Run 'go test -v' to verify your implementation")
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
Generator -> [channel1] -> Loader -> [channel2] -> Processor -> [channel3] -> Saver
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
```

## TCP Chat Server

Este desafio é diferente dos anteriores porque você vai trabalhar com sockets TCP de baixo nível. Quando alguém conecta com telnet localhost 6969, você está recebendo uma conexão TCP bruta. Você precisa ler bytes dessa conexão, interpretar como mensagens, e enviar bytes de volta. É exatamente assim que servidores reais funcionam por baixo do capô.

### A Arquitetura do Sistema

Pensa num chat como Discord ou Slack. Quando você envia uma mensagem, ela precisa chegar em todos os outros usuários que estão online naquele canal. Mas tem algumas complexidades interessantes aqui.

- Primeiro, cada usuário tem sua própria conexão de rede independente, você não pode simplesmente "gritar" a mensagem e todo mundo ouve. Você precisa enviar ativamente para cada conexão individual.
- Segundo, essas conexões podem estar em velocidades diferentes, um usuário pode estar numa conexão lenta de celular enquanto outro está em fibra ótica. Se você esperar o usuário lento responder antes de enviar para o próximo, todos ficam esperando.
- Terceiro, usuários podem desconectar a qualquer momento sem avisar, o cabo de rede pode ser desconectado fisicamente.

### Os três Componentes Principais

Você vai construir três sistemas que trabalham juntos.

1. O primeiro é o **sistema de lobby**, que é uma sala de espera. Imagine que você está organizando um jogo de futebol e precisa esperar pelo menos dois jogadores chegarem antes de começar. Enquanto só tem um jogador, ele fica esperando. Quando o segundo chega, você libera ambos para jogar. Este sistema usa **sync.Cond**, que é uma primitiva de sincronização perfeita para "acordar" múltiplas goroutines quando uma condição muda. É como um alarme que toca para todo mundo ao mesmo tempo.

2. O segundo componente é o **sistema de broadcast**. Quando um cliente envia uma mensagem, você precisa distribuir para todos os outros clientes conectados. A forma mais idiomática em Go é dar para cada cliente seu próprio channel de mensagens. Quando alguém envia uma mensagem, você itera pelos clientes e envia a mensagem para o channel de cada um. Cada cliente tem uma goroutine dedicada lendo desse channel e escrevendo na conexão TCP. Isso resolve o problema de clientes lentos - se o channel de um cliente encher porque ele está lento, você simplesmente pula ele ao invés de bloquear todos os outros.

3. O terceiro componente é o **gerenciamento de conexões**. Cada cliente precisa de pelo menos duas goroutines, uma que fica lendo da conexão TCP (esperando mensagens chegarem) e outra que fica lendo do channel de mensagens e escrevendo na conexão TCP (enviando mensagens para o cliente). Quando um cliente desconecta, você precisa limpar esses recursos - fechar goroutines, remover do map de clientes ativos, fechar channels. Se não fizer isso direito, você tem vazamento de goroutines e memória.

### Como começar

Comece pelo mais simples possível. Implemente apenas o ```Start``` e faça ele aceitar uma conexão. Nem precisa fazer nada com a conexão ainda, só aceitar e imprimir que alguém conectou. Teste com telnet localhost 6969 em outro terminal. Quando você vir "cliente conectou", você sabe que a parte de networking básica funciona.


```go
package main

import (
	"context"
	"net"
	"time"
)

/*
═══════════════════════════════════════════════════════════════════════════
TODO - PASSOS PARA IMPLEMENTAR O TCP CHAT SERVER
═══════════════════════════════════════════════════════════════════════════

1. CRIAR O SERVER
   - Usar net.Listen para escutar em uma porta (ex: :6969)
   - Aceitar conexões de clientes com Accept() em um loop
   - Para cada cliente que conecta, criar uma goroutine

2. LOBBY (SALA DE ESPERA)
   - Usar sync.Cond para bloquear clientes até ter mínimo de 2 players
   - Quando cliente conecta, incrementar contador
   - Se atingir mínimo, fazer Broadcast() para liberar todos
   - Clientes ficam esperando em Wait() até o Broadcast

3. GERENCIAR CLIENTES
   - Guardar cada cliente em um map (clientID -> Client)
   - Cada cliente precisa de um channel para receber mensagens
   - Quando cliente envia mensagem, fazer broadcast para TODOS os outros

4. BROADCAST DE MENSAGENS
   - Ler mensagem do cliente A
   - Enviar para os channels de todos os outros clientes (B, C, D...)
   - Usar goroutines separadas: uma lê, outra escreve

5. HANDLE DISCONNECT
   - Quando cliente desconecta, remover do map
   - Notificar outros clientes
   - Fechar o channel do cliente que saiu

6. GRACEFUL SHUTDOWN
   - Context para cancelar tudo quando server parar
   - Fechar todas as conexões ativas
   - Esperar goroutines terminarem com WaitGroup

═══════════════════════════════════════════════════════════════════════════
*/


func main() {
	fmt.Println("=== TCP Chat Server com Lobby ===")
	fmt.Println("Network Programming: TCP + sync.Cond + Broadcast")

	// Criar servidor que espera mínimo de 2 players
	server := NewChatServer("6969", 2)

	// Context com timeout de 5 minutos
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Iniciar servidor em goroutine
	go func() {
		fmt.Println("👽 Server listening on :6969")
		fmt.Println("Connect with: telnet localhost 6969")
		fmt.Println("Waiting for at least 2 players to start...")
		
		if err := server.Start(ctx); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Simular alguns clients para teste (remova isso depois)
	time.Sleep(1 * time.Second)
	
	fmt.Println("To test:")
	fmt.Println("   Terminal 1: telnet localhost 6969")
	fmt.Println("   Terminal 2: telnet localhost 6969")
	fmt.Println("   Type messages and press Enter")
	fmt.Println("   Press Ctrl+C to stop server")

	// Manter servidor rodando
	<-ctx.Done()
	fmt.Println("Timeout reached, stopping server...")


	server.Stop()
	fmt.Println("Server stopped")
}

// Message representa uma mensagem no chat
type Message struct {
	From    string
	Content string
	Time    time.Time
}

// Client representa um cliente conectado
type Client struct {
	// TODO: adicione campos necessários:
	// - ID único do cliente
	// - Conexão net.Conn
	// - Channel para receber mensagens
	// - Reader/Writer para ler/escrever na conexão
}

// ChatServer gerencia o servidor de chat
type ChatServer struct {
	// TODO: adicione campos necessários:
	// - Porta do servidor
	// - Map de clientes ativos
	// - Mutex para proteger o map
	// - sync.Cond para o lobby
	// - Contador de clientes
	// - Mínimo de players para começar
	// - WaitGroup para coordenação
}

type IChatServer interface {
	Start(ctx context.Context) error
	Stop() error
}

// NewChatServer cria um novo servidor de chat
func NewChatServer(port string, minPlayers int) IChatServer {
	return &ChatServer{
		// TODO: inicialize os campos
	}
}

// Start inicia o servidor TCP
func (s *ChatServer) Start(ctx context.Context) error {
	// TODO:
	// 1. Criar listener TCP com net.Listen("tcp", ":"+port)
	// 2. Loop infinito aceitando conexões com listener.Accept()
	// 3. Para cada conexão aceita, criar um Client e chamar handleClient em goroutine
	// 4. Respeitar ctx.Done() para shutdown

	// Hint:
	// listener, err := net.Listen("tcp", ":6969")
	// for { conn, err := listener.Accept() }

	return nil
}

// handleClient gerencia um cliente conectado
func (s *ChatServer) handleClient(ctx context.Context, conn net.Conn) {
	// TODO:
	// 1. Criar um Client com ID único e conexão
	// 2. Adicionar cliente ao map (thread-safe com mutex)
	// 3. LOBBY: Verificar se atingiu mínimo de players
	//    - Se sim, fazer Broadcast() para liberar todos
	//    - Se não, fazer Wait() para esperar
	// 4. Iniciar duas goroutines:
	//    - readLoop: lê mensagens do cliente
	//    - writeLoop: envia mensagens para o cliente
	// 5. Quando cliente desconectar, remover do map e notificar outros

	defer conn.Close()

	// Hint: Use sync.Cond para o lobby
	// s.cond.L.Lock()
	// s.playerCount++
	// if s.playerCount >= s.minPlayers {
	//     s.cond.Broadcast()
	// } else {
	//     s.cond.Wait()
	// }
	// s.cond.L.Unlock()
}

// readLoop lê mensagens de um cliente
func (s *ChatServer) readLoop(ctx context.Context, client *Client) {
	// TODO:
	// 1. Criar um bufio.Scanner ou bufio.Reader na conexão
	// 2. Loop lendo linhas da conexão
	// 3. Para cada mensagem recebida, fazer broadcast para todos outros clientes
	// 4. Se erro de leitura (cliente desconectou), sair do loop

	// Hint:
	// scanner := bufio.NewScanner(client.conn)
	// for scanner.Scan() {
	//     msg := scanner.Text()
	//     s.broadcast(client.ID, msg)
	// }
}

// writeLoop envia mensagens para um cliente
func (s *ChatServer) writeLoop(ctx context.Context, client *Client) {
	// TODO:
	// 1. Loop lendo do channel de mensagens do cliente
	// 2. Para cada mensagem, escrever na conexão
	// 3. Se ctx cancelar ou channel fechar, sair

	// Hint:
	// for msg := range client.messages {
	//     fmt.Fprintf(client.conn, "%s: %s\n", msg.From, msg.Content)
	// }
}

// broadcast envia uma mensagem para todos os clientes exceto o remetente
func (s *ChatServer) broadcast(fromID string, content string) {
	// TODO:
	// 1. Criar uma mensagem
	// 2. Iterar pelo map de clientes (thread-safe com mutex)
	// 3. Para cada cliente diferente do remetente:
	//    - Enviar mensagem para o channel do cliente
	//    - Usar select com default para não bloquear se channel cheio

	// Hint: Use select com default para não bloquear
	// select {
	// case client.messages <- msg:
	// default:
	//     // Cliente está lento/desconectado, pular
	// }
}

// removeClient remove um cliente e notifica outros
func (s *ChatServer) removeClient(clientID string) {
	// TODO:
	// 1. Lock no mutex
	// 2. Remover cliente do map
	// 3. Fechar o channel de mensagens do cliente
	// 4. Unlock no mutex
	// 5. Fazer broadcast que o cliente saiu
	// 6. Decrementar playerCount
}

// Stop para o servidor gracefully
func (s *ChatServer) Stop() error {
	// TODO:
	// 1. Fechar listener
	// 2. Fechar todas as conexões ativas
	// 3. Esperar WaitGroup terminar

	return nil
}
```

- The test file is: 

```go
package main

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
)

func TestServer_AcceptsConnections(t *testing.T) {
	server := NewChatServer("18080", 1) // porta diferente para não conflitar

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Iniciar servidor
	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond) // dar tempo pro servidor subir

	// Conectar cliente
	conn, err := net.Dial("tcp", "localhost:18080")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	t.Log("Client connected successfully")
}

func TestServer_LobbyWaitsForMinPlayers(t *testing.T) {
	server := NewChatServer("18081", 2) // precisa de 2 players

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Conectar primeiro cliente
	conn1, err := net.Dial("tcp", "localhost:18081")
	if err != nil {
		t.Fatalf("Client 1 failed to connect: %v", err)
	}
	defer conn1.Close()

	// Dar tempo para entrar no lobby
	time.Sleep(200 * time.Millisecond)

	// Conectar segundo cliente
	conn2, err := net.Dial("tcp", "localhost:18081")
	if err != nil {
		t.Fatalf("Client 2 failed to connect: %v", err)
	}
	defer conn2.Close()

	// Quando 2o cliente conecta, ambos devem ser liberados do lobby
	time.Sleep(200 * time.Millisecond)

	t.Log("Lobby released after 2 players connected")
}

func TestServer_BroadcastMessage(t *testing.T) {
	server := NewChatServer("18082", 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Conectar dois clientes
	conn1, err := net.Dial("tcp", "localhost:18082")
	if err != nil {
		t.Fatalf("Client 1 failed: %v", err)
	}
	defer conn1.Close()

	conn2, err := net.Dial("tcp", "localhost:18082")
	if err != nil {
		t.Fatalf("Client 2 failed: %v", err)
	}
	defer conn2.Close()

	// Esperar lobby liberar
	time.Sleep(300 * time.Millisecond)

	// Cliente 1 envia mensagem
	fmt.Fprintf(conn1, "Hello from client 1\n")

	// Cliente 2 deve receber
	reader := bufio.NewReader(conn2)
	conn2.SetReadDeadline(time.Now().Add(2 * time.Second))

	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Client 2 didn't receive message: %v", err)
	}

	if !strings.Contains(line, "Hello from client 1") {
		t.Errorf("Expected message with 'Hello from client 1', got: %s", line)
	}

	t.Log("Message broadcast successfully")
}

func TestServer_MultipleClients(t *testing.T) {
	server := NewChatServer("18083", 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Conectar 3 clientes
	var conns []net.Conn
	for i := 0; i < 3; i++ {
		conn, err := net.Dial("tcp", "localhost:18083")
		if err != nil {
			t.Fatalf("Client %d failed: %v", i+1, err)
		}
		conns = append(conns, conn)
		defer conn.Close()
	}

	// Esperar lobby liberar
	time.Sleep(300 * time.Millisecond)

	// Cliente 1 envia mensagem
	fmt.Fprintf(conns[0], "Broadcast test\n")

	// Clientes 2 e 3 devem receber
	for i := 1; i < 3; i++ {
		reader := bufio.NewReader(conns[i])
		conns[i].SetReadDeadline(time.Now().Add(2 * time.Second))

		line, err := reader.ReadString('\n')
		if err != nil {
			t.Errorf("Client %d didn't receive: %v", i+1, err)
			continue
		}

		if !strings.Contains(line, "Broadcast test") {
			t.Errorf("Client %d got wrong message: %s", i+1, line)
		}
	}

	t.Log("Message broadcast to all clients")
}

func TestServer_ClientDisconnect(t *testing.T) {
	server := NewChatServer("18084", 2)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Conectar 3 clientes
	conn1, _ := net.Dial("tcp", "localhost:18084")
	conn2, _ := net.Dial("tcp", "localhost:18084")
	conn3, _ := net.Dial("tcp", "localhost:18084")

	time.Sleep(300 * time.Millisecond)

	// Cliente 1 desconecta
	conn1.Close()
	time.Sleep(200 * time.Millisecond)

	// Cliente 2 envia mensagem
	fmt.Fprintf(conn2, "After disconnect\n")

	// Cliente 3 deve receber (cliente 1 não, pois desconectou)
	reader := bufio.NewReader(conn3)
	conn3.SetReadDeadline(time.Now().Add(2 * time.Second))

	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("Client 3 should still receive: %v", err)
	}

	if !strings.Contains(line, "After disconnect") {
		t.Errorf("Got wrong message: %s", line)
	}

	conn2.Close()
	conn3.Close()

	t.Log("Server handles disconnect correctly")
}

func TestServer_EmptyMessage(t *testing.T) {
	server := NewChatServer("18085", 1)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", "localhost:18085")
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Enviar mensagem vazia
	fmt.Fprintf(conn, "\n")

	time.Sleep(100 * time.Millisecond)

	t.Log("Server handles empty messages")
}

// Teste de stress - múltiplas mensagens rápidas
func TestServer_RapidMessages(t *testing.T) {
	server := NewChatServer("18086", 2)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go server.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	conn1, _ := net.Dial("tcp", "localhost:18086")
	defer conn1.Close()

	conn2, _ := net.Dial("tcp", "localhost:18086")
	defer conn2.Close()

	time.Sleep(300 * time.Millisecond)

	// Enviar 50 mensagens rápidas
	go func() {
		for i := 0; i < 50; i++ {
			fmt.Fprintf(conn1, "Message %d\n", i)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Cliente 2 deve receber todas (ou a maioria se buffer encher)
	reader := bufio.NewReader(conn2)
	received := 0

	for i := 0; i < 50; i++ {
		conn2.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		received++
	}

	if received < 45 { // Aceitar perder algumas mensagens por buffer cheio
		t.Errorf("Only received %d/50 messages", received)
	} else {
		t.Logf("Received %d/50 rapid messages", received)
	}
}
```

## Rate Limiter Com Token Bucket

Você vai implementar um rate limiter thread-safe usando o algoritmo **Token Bucket**. 

Este é um componente fundamental em sistemas de produção, toda API pública usa alguma forma de rate limiting. Empresas como Stripe, GitHub, e AWS usam variações desse algoritmo para proteger seus serviços contra sobrecarga.

**O que você vai construir:** Um rate limiter que aceita ou rejeita requisições baseado em um bucket de tokens que se reabastece continuamente. Múltiplas goroutines vão tentar consumir tokens simultaneamente enquanto uma goroutine em background adiciona tokens periodicamente.

### Introdução

**Por quê fazer esse challenge?**

Rate limiting aparece em todo lugar em engenharia de backend. Quando você faz uma chamada para a API do GitHub, eles te dão 5000 requests/hora. Quando você usa Redis Cloud, eles limitam operações por segundo. Quando você constrói uma API REST, você precisa proteger seus endpoints de abuse.

**Empresas que usam:** Stripe (payment processing), CloudFlare (DDoS protection), Redis (cloud limits), toda API REST moderna. Em entrevistas para Tailscale, LiveKit, ou GitLab, e FAANGS saber implementar rate limiting do zero é diferencial forte.

**O que exatamente você vai construir?**

Um rate **limiter** baseado em **Token Bucket** que:

- Mantém um bucket com capacidade máxima de N tokens
- Adiciona tokens ao bucket a uma taxa constante (ex: 10 tokens/segundo)
- Quando uma requisição chega, tenta consumir 1 token
- Se tem token disponível -> aceita a requisição
- Se não tem token -> rejeita (ou opcionalmente espera)

Você vai testar isso simulando 100 goroutines fazendo requests simultaneamente. Algumas vão passar, outras vão ser rejeitadas conforme os tokens esgotam e se reabastecem.

**Quais são os 2-3 desafios principais?**

1. **Reabastecimento contínuo de tokens:** Como adicionar tokens periodicamente sem criar um timer por requisição? Pense direito para não gerar goroutines leaks.
2. **sincronização entre consumidores e producers:** Múltiplas goroutines tentando consumir tokens simultaneamente, enquanto uma goroutine adiciona tokens. Você precisa de mutex, mas onde? Se for um mutex geral, vai ter uma performance ruim. Se não sincronizar corretamente, perde token ou conta errado.
3. **Modelar o bucket de forma eficiente:** *Não é permitido* ter um array gigante com todos os tokens. Precisa apenas rastrear quantos tokens existem atualmente e quando foi a última vez que adicionou tokens. É um problema de modelagem de estado.

### Background Técnico

**Pattern: Token Bucket Algorithm**

Imagine um balde que comporta no máximo 100 moedas (tokens). A cada segundo, alguém joga 10 moedas novas no balde. Se já está cheio, as moedas extras caem fora (capacidade máxima). Quando uma pessoa quer fazer alguma coisa, ela precisa pegar uma moeda do balde. Se não tem moeda disponível, ela não pode prosseguir.

**Token bucket vs Leaky Bucket:** Token Bucket permite bursts (se o bucket está cheio, você pode consumir vários tokens rapidamente). Leaky Bucket Mantém taxa constante (como um funil que drena em velocidade fixa). Para APIs, Token Bucket é mais comum pois permite clientes fazer pequenos bursts sem punição.

#### DSA Usada

Por mais que tenha "bucket" no nome, a implementação não tem um balde como estrutura literal. O truque está em **rastrear estado em vez de elementos individuais:**

```
Estado do bucket:
- tokens (int): quantos tokens existem agora
- capacity (int): máximo de tokens permitidos
- refillRate (int): tokens adicionados por intervalo
- lastRefill (time.Time): quando foi o último reabastecimento
```

Quando alguém tenta consumir um token, primeiro deve calcular **quantos tokens deveriam ter sido adicionados desde lastRefill**, adiciona eles respeitando a capcidade, e atualiza lastRefill, e então tenta consumir.

Isso é mais eficiente do que usar um timer, pois usa **lazy evaluation**, só calcula tokens quando precisam, não fica acordando goroutine a cada tick...

### Libs Go Relavantes

- sync.Mutex: Proteger leitura/escrita do estado do bucket
- time.Time e time.Since(): Calcular quanto tempo passou desde último refill
- time.Duration: Definir intervalos de refill (ex: 100ms)

**NÃO precisa de**:

- time.Ticker (abordagem mais simples usa lazy evaluation)
- Channels (adiciona complexidade desnecessária aqui)

### Trade-offs Principais

**Precisão vs Performance**

- Usar mutex para cada operação garante precisão mas pode criar contenção em alta carga
- Usar atomic operations é mais rápido mas dificulta lógica de refill
- Para este challenge, prefira clareza e correção (use mutex)

**Lazy Refill vs Active Refill**

- Lazy: calcula tokens na hora da requisição (mais simples, menos goroutines)
- Active: goroutine em background usando ticker (mais preciso em cenários específicos)
- Vamos usar Lazy porque é mais idiomático para este caso

**Rejeitar vs bloquear**

- Rejeitar imediatamente: retorna erro se não tem token
- Bloquear e esperar: goroutine espera até token ficar disponível
- Vamos implementar rejeição, pois bloquear complica testes e pode causar deadlocks.

### Função main:

```go
package main

import (
	"fmt"
	"sync"
	"time"
)

// TokenBucket implementa rate limiting usando o algoritmo Token Bucket
type TokenBucket struct {
	// TODO: adicionar campos necessários
	// Dica: você precisa rastrear:
	// - quantos tokens existem agora
	// - capacidade máxima
	// - taxa de reabastecimento (tokens por intervalo)
	// - intervalo de reabastecimento
	// - timestamp do último refill
	// - mutex para proteger acesso concorrente
}

// NewTokenBucket cria um novo rate limiter
// capacity: número máximo de tokens no bucket
// refillRate: quantos tokens adicionar por intervalo
// refillInterval: com que frequência adicionar tokens (ex: 100ms)
func NewTokenBucket(capacity int, refillRate int, refillInterval time.Duration) *TokenBucket {
	// TODO: inicializar bucket começando cheio (todos os tokens disponíveis)
	return nil
}

// Allow tenta consumir um token
// Retorna true se conseguiu (requisição permitida)
// Retorna false se não tem tokens disponíveis (requisição rejeitada)
func (tb *TokenBucket) Allow() bool {
	// TODO: implementar lógica
	// Passos:
	// 1. Lock do mutex
	// 2. Calcular quantos tokens deveriam ter sido adicionados desde lastRefill
	// 3. Adicionar esses tokens (respeitando capacidade máxima)
	// 4. Atualizar lastRefill para agora
	// 5. Tentar consumir 1 token
	// 6. Unlock e retornar resultado
	return false
}

// refill é um método helper para calcular e adicionar tokens
// Retorna quantos tokens foram adicionados
func (tb *TokenBucket) refill() int {
	// TODO: implementar cálculo de refill
	// Quanto tempo passou desde lastRefill?
	// Quantos "intervalos" completos aconteceram nesse tempo?
	// Quantos tokens isso representa?
	// Não esqueça de respeitar a capacidade máxima!
	return 0
}

// Stats retorna estatísticas atuais do bucket (útil para debugging)
func (tb *TokenBucket) Stats() (available int, capacity int) {
	// TODO: retornar tokens disponíveis e capacidade
	// Precisa de lock? Por quê?
	return 0, 0
}

func main() {
	// Criar rate limiter: 10 tokens no máximo, adiciona 5 tokens a cada 100ms
	// Isso dá ~50 requests/segundo no steady state
	limiter := NewTokenBucket(10, 5, 100*time.Millisecond)

	fmt.Println("Starting rate limiter test...")
	fmt.Printf("Bucket: capacity=%d, refill rate=%d tokens per 100ms\n", 10, 5)
	
	// TODO: Simular múltiplas goroutines fazendo requests
	// Dica: use WaitGroup, lance ~50 goroutines, cada uma tenta Allow()
	// Conte quantas foram aceitas vs rejeitadas
	// Print os resultados
	
	fmt.Println("Test completed!")
}
```

### Testes

```go
package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestTokenBucket_InitialTokens(t *testing.T) {
	// Bucket deve começar cheio
	tb := NewTokenBucket(10, 5, 100*time.Millisecond)
	
	available, capacity := tb.Stats()
	if available != capacity {
		t.Errorf("expected bucket to start full: got %d/%d", available, capacity)
	}
}

func TestTokenBucket_ConsumeAllTokens(t *testing.T) {
	tb := NewTokenBucket(5, 1, 100*time.Millisecond)
	
	// Deve conseguir consumir exatamente 5 tokens
	for i := 0; i < 5; i++ {
		if !tb.Allow() {
			t.Fatalf("expected token %d to be available", i+1)
		}
	}
	
	// Sexto deve falhar
	if tb.Allow() {
		t.Error("expected 6th request to be rejected")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	// Bucket pequeno: 5 tokens, adiciona 5 a cada 50ms
	tb := NewTokenBucket(5, 5, 50*time.Millisecond)
	
	// Consumir todos
	for i := 0; i < 5; i++ {
		tb.Allow()
	}
	
	// Verificar que está vazio
	if tb.Allow() {
		t.Error("bucket should be empty")
	}
	
	// Esperar tempo suficiente para refill completo
	time.Sleep(60 * time.Millisecond)
	
	// Deve ter 5 tokens novamente
	allowed := 0
	for i := 0; i < 10; i++ {
		if tb.Allow() {
			allowed++
		}
	}
	
	if allowed != 5 {
		t.Errorf("expected 5 tokens after refill, got %d", allowed)
	}
}

func TestTokenBucket_ConcurrentAccess(t *testing.T) {
	tb := NewTokenBucket(100, 10, 50*time.Millisecond)
	
	var allowed atomic.Int32
	var rejected atomic.Int32
	
	var wg sync.WaitGroup
	// 200 goroutines competindo por 100 tokens
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tb.Allow() {
				allowed.Add(1)
			} else {
				rejected.Add(1)
			}
		}()
	}
	
	wg.Wait()
	
	total := int(allowed.Load()) + int(rejected.Load())
	if total != 200 {
		t.Errorf("expected 200 total requests, got %d", total)
	}
	
	// Deve ter aceito aproximadamente 100 (pode variar um pouco por causa de refill)
	if allowed.Load() < 95 || allowed.Load() > 105 {
		t.Errorf("expected ~100 allowed, got %d", allowed.Load())
	}
}

func TestTokenBucket_RateLimiting(t *testing.T) {
	// 10 tokens, refill 10 a cada 100ms = 100 tokens/segundo
	tb := NewTokenBucket(10, 10, 100*time.Millisecond)
	
	start := time.Now()
	allowed := 0
	
	// Tentar consumir 50 tokens o mais rápido possível
	for allowed < 50 {
		if tb.Allow() {
			allowed++
		} else {
			time.Sleep(10 * time.Millisecond) // Pequeno sleep para não criar busy loop
		}
	}
	
	elapsed := time.Since(start)
	
	// 50 tokens a 100/segundo deveria levar ~500ms
	// Damos margem de erro (300ms a 700ms)
	if elapsed < 300*time.Millisecond || elapsed > 700*time.Millisecond {
		t.Errorf("expected ~500ms to get 50 tokens, took %v", elapsed)
	}
}

// IMPORTANTE: Rode com go test -race
func TestTokenBucket_RaceConditions(t *testing.T) {
	tb := NewTokenBucket(50, 25, 50*time.Millisecond)
	
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				tb.Allow()
				time.Sleep(5 * time.Millisecond)
			}
		}()
	}
	
	wg.Wait()
}
```

## Worker Pool com Priority Quieu

Você vai implementar um sistema de processamento de jobs com priorização thread-safe. Pensa em Sidekiq, Celery, ou qualquer background job processor que toda startup usa.

Jobs não são todos iguais. Alguns são críticos (email de recuperação de senha), outros podem esperar (gerar relatório mensal). Você precisa garantir que jobs de alta prioridade sejam processados primeiro, mesmo que cheguem depois.

### Por Que Este Challenge Importa

Todo backend de verdade tem um job queue. Quando usuário faz "esqueci minha senha" no GitHub, esse job não pode esperar atrás de 10 mil jobs de "gerar relatório de commits". Priorização é crítica.

Empresas que usam: Sidekiq (Ruby on Rails), Celery (Python/Django), Bull (Node.js), literalmente toda empresa que você está mirando (GitLab, Render, Tailscale). Em entrevistas, "design a task scheduler" ou "implement a job queue with priorities" são perguntas clássicas.

O que você vai construir:

Um worker pool onde múltiplos workers processam jobs de uma priority queue compartilhada. Jobs de maior prioridade (menor número) são processados primeiro. Múltiplos producers adicionando jobs simultaneamente, múltiplos workers consumindo. Tudo thread-safe.
Os Desafios Principais
1. Heap + Concorrência
container/heap do Go não é thread-safe. Se múltiplas goroutines chamarem Push/Pop simultaneamente, a heap corrompe. Você precisa envolver com mutex, mas onde? Lock muito amplo mata performance. Lock insuficiente causa race conditions.
2. Workers Bloqueando Eficientemente
Workers precisam bloquear quando queue está vazia, mas acordar imediatamente quando job chegar ou shutdown acontecer. Não pode ficar em busy loop (desperdiça CPU). Não pode usar polling com sleep (adiciona latência). A solução é sync.Cond.
3. Backpressure
O que acontece quando jobs chegam mais rápido que workers conseguem processar? Queue cresce infinitamente até matar o processo por falta de memória. Você precisa decidir: bloquear producers (backpressure), dropar jobs (perder dados), ou deixar unbounded (risco de OOM).
Background Técnico
Min Heap para Priority Queue
Uma heap é árvore binária onde cada nó é menor (ou maior) que seus filhos. container/heap do Go implementa min-heap: menor elemento sempre no topo.
Para priority queue onde menor priority number = maior prioridade, min-heap é perfeito. Job com priority 0 (Critical) vem antes de priority 3 (Low).
Interface heap.Interface exige:

Len(), Less(), Swap(): para ordenação
Push(x), Pop(): para adicionar/remover

IMPORTANTE: Você sempre chama heap.Push(h, item) (função do package), nunca h.Push(item) diretamente. O package gerencia invariantes da heap.
Complexidade: Push e Pop são O(log n). Muito mais eficiente que manter slice ordenado (O(n) para inserção).
sync.Cond para Coordenação
Você já usou sync.Cond no TCP Chat Server. É exatamente o mesmo pattern aqui.
Workers chamam cond.Wait() quando queue vazia (bloqueia). Quando Enqueue adiciona job, chama cond.Signal() (acorda um worker). Quando Shutdown, chama cond.Broadcast() (acorda todos).
Regra de ouro: cond.Wait() SEMPRE dentro de um loop que verifica condição, SEMPRE com lock do mutex associado.
Backpressure Strategies
Unbounded Queue: Queue cresce infinitamente. Nunca perde jobs, mas pode matar processo por falta de memória.
Bounded + Bloqueio: Queue tem tamanho máximo. Quando cheia, Enqueue bloqueia até ter espaço. Aplica backpressure nos producers.
Bounded + Drop: Queue tem máximo. Quando cheia, Enqueue retorna erro e incrementa contador de dropped. Producer decide o que fazer.
Não tem resposta certa universal. Sidekiq usa bounded + bloqueio. Kafka deixa producer escolher. Para este challenge, vamos fazer bounded + drop (mais simples).
Libs Go Relevantes
Precisa:

container/heap: implementação de heap
sync.Mutex: proteger acesso à heap
sync.Cond: coordenar workers
sync.WaitGroup: esperar workers em shutdown

Trade-offs:
Precisão vs Performance: Lock em toda operação garante correção mas pode criar contenção. Para este challenge, prefira correção (use mutex).
Drain vs Cancel em Shutdown: Quando shutdown acontece, você drena queue (processa jobs restantes) ou cancela tudo? Vamos drenar, mas só jobs já enfileirados.
Como Começar
Passo 1: Implementar JobHeap (15 min)
Comece com a heap isolada, sem concorrência. Implemente os 5 métodos de heap.Interface. Rode teste básico até passar.
Dica: em Less(), menor priority number vem primeiro. Se empate, pode desempatar por ID.
Passo 2: PriorityQueue Thread-Safe (20 min)
Envolva heap com mutex e cond. Campos: heap, mutex, cond, shutdown flag, métricas.
Dequeue() é o mais interessante: loop com cond.Wait() enquanto queue vazia E não está em shutdown. Quando sair do loop, verificar se é porque tem job ou porque shutdown aconteceu.
Passo 3: WorkerPool (15 min)
Com queue funcionando, pool é simples. Cada worker é loop: Dequeue -> Processa -> Repete. Use WaitGroup para coordenar shutdown.
Passo 4: Testar com Race Detector (10 min)
Rode go test -race -v. O teste TestPriorityOrdering é crítico: verifica que jobs são processados por prioridade. Se passar com -race, está correto.

Decisões de Implementação
1. Bounded ou Unbounded?
Comece bounded. É mais seguro e força você a pensar em backpressure. maxSize como parâmetro do construtor.
2. Como Acordar Workers?
Em Enqueue(): depois de Push, chame cond.Signal() (acorda um worker).
Em Shutdown(): depois de setar flag, chame cond.Broadcast() (acorda todos).
3. Métricas?
Adicione campos int para enqueued, processed, dropped. Útil para debugging e produção real.

```
package main

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

/*
TODO - PASSOS PARA IMPLEMENTAR WORKER POOL COM PRIORITY QUEUE

1. IMPLEMENTAR JobHeap COM heap.Interface
   - Len(), Less(), Swap() para ordenação
   - Push() e Pop() para container/heap
   - Menor priority number = maior prioridade (vem primeiro)

2. CRIAR PriorityQueue THREAD-SAFE
   - Envolver JobHeap com mutex
   - Usar sync.Cond para acordar workers quando job chegar
   - Implementar Enqueue (adiciona job, sinaliza workers)
   - Implementar Dequeue (bloqueia se vazio, acorda em shutdown)

3. IMPLEMENTAR WorkerPool
   - Criar N workers em goroutines
   - Cada worker loop: Dequeue -> Processa -> Repete
   - WaitGroup para coordenar shutdown
   
4. GRACEFUL SHUTDOWN
   - Parar de aceitar novos jobs
   - Broadcast para acordar todos os workers bloqueados
   - Esperar workers terminarem jobs atuais
*/
func main() {
	fmt.Println("=== Worker Pool com Priority Queue ===\n")

	processor := func(job *Job) error {
		priorityName := map[int]string{
			PriorityCritical: "CRITICAL",
			PriorityHigh:     "HIGH",
			PriorityNormal:   "NORMAL",
			PriorityLow:      "LOW",
		}
		fmt.Printf("[Worker] Processing job %d [%s]: %s\n",
			job.ID, priorityName[job.Priority], job.Payload)
		time.Sleep(100 * time.Millisecond)
		return nil
	}

	pool := NewWorkerPool(3, 10, processor)
	pool.Start()

	jobs := []*Job{
		{ID: 1, Priority: PriorityNormal, Payload: "Send newsletter"},
		{ID: 2, Priority: PriorityCritical, Payload: "Password reset email"},
		{ID: 3, Priority: PriorityLow, Payload: "Cleanup old logs"},
		{ID: 4, Priority: PriorityHigh, Payload: "Welcome email"},
		{ID: 5, Priority: PriorityCritical, Payload: "Payment confirmation"},
		{ID: 6, Priority: PriorityNormal, Payload: "Weekly report"},
		{ID: 7, Priority: PriorityLow, Payload: "Aggregate metrics"},
		{ID: 8, Priority: PriorityHigh, Payload: "Push notification"},
	}

	fmt.Println("Submitting jobs...")
	for _, job := range jobs {
		if err := pool.Submit(job); err != nil {
			fmt.Printf("Failed to submit job %d: %v\n", job.ID, err)
		}
	}

	time.Sleep(2 * time.Second)

	enq, proc, drop, qs := pool.Stats()
	fmt.Printf("\n=== Stats ===\n")
	fmt.Printf("Enqueued: %d, Processed: %d, Dropped: %d, Queue Size: %d\n", enq, proc, drop, qs)

	fmt.Println("\nShutting down...")
	pool.Shutdown()
	fmt.Println("All workers stopped. Done!")
}

const (
	PriorityCritical = 0 // Password reset, payment confirmation
	PriorityHigh     = 1 // Welcome emails, notifications
	PriorityNormal   = 2 // Newsletter, analytics
	PriorityLow      = 3 // Cleanup tasks, logs aggregation
)

type Job struct {
	ID       int
	Priority int
	Payload  string
}

type JobHeap []*Job

func (h JobHeap) Len() int { return len(h) }

func (h JobHeap) Less(i, j int) bool {
	return h[i].Priority < h[j].Priority
}

func (h JobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *JobHeap) Push(x any) {
	*h = append(*h, x.(*Job))
}

func (h *JobHeap) Pop() any {
	old := *h
	n := len(old)
	job := old[n-1]
	*h = old[:n-1]
	return job
}

type PriorityQueue struct {
	heap           *JobHeap
	maxSize        int
	mu             sync.Mutex
	cond           *sync.Cond
	isShuttingDown bool
	enqueued       int
	processed      int
	dropped        int
}

func NewPriorityQueue(maxSize int) *PriorityQueue {
	pq := &PriorityQueue{
		heap:    &JobHeap{},
		maxSize: maxSize,
	}
	heap.Init(pq.heap)
	pq.cond = sync.NewCond(&pq.mu)
	return pq
}

func (pq *PriorityQueue) Enqueue(job *Job) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.isShuttingDown {
		return fmt.Errorf("shutting down")
	}

	if pq.maxSize > 0 && pq.heap.Len() >= pq.maxSize {
		pq.dropped++
		return fmt.Errorf("queue full")
	}

	heap.Push(pq.heap, job)
	pq.enqueued++
	pq.cond.Signal()
	return nil
}

func (pq *PriorityQueue) Dequeue() (*Job, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for pq.heap.Len() == 0 && !pq.isShuttingDown {
		pq.cond.Wait()
	}

	if pq.isShuttingDown && pq.heap.Len() == 0 {
		return nil, false
	}

	job := heap.Pop(pq.heap).(*Job)
	pq.processed++
	return job, true
}

func (pq *PriorityQueue) Shutdown() {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.isShuttingDown = true
	pq.cond.Broadcast()
}

func (pq *PriorityQueue) Stats() (enqueued, processed, dropped, queueSize int) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return pq.enqueued, pq.processed, pq.dropped, pq.heap.Len()
}

type WorkerPool struct {
	queue       *PriorityQueue
	numWorkers  int
	processFunc func(*Job) error
	wg          sync.WaitGroup
}

func NewWorkerPool(numWorkers, queueSize int, processor func(*Job) error) *WorkerPool {
	return &WorkerPool{
		queue:       NewPriorityQueue(queueSize),
		numWorkers:  numWorkers,
		processFunc: processor,
	}
}

func (wp *WorkerPool) Start() {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	for {
		job, ok := wp.queue.Dequeue()
		if !ok {
			return
		}
		if err := wp.processFunc(job); err != nil {
			fmt.Printf("Worker %d error processing job %d: %v\n", id, job.ID, err)
		}
	}
}

func (wp *WorkerPool) Submit(job *Job) error {
	return wp.queue.Enqueue(job)
}

func (wp *WorkerPool) Shutdown() {
	wp.queue.Shutdown()
	wp.wg.Wait()
}

func (wp *WorkerPool) Stats() (enqueued, processed, dropped, queueSize int) {
	return wp.queue.Stats()
}

```
testes:

```
package main

import (
	"container/heap"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestJobHeapBasic testa operações básicas da heap
func TestJobHeapBasic(t *testing.T) {
	h := &JobHeap{}
	heap.Init(h)

	jobs := []*Job{
		{ID: 1, Priority: PriorityNormal},
		{ID: 2, Priority: PriorityCritical},
		{ID: 3, Priority: PriorityLow},
		{ID: 4, Priority: PriorityHigh},
	}

	// Push jobs
	for _, job := range jobs {
		heap.Push(h, job)
	}

	// Pop deve retornar em ordem de prioridade
	expected := []int{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}
	for i, exp := range expected {
		job := heap.Pop(h).(*Job)
		if job.Priority != exp {
			t.Errorf("Pop %d: expected priority %d, got %d", i, exp, job.Priority)
		}
	}
}

// TestPriorityQueueThreadSafety testa acesso concorrente
func TestPriorityQueueThreadSafety(t *testing.T) {
	pq := NewPriorityQueue(100)

	var wg sync.WaitGroup
	numProducers := 10
	numConsumers := 5
	jobsPerProducer := 20

	// Producers
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < jobsPerProducer; j++ {
				job := &Job{
					ID:       id*jobsPerProducer + j,
					Priority: j % 4, // Variar prioridades
					Payload:  "test",
				}
				pq.Enqueue(job)
			}
		}(i)
	}

	// Consumers
	consumed := int32(0)
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				job, ok := pq.Dequeue()
				if !ok {
					return
				}
				if job != nil {
					atomic.AddInt32(&consumed, 1)
				}
			}
		}()
	}

	// Esperar producers terminarem
	time.Sleep(100 * time.Millisecond)

	// Shutdown
	pq.Shutdown()
	wg.Wait()

	expected := int32(numProducers * jobsPerProducer)
	if consumed != expected {
		t.Errorf("Expected %d jobs consumed, got %d", expected, consumed)
	}
}

// TestBoundedQueue testa comportamento quando queue enche
func TestBoundedQueue(t *testing.T) {
	maxSize := 5
	pq := NewPriorityQueue(maxSize)

	// Encher queue
	for i := 0; i < maxSize; i++ {
		err := pq.Enqueue(&Job{ID: i, Priority: PriorityNormal})
		if err != nil {
			t.Fatalf("Failed to enqueue job %d: %v", i, err)
		}
	}

	// Próximo enqueue deve falhar ou bloquear dependendo da implementação
	err := pq.Enqueue(&Job{ID: 999, Priority: PriorityCritical})
	if err == nil && pq.Len() > maxSize {
		t.Error("Queue exceeded max size without error")
	}
}

// TestWorkerPoolProcessing testa que workers processam jobs
func TestWorkerPoolProcessing(t *testing.T) {
	processed := int32(0)
	processor := func(job *Job) error {
		atomic.AddInt32(&processed, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	pool := NewWorkerPool(3, 20, processor)
	pool.Start()

	// Submit jobs
	numJobs := 15
	for i := 0; i < numJobs; i++ {
		pool.Submit(&Job{
			ID:       i,
			Priority: i % 4,
			Payload:  "test",
		})
	}

	// Esperar processar
	time.Sleep(500 * time.Millisecond)
	pool.Shutdown()

	if processed != int32(numJobs) {
		t.Errorf("Expected %d jobs processed, got %d", numJobs, processed)
	}
}

// TestPriorityOrdering testa que jobs de alta prioridade são processados primeiro
func TestPriorityOrdering(t *testing.T) {
	var mu sync.Mutex
	var processOrder []int

	processor := func(job *Job) error {
		mu.Lock()
		processOrder = append(processOrder, job.ID)
		mu.Unlock()
		time.Sleep(50 * time.Millisecond)
		return nil
	}

	pool := NewWorkerPool(1, 10, processor) // 1 worker para ordem determinística
	pool.Start()

	// Submit em ordem aleatória mas IDs baixos = prioridade alta
	jobs := []*Job{
		{ID: 3, Priority: PriorityLow},
		{ID: 1, Priority: PriorityCritical},
		{ID: 4, Priority: PriorityLow},
		{ID: 2, Priority: PriorityHigh},
	}

	for _, job := range jobs {
		pool.Submit(job)
	}

	time.Sleep(500 * time.Millisecond)
	pool.Shutdown()

	// Verificar que processou em ordem de prioridade
	if len(processOrder) != 4 {
		t.Fatalf("Expected 4 jobs processed, got %d", len(processOrder))
	}

	// ID 1 (Critical) deve vir antes de ID 2 (High)
	// ID 2 (High) deve vir antes de IDs 3,4 (Low)
	if processOrder[0] != 1 || processOrder[1] != 2 {
		t.Errorf("Wrong processing order: %v", processOrder)
	}
}

```