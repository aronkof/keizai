package inter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	INTER_LOGIN_URL          = "https://contadigital.bancointer.com.br/"
	ACCOUNT_TRANSACTIONS_URL = "https://cd.web.bancointer.com.br/ib-pfj/extrato/v3"

	TOKEN_REQUEST_SUBSTRING = "qrcode/token"
	CHECK_REQUEST_SUBSTRING = "qrcode/check"

	CHECK_REQUEST_SUCCESS = "SUCCESS"

	DEFAULT_DATE_FORMAT = "02-01-2006"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type CheckResponse struct {
	Status    string `json:"status"`
	LoginData string `json:"dadosLogin"`
}

type BearerToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type InterToken struct {
	BearerToken BearerToken `json:"bearerToken"`
}

type InterClient struct {
	bearerToken string
	httpClient  *http.Client
}

type Transaction struct {
	Date        string  `json:"dataLancamento"`
	Description string  `json:"descricao"`
	Type        string  `json:"tipo"`
	Code        string  `json:"codHist"`
	Currency    string  `json:"moeda"`
	Value       float64 `json:"valor"`
	Balance     float64 `json:"saldo"`
}

func DecodeAccessToken(body []byte) (InterToken, error) {
	b64 := make([]byte, base64.StdEncoding.DecodedLen(len(body)))
	n, err := base64.StdEncoding.Decode(b64, body)
	if err != nil {
		return InterToken{}, err
	}

	decoded := b64[:n]

	var interToken InterToken
	err = json.Unmarshal(decoded, &interToken)
	if err != nil {
		log.Fatal("could not unmarshal inter token")
	}

	return interToken, nil
}

func NewInterClient(bearer string) *InterClient {
	httpClient := http.Client{}
	return &InterClient{bearer, &httpClient}
}

func (ic *InterClient) applyDefaultHeaders(req *http.Request) {
	req.Header.Add("accept", "*/*")
	req.Header.Add("accept-encoding", "gzip, deflate, br")
	req.Header.Add("accept-language", "en-US,en;q=0.9")
	req.Header.Add("authorization", fmt.Sprintf("bearer %s", ic.bearerToken))
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://contadigital.bancointer.com.br")
	req.Header.Add("x-inter-organization", "IBPF")
	req.Header.Add("x-origin-device", "WEB")
}

func (ic *InterClient) GetTransactions() []Transaction {
	now := time.Now()
	initialDate := now.AddDate(0, 0, -30).Format(DEFAULT_DATE_FORMAT)
	endDate := now.Format(DEFAULT_DATE_FORMAT)
	requestUrl := ACCOUNT_TRANSACTIONS_URL + fmt.Sprintf("?data-inicio=%s&data-fim=%s", initialDate, endDate)

	fmt.Printf("REQUESTING TRANSACTIONS FROM %s TO %s \n", initialDate, endDate)
	fmt.Println("THE REQUEST URL IS: " + requestUrl)

	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		log.Fatal("error creating GetTransactions request")
	}

	ic.applyDefaultHeaders(req)

	res, err := ic.httpClient.Do(req)
	if err != nil {
		log.Fatal(fmt.Errorf("%w, error performing GetTransactions request", err))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(fmt.Errorf("%w, error parsing GetTransactions body", err))
	}

	var transactions []Transaction
	err = json.Unmarshal(body, &transactions)
	if err != nil {
		log.Fatal(fmt.Errorf("%w, error unmarshaling GetTransactions body", err))
	}

	return transactions
}

func GetLocalCurrentTransactions() []Transaction {
	localTxnsJSON, err := os.Open("./data/current-transactions.json")
	if err != nil {
		log.Fatal("could not open local transactions json file")
	}
	defer localTxnsJSON.Close()

	var transactions []Transaction

	jsonData, err := ioutil.ReadAll(localTxnsJSON)
	if err != nil {
		log.Fatal("could not read from local transactions json")
	}

	err = json.Unmarshal(jsonData, &transactions)
	if err != nil {
		log.Fatal(fmt.Errorf("%w, error unmarshaling local transactions json data", err))
	}

	return transactions
}
