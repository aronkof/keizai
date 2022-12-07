package browserbot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aronkof/keizai/inter"
	"github.com/aronkof/keizai/qrcode"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

const CAPTURE_AUTH_TIMEOUT = time.Second * 60

func GetInterAuthToken() string {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	tokenCh := make(chan string)
	defer close(tokenCh)

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if ev, ok := ev.(*network.EventResponseReceived); ok {
			if ev.Type != "XHR" {
				return
			}

			if strings.Contains(ev.Response.URL, inter.TOKEN_REQUEST_SUBSTRING) {
				go renderAccessQRCode(ctx, ev)
			}

			if strings.Contains(ev.Response.URL, inter.CHECK_REQUEST_SUBSTRING) {
				go extractBearerToken(ctx, ev, tokenCh)
			}
		}
	})

	err := chromedp.Run(ctx, interAuth())
	if err != nil {
		log.Fatal(err)
	}

	accessToken := <-tokenCh

	return accessToken
}

// This ChromeDP tasks will:
// 1. Navigate to the Inter web internet banking home;
// 2. Wait for a button to be rendered and click on it to request the 2FA QRCode
// 3. Hang for 1 minute until: waiting for the app to read/check the QR code and proceed with the authentication
// 4. There will be a thread running on the background to fetch the auth token
// 5. The context will then be canceled
func interAuth() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(inter.INTER_LOGIN_URL),
		chromedp.WaitVisible(`//button`),
		chromedp.Click(`//button`),
		chromedp.Sleep(CAPTURE_AUTH_TIMEOUT),
	}
}

func renderAccessQRCode(ctx context.Context, ev *network.EventResponseReceived) {
	c := chromedp.FromContext(ctx)
	rbp := network.GetResponseBody(ev.RequestID)
	body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		log.Fatal(fmt.Errorf("%w, could not get response body from token request", err))
	}

	var tokenResponse inter.TokenResponse
	err = json.Unmarshal(body, &tokenResponse)
	if err != nil {
		log.Fatal("could not parse response body from token request")
	}

	qrcode.RenderQRCode("Inter 2FA QR CODE:", tokenResponse.Token)
}

func extractBearerToken(ctx context.Context, ev *network.EventResponseReceived, tokenCh chan string) {
	c := chromedp.FromContext(ctx)
	rbp := network.GetResponseBody(ev.RequestID)
	body, err := rbp.Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		log.Fatal("could not get response body from `check` request")
	}

	var checkResponse inter.CheckResponse

	err = json.Unmarshal(body, &checkResponse)
	if err != nil {
		log.Fatal("could not parse response body from `check` request")
	}

	if checkResponse.LoginData == "" {
		return
	}

	interToken, err := inter.DecodeAccessToken([]byte(checkResponse.LoginData))
	if err != nil {
		log.Fatal(fmt.Errorf("%w, error decoding access token", err))
	}

	tokenCh <- interToken.BearerToken.AccessToken
}
