package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type PaymentService struct {
	APIURL string
	APIKey string
}

func NewPaymentService(apiUrl, apiKey string) *PaymentService {
	return &PaymentService{
		APIURL: apiUrl,
		APIKey: apiKey,
	}
}

func (ps *PaymentService) CreatePayment(ordeID string, amount float64, currency, description, returnURL string) (string, error) {
	requestBody := map[string]interface{}{
		"amount": map[string]interface{}{
			"value":    amount,
			"currency": currency,
		},
		"capture":     true,
		"description": description,
		"confirmation": map[string]string{
			"type":       "redirect",
			"return_url": returnURL,
		},
		"metadata": map[string]string{
			"order_id": ordeID,
		},
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", ps.APIURL+"/v3/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Basic"+ps.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", errors.New("Failed to create payment")
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	confirmation, ok := response["confirmation"].(map[string]interface{})
	if !ok {
		return "", errors.New("invalid resonse format")
	}

	return confirmation["confirmation_url"].(string), nil

}

func (ps *PaymentService) GetPaymentStatus(paymentID string) (string, error) {
	req, err := http.NewRequest("GET", ps.APIURL+"/v3/payments"+paymentID, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Basic "+ps.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get payment status")
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	status, ok := response["status"].(string)
	if !ok {
		return "", errors.New("invalid responce format")
	}

	return status, nil
}
