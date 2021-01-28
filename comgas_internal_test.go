package comgas

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func Test(t *testing.T) {
	f := NewFlow(false)
	user, err := readFile("config_test.json")
	if err != nil {
		t.Logf("readFile err: %v", err)
	}
	f.User = user
	invoice, err := f.InvoiceFlow()
	if err != nil {
		t.Fatalf("f.InvoiceFlow err: %v", err)
	}
	t.Logf("invoice: %#v", invoice)
}

func readFile(filename string) (UserData, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return UserData{}, err
	}
	output := UserData{}
	err = json.Unmarshal(file, &output)
	if err != nil {
		return UserData{}, err
	}
	return output, nil
}
