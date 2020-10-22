package comgas

import (
	"context"
	"fmt"
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
}

//InvoiceFlow crawls through the enel page
func (flow *Flow) InvoiceFlow() (Invoice, error) {
	for i := range flow.cancel {
		defer flow.cancel[i]()
	}

	err := flow.login()
	if err != nil {
		log.Println(err)
		return Invoice{}, err
	}

	invoice, err := flow.invoiceData()
	if err != nil {
		log.Println(err)
		return Invoice{}, err
	}
	return invoice, nil
}

//InvoiceFlow crawls through the comgas page
func (flow *Flow) invoiceData() (Invoice, error) {
	log.Println("Starting Invoice Flow")

	var (
		dueDate   = ""
		barCode   = ""
		value     = ""
		status    = "unknown"
		outerHTML = ""
	)

	err := chromedp.Run(flow.c,
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

	if err != nil {
		return Invoice{}, err
	}

	if strings.Contains(outerHTML, "A vencer") {
		status = "pending"
	}
	log.Println("Successfully finished Invoice flow")
	return Invoice{
		Value:   strings.Replace(strings.TrimPrefix(value, "R$ "), ",", ".", -1),
		BarCode: strings.Replace(barCode, " ", "", -1),
		DueDate: dueDate,
		Status:  status,
	}, nil
}

func (flow *Flow) login() error {
	log.Println("Starting login flow")
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
		return err
	}
	if !strings.Contains(output, "Bem-vindo") {
		return fmt.Errorf("failed to load login page")
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
		return err
	}
	log.Println("Successfully finished login flow")
	return nil
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
