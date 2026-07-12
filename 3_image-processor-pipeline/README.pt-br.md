# Image Processing Pipeline

🇺🇸 [English version](README.md)

**Categoria:** Concorrência / Pipeline
**Tempo estimado:** ~1h30

## O que é

Um pipeline concorrente de 4 estágios que lista arquivos de imagem, carrega do disco, converte pra escala de cinza e salva o resultado, com cada estágio rodando na sua própria goroutine e conectado por channels.

## O que você aprende

- Conectar um **pipeline pattern**: `Generator -> Loader -> Processor -> Saver`, onde cada estágio lê do channel de saída do estágio anterior e escreve no seu próprio.
- Graceful shutdown em cadeia: cada estágio só fecha seu channel de saída depois que o channel de entrada fechou.
- Cancelamento por `context` propagado por todos os estágios, pra que o pipeline inteiro possa ser interrompido no meio da execução.

## O que foi implementado

- `NewPipeline(inputDir, outputDir string) *Pipeline` e `Run(ctx context.Context) error` orquestrando os quatro estágios.
- `generator` varre o diretório de input procurando arquivos `.jpg`/`.png`.
- `loader` decodifica cada arquivo pra uma `image.Image`.
- `processor` converte a imagem carregada pra escala de cinza.
- `saver` escreve a imagem processada no diretório de output.
- Os testes cobrem o fluxo básico, diretório de input vazio, cancelamento de contexto no meio do pipeline, e múltiplas imagens processadas concorrentemente.

## Decisões de design

- Cada estágio é dono de exatamente um channel de saída e é o único escritor nele, então não precisa de mutex em nenhum lugar do pipeline.
- O `context.Context` é passado por todas as assinaturas de função de cada estágio, não só checado uma vez no topo, pra que o cancelamento tenha efeito entre quaisquer dois passos do pipeline.

## Como rodar

```bash
go run .
go test ./...
```
