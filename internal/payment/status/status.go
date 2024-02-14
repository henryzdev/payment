package status

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"payment/internal/config"
	"time"
)

type PaymentStatus struct {
	Status string `json:"status"`
}

func checkPaymentStatus(paymentID int, ch chan<- bool) {
	url := fmt.Sprintf("%s/%v", config.MPurl, paymentID)
	for {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Erro ao criar requisição:", err)
			ch <- false
			return
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.MpToken ))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Erro ao enviar requisição:", err)
			ch <- false
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Erro ao ler resposta:", err)
			ch <- false
			return
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Println("Erro na resposta da API:", url, resp.Status)
			ch <- false
			return
		}

		var paymentStatus PaymentStatus
		if err := json.Unmarshal(body, &paymentStatus); err != nil {
			fmt.Println("Erro ao decodificar JSON:", err)
			ch <- false
			return
		}

		fmt.Printf("Pagamento de %v: %v\n", paymentID, paymentStatus.Status)

		if paymentStatus.Status == "approved" {
			fmt.Println("Pagamento aprovado!")
			ch <- true
			return
		}else if paymentStatus.Status == "cancelled"{
			fmt.Println("Pagamento cancelado!")
			ch <- false
			return
		}

		time.Sleep(2 * time.Second)
	}
}

func Status(ID int, ch chan<- bool) {
	go checkPaymentStatus(ID, ch)
	
}