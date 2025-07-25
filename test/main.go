package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

const totalRequests = 20000
const endpoint = "http://localhost:9999/payments"

type Payment struct {
	CorrelationId string  `json:"correlationId"`
	Amount        float64 `json:"amount"`
}

func main() {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalAmount float64
	var failedCount int

	concurrency := 120
	semaphore := make(chan struct{}, concurrency)

	rand.Seed(time.Now().UnixNano())
	start := time.Now()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		semaphore <- struct{}{}

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }()

			amount := randomAmount()

			payment := Payment{
				CorrelationId: uuid.NewString(),
				Amount:        amount,
			}

			body, err := json.Marshal(payment)
			if err != nil {
				log.Printf("Erro ao gerar JSON: %v", err)
				mu.Lock()
				failedCount++
				totalAmount += amount
				mu.Unlock()
				return
			}

			req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(body))
			if err != nil {
				log.Printf("Erro ao criar request: %v", err)
				mu.Lock()
				failedCount++
				totalAmount += amount
				mu.Unlock()
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Printf("Erro ao enviar request: %v", err)
				mu.Lock()
				failedCount++
				totalAmount += amount
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Status inesperado: %d", resp.StatusCode)
				mu.Lock()
				failedCount++
				mu.Unlock()
			}

			// Soma o valor mesmo que tenha dado erro
			mu.Lock()
			totalAmount += amount
			mu.Unlock()
		}()
	}

	wg.Wait()
	end := time.Now()

	fmt.Println("Todas as requisições foram enviadas às", end.Format("15:04:05"))
	fmt.Printf("Valor total enviado (tentado): R$ %.2f\n", totalAmount)
	fmt.Printf("Total de falhas: %d\n", failedCount)
	fmt.Printf("Tempo total de execução: %s\n", end.Sub(start).Truncate(time.Millisecond))
}

func randomAmount() float64 {
	return float64(rand.Intn(999999)+1) / 100.0 // valores entre 0.01 e 9999.99
}
