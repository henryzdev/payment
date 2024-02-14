package payment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"payment/internal/config"
	"payment/internal/model"
	"payment/internal/payment/status"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentService interface {
	CreatePayment(paymentData *model.CreatePayment) error
}

type paymentService struct {
	conn *amqp.Connection
}

func NewPaymentService(conn *amqp.Connection) PaymentService {
	return &paymentService{conn}
}

type PaymentResponse struct {
	QRCode string `json:"qr_code"`
	ID     int    `json:"id"`
}

type CardTokenResponse struct{
	TokenID string `json:"id"`
}


func generateRandomIdempotencyKey() string {
	rand.Seed(time.Now().UnixNano())
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 15)
	for i := 0; i < 15; i++ {
		code[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(code)
}

func (s *paymentService) SendPaidProduct(productID, userID int) error {
	ch, err := s.conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"paid_products",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	message := model.PaidProdutcs{
		ProductID: productID,
		UserID:    userID,
		Status:    "paid",
	}
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        messageJSON,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}


func (s *paymentService) createPaymentPix(paymentData *model.CreatePayment, ch chan bool) (*PaymentResponse, error) {
	payload := map[string]interface{}{
		"transaction_amount": paymentData.Price,
		"payment_method_id": "pix",
		"payer": map[string]interface{}{
			"email": paymentData.PayerEmail,
		},
	}
	expirationTime := time.Now().Add(8 * time.Minute)
	expirationFormatted := expirationTime.Format("2006-01-02T15:04:05.000-07:00")
	payload["date_of_expiration"] = expirationFormatted
	
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", config.MPurl, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.MpToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return nil, err
	}
	if resp.StatusCode == http.StatusCreated {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		if id, ok := result["id"].(float64); ok {
			fmt.Println("PIX gerado, ID :", int(id))
			go status.Status(int(id), ch)
			var qrCode string
			go func(){
				isPaid := <-ch
				if isPaid{
					s.SendPaidProduct(int(id), paymentData.From)
				}else{
					fmt.Println("Pagamento cancelado!")
				}

			}()
			if poi, ok := result["point_of_interaction"].(map[string]interface{}); ok {
				if qr, ok := poi["transaction_data"].(map[string]interface{})["qr_code"]; ok {
					fmt.Println("QR Code:", qr)
					qrCode = qr.(string)
				}
			}

			paymentResponse := &PaymentResponse{
				QRCode: qrCode,
				ID:     int(id),
			}
			return paymentResponse, nil
		}
		fmt.Println("Campo 'id' não encontrado no JSON.")
	} else {
		return nil, errors.New("status Code != 200")
	}
	return nil, nil
}

func GetPixInfo(pixID int) (*PaymentResponse, error) {
	url := fmt.Sprintf("%s/%d", config.MPurl, pixID)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Erro ao criar a requisição:", err)
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.MpToken))

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Erro ao enviar a requisição:", err)
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Erro ao ler corpo da resposta:", err)
		return nil, err
	}

	if response.StatusCode == http.StatusCreated {
		var result PaymentResponse
		if err := json.Unmarshal(body, &result); err != nil {
			fmt.Println("Erro ao analisar o JSON da resposta:", err)
			return nil, err
		}

		return &result, nil
	}

	fmt.Println("Erro na requisição:", response.Status)
	fmt.Println("Corpo da resposta:", string(body))
	return nil, errors.New(fmt.Sprintf("Erro na requisição: %s", response.Status))
}

func (ps *paymentService) GetQRCode(paymentData *model.CreatePayment) (string, error) {
	ch := make(chan bool)
	paymentResponse, err := ps.createPaymentPix(paymentData, ch) // Adicionado o canal como argumento
	if err != nil {
		return "", err
	}

	if paymentResponse != nil {
		return paymentResponse.QRCode, nil
	}

	return "", errors.New("Resposta de pagamento vazia")
}

