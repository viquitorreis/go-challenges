# Worker Pool Challenge (30 min)

## Task: URL Health Checker

Implement a worker pool that checks the "health" of a list of URLs concurrently - but with constraints.
Signature

```
go
type CheckResult struct {
    URL      string
    Healthy  bool
    Err      error
    Duration time.Duration
}

func CheckURLs(urls []string, numWorkers int, timeout time.Duration) []CheckResult
```

**Requirements**

- Spin up numWorkers goroutines processing a job channel
- Each "check" is mocked - don't make real HTTP calls, simulate with:

```go
go  func mockCheck(url string) (bool, error) {
      time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
      if strings.Contains(url, "bad") {
          return false, errors.New("connection refused")
      }
      return true, nil
  }
```

- Apply timeout per URL - if mockCheck takes longer than timeout, result should be Healthy: false, Err: context deadline exceeded
- Collect all results - order doesn't need to match input order, but all URLs must be in the output
- Use sync.WaitGroup + channels (no errgroup, no external libs)

**Twist (a diferença de hoje)**

`mockCheck` não recebe `context.Context` - você precisa aplicar o timeout de fora, usando select com um channel de resultado + time.After (ou context.WithTimeout + goroutine wrapper). Isso é o pattern real de "timeout em código que não suporta context nativamente".

**Test input**

```go
gourls := []string{"good1.com", "bad1.com", "good2.com", "bad2.com", "good3.com"}
results := CheckURLs(urls, 2, 250*time.Millisecond)
// len(results) == 5
```

**O que será avaliado**

- Worker pool correto (N workers, M jobs)
- Pattern de timeout externo (sem context nativo na função mockada)
- Sem deadlocks, sem goroutine leaks
- Collecting results de forma concurrent-safe