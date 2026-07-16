# Image Processing Pipeline

🇧🇷 [Versão em Português](README.pt-br.md)

**Category:** Concurrency / Pipeline
**Estimated time:** ~1.5 hours

## What it is

A 4-stage concurrent pipeline that lists image files, loads them from disk, converts them to grayscale, and saves the result, with each stage running in its own goroutine and connected by channels.

## What you'll learn

- Wiring a **pipeline pattern**: `Generator -> Loader -> Processor -> Saver`, where each stage reads from the previous stage's output channel and writes to its own.
- Graceful shutdown across a chain: each stage closes its own output channel only after its input channel closes.
- `context` cancellation propagated through every stage so the whole pipeline can be stopped mid-flight.

## What's implemented

- `NewPipeline(inputDir, outputDir string) *Pipeline` and `Run(ctx context.Context) error` orchestrating the four stages.
- `generator` walks the input directory for `.jpg`/`.png` files.
- `loader` decodes each file into an `image.Image`.
- `processor` converts the loaded image to grayscale.
- `saver` writes the processed image to the output directory.
- Tests cover the basic flow, an empty input directory, context cancellation mid-pipeline, and multiple images processed concurrently.

## Design decisions

- Each stage owns exactly one output channel and is the only writer to it, so no mutex is needed anywhere in the pipeline.
- `context.Context` is threaded through every stage function signature, not just checked once at the top, so cancellation takes effect between any two pipeline steps.

## How to run

```bash
go run .
go test ./...
```