func detectCardType(cardNumber string) string {
    if len(cardNumber) < 4 {
        return "Desconhecido"
    }
    firstSixDigits := cardNumber[:6]
    switch {
    case strings.HasPrefix(firstSixDigits, "4"):
        return "visa"
    case strings.HasPrefix(firstSixDigits, "5"):
        return "mastercard"
    case strings.HasPrefix(firstSixDigits, "34"), strings.HasPrefix(firstSixDigits, "37"):
        return "American Express"
    case strings.HasPrefix(firstSixDigits, "6"):
        return "Discover"
    case strings.HasPrefix(firstSixDigits, "30"), strings.HasPrefix(firstSixDigits, "36"), strings.HasPrefix(firstSixDigits, "38"):
        return "Diners Club"
    case strings.HasPrefix(firstSixDigits, "35"):
        return "JCB"
    default:
        return "Desconhecido"
    }
}


func useCreditCard(paymentData *model.CreatePayment, flagCard string, cardToken string) error {
    payload := map[string]interface{}{
       "transaction_amount": paymentData.Price,
	   "token": fmt.Sprintf("%s", cardToken),
	   "description": "Cobrança ecommerce andreia",
	   "installments": 1,
	   "payment_method_id": flagCard,
	   "payer": map[string]interface{}{
		"email": paymentData.PayerEmail,
	   },
    }
    paymentJSON, err := json.Marshal(payload)
    if err != nil {
        return err
    }
	fmt.Println(string(paymentJSON))
    req, err := http.NewRequest("POST", config.MPurl, bytes.NewReader(paymentJSON))
    if err != nil {
        return err
		
    }
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.MpToken))
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return err
    }
    fmt.Println(string(body))
    return nil
}

func createPaymentCreditCard(paymentData *model.CreatePayment) error  {
	var PaymentCreditCardData model.PaymentCreditCard
	PaymentCreditCardData.CardNumber = paymentData.CardNumber
	PaymentCreditCardData.ExpirationMonth = paymentData.ExpirationMonth
	PaymentCreditCardData.ExpirationYear = paymentData.ExpirationYear
	PaymentCreditCardData.Cvv = paymentData.SecurityCode
	PaymentCreditCardData.CardHolder.Name = paymentData.Name
	PaymentCreditCardData.CardHolder.Identification.Type = paymentData.CPF
	PaymentCreditCardData.CardHolder.Identification.Number = paymentData.CPF
	
	payload := map[string]interface{}{
		"card_number": PaymentCreditCardData.CardNumber,
		"expiration_month": PaymentCreditCardData.ExpirationMonth,
		"expiration_year": PaymentCreditCardData.ExpirationYear,
		"security_code": PaymentCreditCardData.Cvv,
		"card_holder": map[string]interface{}{
			"name": PaymentCreditCardData.CardHolder.Name,
			"identification": map[string]interface{}{
				"type": PaymentCreditCardData.CardHolder.Identification.Type,
				"number": PaymentCreditCardData.CardHolder.Identification.Number,
			},
		},
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil{
		return err
	}
	req, err := http.NewRequest("POST", config.MpUrlCCToken, strings.NewReader(string(jsonPayload)))
	if err != nil {
		return err
	}
	idempotency := generateRandomIdempotencyKey()
	req.Header.Set("X-Idempotency-Key", idempotency)
	req.Header.Set("Authorization",	fmt.Sprintf("Bearer %s", config.MpToken))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var token CardTokenResponse
	err = json.Unmarshal(body, &token)
	if err != nil {
		return err
	}
	flag := detectCardType(paymentData.CardNumber)
	fmt.Println(flag, token.TokenID)
	err = useCreditCard(paymentData, flag, token.TokenID)
	if err != nil {
		return err
	}
	return nil
}

func (ps *paymentService) CreatePayment(paymentData *model.CreatePayment) error {
	switch paymentData.PaymentMethod {
	case "credit_card":
		fmt.Println("PAGAMENTO CREDIT CARD")
		err := createPaymentCreditCard(paymentData)
		if err != nil{
			fmt.Println(err)
		}
	case "pix":
		fmt.Println("PAGAMENTO PIX")
		ch := make(chan bool)
		qr_code, err := ps.createPaymentPix(paymentData, ch)
		if err != nil{
			log.Fatal(err)
		}
		fmt.Println(qr_code)
		go func() {
			paymentApproved := <-ch
			if paymentApproved {
				fmt.Println("Pagamento aprovado!")
			} else {
				fmt.Println("Pagamento cancelado!")
			}
		}()
		default:
			return errors.New("Método de pagamento inválido")
	}
	return nil
}