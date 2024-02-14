package model

type CreatePayment struct {
	ID            int     `json:"id"`
	CPF           string  `json:"cpf"`
	Name          string  `json:"name"`
	CardNumber    string  `json:"card_number"`
	ExpirationMonth    string  `json:"expiration_month"`
	ExpirationYear     string  `json:"expiration_year"`
	SecurityCode       string  `json:"security_code"`
	PayerEmail    string  `json:"PayerEmail"`
	From          int     `json:"from"`
	Price         float32 `json:"price"`
	ProductID     int     `json:"productID"`
	PaymentMethod string  `json:"paymentMethod"`
}

type PaymentCreditCard struct {
	CardNumber 		string `json:"credit_card"`
	ExpirationMonth string `json:"expiration_month"`
	ExpirationYear  string `json:"expiration_year"`
	Cvv 		    string `json:"security_code"`
	CardHolder struct{
		Name string `json:"name"`
		Identification struct{
			Type   string `json:"type"`
			Number string `json:"number"`
		}
	}
}

type PaidProdutcs struct {
	ProductID int `json:"product_id"`
	UserID 	  int `json:"user_id"`
	Status    string `json:"status"`
}