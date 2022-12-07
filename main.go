package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/aronkof/keizai/browserbot"
	"github.com/aronkof/keizai/inter"
)

func main() {
	readLocalTxns := flag.Bool("local", false, "read from current-transactions.json local file")
	flag.Parse()

	var txns []inter.Transaction

	if *readLocalTxns {
		txns = getLocalTxns()
		printKeizaiSum(txns)
		return
	}

	txns = getInterTxns()
	printKeizaiSum(txns)
}

func printKeizaiSum(txns []inter.Transaction) {
	categorizedTxns, total := CategorizedTxns(txns)

	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
	for k, v := range categorizedTxns {
		fmt.Fprintf(w, "CATEGORIA %v: \t\t%.2f \n", k, v)
	}

	fmt.Fprintf(w, "\ntotal em R$ do mês: \t\t\t\t\t\t%.2f \n", total)

	w.Flush()
}

func getLocalTxns() []inter.Transaction {
	return inter.GetLocalCurrentTransactions()
}

func getInterTxns() []inter.Transaction {
	interAuthToken := browserbot.GetInterAuthToken()
	interClient := inter.NewInterClient(interAuthToken)
	return interClient.GetTransactions()
}
