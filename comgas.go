package comgas

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/knq/chromedp/kb"
	"github.com/ozzono/normalize"
)

var (
	configPath string
)

// Flow contains and the data and methods needed to crawl through the enel webpage
type Flow struct {
	c       context.Context
	User    UserData
	Invoice Invoice
	cancel  func()
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
	Name string `json:"name"`
}

//InvoiceFlow crawls through the enel page
func (flow *Flow) InvoiceFlow() (Invoice, error) {
	defer flow.cancel()

	if err := flow.checkUserData(); err != nil {
		return Invoice{}, fmt.Errorf("")
	}

	err := flow.login()
	if err != nil {
		log.Println(err)
		return Invoice{}, err
	}

	invoice, err := flow.invoiceFlow()
	if err != nil {
		log.Println(err)
		return Invoice{}, err
	}
	return invoice, nil
}

//InvoiceFlow crawls through the comgas page
func (flow *Flow) invoiceFlow() (Invoice, error) {
	log.Println("starting Invoice Flow")

	output := ""

	err := chromedp.Run(flow.c,
		chromedp.WaitVisible("i.icon-feather-log-out"),
		chromedp.Text(
			`document.querySelector("body > div.container.loggedin-container > div.row.loggedin-row > div.flex-container > div:nth-child(1) > div:nth-child(2) > div > section > a > span")`,
			&output,
			chromedp.ByJSPath,
		),
	)
	if err != nil {
		return Invoice{}, err
	}
	if !strings.Contains(strings.ToLower(output), "sair") {
		return Invoice{}, fmt.Errorf("unable to continue flow; the page has changed")
	}

	err = chromedp.Run(flow.c,
		chromedp.OuterHTML(
			`document.querySelector("#mainContent > app-root > div > div > div > div.row.hidden-xs > div > div > table > tbody > tr:nth-child(1) > td:nth-child(1) > i")`,
			&output,
			chromedp.ByJSPath,
		),
	)
	if err != nil {
		return Invoice{}, err
	}
	if !strings.Contains(strings.ToLower(output), "a vencer") {
		return Invoice{}, fmt.Errorf("no pending invoices")
	}

	invoice := Invoice{Status: "pending"}

	err = chromedp.Run(flow.c,
		chromedp.Text(
			`document.querySelector("#mainContent > app-root > div > div > div > div.row.hidden-xs > div > div > table > tbody > tr:nth-child(1) > td:nth-child(2)")`,
			&invoice.DueDate,
			chromedp.ByJSPath,
		),
		chromedp.Text(
			`document.querySelector("#mainContent > app-root > div > div > div > div.row.hidden-xs > div > div > table > tbody > tr:nth-child(1) > td:nth-child(3)")`,
			&invoice.Value,
			chromedp.ByJSPath,
		),
	)

	if err != nil {
		return Invoice{}, err
	}
	if len(invoice.Value) == 0 {
		return Invoice{}, fmt.Errorf("failed to fetch invoice value")
	}
	if len(invoice.DueDate) == 0 {
		return Invoice{}, fmt.Errorf("failed to fetch invoice duedate")
	}

	err = chromedp.Run(flow.c,
		chromedp.Click(`document.querySelector("#mainContent > app-root > div > div > div > div.row.hidden-xs > div > div > table > tbody > tr:nth-child(1) > td.ng-scope > span")`,
			chromedp.NodeVisible,
			chromedp.ByJSPath,
		),
		chromedp.WaitVisible("i.icon-fechar"),
		chromedp.Text(
			`document.querySelector("body > div.modal.fade.ng-isolate-scope.in > div > div > div.modal-body.ng-scope > div > div > div.panel.ng-scope > div > p:nth-child(1) > strong")`,
			&output,
			chromedp.ByJSPath,
		),
	)
	if err != nil {
		return Invoice{}, err
	}
	if !strings.Contains(strings.ToLower(normalize.Norm(output)), "numero do codigo de barra") {
		return Invoice{}, fmt.Errorf("unable to continue flow; the page has changed")
	}

	err = chromedp.Run(flow.c,
		chromedp.Text(
			`document.querySelector("body > div.modal.fade.ng-isolate-scope.in > div > div > div.modal-body.ng-scope > div > div > div.panel.ng-scope > div > p:nth-child(2) > span")`,
			&invoice.BarCode,
			chromedp.ByJSPath,
		),
	)
	if err != nil {
		return Invoice{}, err
	}
	if len(invoice.BarCode) == 0 {
		return Invoice{}, fmt.Errorf("failed to fetch invoice barcode")
	}

	log.Println("Successfully finished Invoice flow")
	invoice.Value = strings.TrimPrefix(invoice.Value, "R$ ")
	invoice.BarCode = strings.Replace(invoice.BarCode, " ", "", -1)
	return invoice, nil
}

func (flow *Flow) login() error {
	log.Println("starting login flow")
	output := ""
	err := chromedp.Run(flow.c,
		chromedp.Navigate("https://virtual.comgas.com.br/#/comgasvirtual/historicoFaturas"),
		chromedp.WaitVisible("div.dados-login"),
		chromedp.Text(
			`document.querySelector("#loginModal > form > div > div.header-geral > span.label-principal.ng-scope")`,
			&output,
			chromedp.ByJSPath,
		),
		chromedp.Sleep(5*time.Second),
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
		chromedp.WaitVisible("h4.ng-binding"),
		chromedp.Text(
			`document.querySelector("body > div.container.loggedin-container > div.row.loggedin-row > div.flex-container > div:nth-child(1) > div:nth-child(2) > div > section > a > span")`,
			&output,
			chromedp.ByJSPath,
		),
	)
	if err != nil {
		return err
	}
	if len(output) == 0 {
		return fmt.Errorf("unable to continue flow; the page has changed")
	}
	if !strings.Contains(strings.ToLower(normalize.Norm(output)), "sair") {
		return fmt.Errorf("failed to log in; user name does no match: %s!=%s", output, flow.User.Name)
	}
	log.Println("successfully finished login flow")
	return nil
}

func (flow *Flow) checkUserData() error {
	if len(flow.User.CPF) == 0 {
		return fmt.Errorf("invalid flow.User.CPF; cannot be empty")
	}
	if len(flow.User.Code) == 0 {
		return fmt.Errorf("invalid flow.User.Code; cannot be empty")
	}
	if len(flow.User.Name) == 0 {
		return fmt.Errorf("invalid flow.User.Name; cannot be empty")
	}
	return nil
}

//NewFlow creates a flow with context besides user and invoice data
func NewFlow(headless bool) Flow {
	ctx, cancel := setContext(headless)
	return Flow{c: ctx, cancel: cancel}
}

func setContext(headless bool) (context.Context, func()) {
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
	return ctx, func() {
		for i := range outputFunc {
			outputFunc[i]()
		}
	}
}
