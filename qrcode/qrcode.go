package qrcode

import (
	"fmt"
	"os"

	"github.com/mdp/qrterminal"
)

func RenderQRCode(title string, codeString string) {
	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.WHITE,
		WhiteChar: qrterminal.BLACK,
		QuietZone: 1,
	}

	fmt.Printf("\n\n %s:", title)
	qrterminal.GenerateWithConfig(codeString, config)
}
