package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/knq/chromedp"
	"github.com/knq/chromedp/kb"
)

var (
	configPath string
)

// Flow contains and the data and methods needed to crawl through the enel webpage
type Flow struct {
	c       context.Context
	User    UserData
	Invoice Invoice
	cancel  []context.CancelFunc
}

//Invoice has all the invoice data needed for payment
type Invoice struct {
	DueDate string
	Value   string
	BarCode string
	Status  string
}

//UserData has all the needed data to login
type UserData struct {
	CPF  string `json:"cpf"`
	Code string `json:"code"`
	// Name  string `json:"name"`
}

func init() {
	flag.StringVar(&configPath, "user-data", "config.json", "Sets the path for the user data JSON file")
}

func main() {
	flag.Parse()
	f := NewFlow(false)
	for i := range f.cancel {
		defer f.cancel[i]()
	}
	user, err := setConfig()
	if err != nil {
		log.Printf("setConfig: %v", err)
	}
	f.User = user
	invoice, err := f.invoiceFlow()
	if err != nil {
		log.Printf("invoiceFlow: %v", err)
	}
	log.Printf("invoice %#v", invoice)

}

func (flow *Flow) invoiceFlow() (Invoice, error) {
	output := ""
	err := chromedp.Run(flow.c,
		chromedp.Navigate("https://virtual.comgas.com.br/#/comgasvirtual/historicoFaturas"),
		chromedp.WaitVisible("div.dados-login"),
		chromedp.Text(
			`document.querySelector("#loginModal > form > div > div.header-geral > span.label-principal.ng-scope")`,
			&output,
			chromedp.ByJSPath,
		),
		chromedp.Sleep(5),
	)

	if err != nil {
		return Invoice{}, err
	}
	if !strings.Contains(output, "Bem-vindo") {
		return Invoice{}, fmt.Errorf("failed to load login page")
	}

	err = chromedp.Run(flow.c,
		chromedp.Click(`document.querySelector("#cpf")`, chromedp.NodeVisible, chromedp.ByJSPath),
		chromedp.SendKeys(`document.querySelector("#cpf")`, kb.End+flow.User.CPF, chromedp.ByJSPath),

		chromedp.Click(`document.querySelector("#modalTxtCodigoUsuario")`, chromedp.NodeVisible, chromedp.ByJSPath),
		chromedp.SendKeys(`document.querySelector("#modalTxtCodigoUsuario")`, kb.End+flow.User.Code, chromedp.ByJSPath),

		chromedp.Click(`document.querySelector("#btnEnviar > span")`, chromedp.NodeVisible, chromedp.ByJSPath),
		chromedp.Sleep(5),
	)
	if err != nil {
		return Invoice{}, err
	}

	var (
		dueDate   = ""
		barCode   = ""
		value     = ""
		status    = "unknown"
		outerHTML = ""
	)

	err = chromedp.Run(flow.c,
		chromedp.WaitVisible("i.icon-feather-log-out"),

		chromedp.Text(
			`document.querySelector("#mainContent > div > div > div.row.hidden-xs > div > div > table > tbody > tr:nth-child(1) > td:nth-child(2)")`,
			&dueDate,
			chromedp.ByJSPath,
		),

		chromedp.Text(
			`document.querySelector("#mainContent > div > div > div.row.hidden-xs > div > div > table > tbody > tr:nth-child(1) > td:nth-child(3)")`,
			&value,
			chromedp.ByJSPath,
		),

		chromedp.OuterHTML(
			`document.querySelector("#mainContent > div > div > div.row.hidden-xs > div > div > table > tbody > tr:nth-child(1) > td:nth-child(1) > i")`,
			&outerHTML,
			chromedp.ByJSPath,
		),

		chromedp.Click(
			`document.querySelector("#mainContent > div > div > div.row.hidden-xs > div > div > table > tbody > tr:nth-child(1) > td.ng-scope > span")`,
			chromedp.NodeVisible,
			chromedp.ByJSPath,
		),
		chromedp.WaitVisible("body > div.modal.fade.ng-isolate-scope.in > div > div > div.modal-body.ng-scope > div > div > div.panel.ng-scope > div > p:nth-child(1) > strong"),

		chromedp.Text(
			`document.querySelector("body > div.modal.fade.ng-isolate-scope.in > div > div > div.modal-body.ng-scope > div > div > div.panel.ng-scope > div > p:nth-child(2) > span")`,
			&barCode,
			chromedp.ByJSPath,
		),
	)
	if strings.Contains(outerHTML, "A vencer") {
		status = "pending"
	}
	return Invoice{
		Value:   strings.Replace(strings.TrimPrefix(value, "R$ "), ",", ".", -1),
		BarCode: strings.Replace(barCode, " ", "", -1),
		DueDate: dueDate,
		Status:  status,
	}, nil
}

//NewFlow creates a flow with context besides user and invoice data
func NewFlow(headless bool) Flow {
	ctx, cancel := setContext(headless)
	return Flow{c: ctx, cancel: cancel}
}

func setContext(headless bool) (context.Context, []context.CancelFunc) {
	outputFunc := []context.CancelFunc{}
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		// Set the headless flag to false to display the browser window
		chromedp.Flag("headless", headless),
	)
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	outputFunc = append(outputFunc, cancel)
	ctx, cancel = chromedp.NewContext(ctx)
	outputFunc = append(outputFunc, cancel)
	return ctx, outputFunc
}

func setConfig() (UserData, error) {
	if len(configPath) == 0 {
		return UserData{}, fmt.Errorf("invalid path; cannot be empty")
	}
	jsonFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return UserData{}, err
	}
	config := UserData{}
	err = json.Unmarshal(jsonFile, &config)
	return config, err
}