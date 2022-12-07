package main

import (
	"fmt"
	"strings"

	"github.com/aronkof/keizai/inter"
)

const (
	MERCADO      = "MERCADO"
	TELMA        = "TELMA"
	LUNA         = "LUNA"
	ASSINATURAS  = "ASSINATURAS"
	LUZ          = "LUZ"
	GAS          = "GAS"
	APTO         = "APTO"
	DESCONHECIDO = "DESCONHECIDO"
)

var groups map[string]string = map[string]string{
	"sacolao":                         MERCADO,
	"coco verde":                      MERCADO,
	"atacadao":                        MERCADO,
	"pao de acucar":                   MERCADO,
	"madrid":                          MERCADO,
	"horti":                           MERCADO,
	"hortifruti":                      MERCADO,
	"bolo da madre":                   MERCADO,
	"telma":                           TELMA,
	"petz":                            LUNA,
	"magic animal":                    LUNA,
	"alojamento de animais":           LUNA,
	"spotify":                         ASSINATURAS,
	"netflix":                         ASSINATURAS,
	"eletropaulo":                     LUZ,
	"enel":                            LUZ,
	"condominio":                      APTO,
	"pagamento de titulo - pagamento": APTO,
	"comgas":                          GAS,
}

func CategorizedTxns(txns []inter.Transaction) (map[string]float64, float64) {
	categorized := make(map[string]float64)
	var total float64

	for _, txn := range txns {
		if txn.Type != "D" {
			continue
		}

		total += txn.Value

		wasCategorized := false

		normalizedDesc := strings.ToLower(txn.Description)
		for k, v := range groups {
			if strings.Contains(normalizedDesc, k) {
				categorized[v] += txn.Value
				wasCategorized = true
				break
			}
		}

		if !wasCategorized {
			categorized[DESCONHECIDO] += txn.Value
			fmt.Printf("[WARNING] categoria desconhecida: %v, valor: %.2f \n", txn.Description, txn.Value)
		}
	}

	return categorized, total
}
