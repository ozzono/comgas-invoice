# COMGAS query flow
##### _Leia em português [aqui](https://github.com/ozzono/comgas_invoice/blob/master/README_pt.md)._
This package uses [chromedp](https://github.com/chromedp/chromedp) to navigate through the [comgas](https://virtual.comgas.com.br/#/comgasvirtual/historicoFaturas) user page.

It requires:
- CPF
- userCode
- name

And returns:
- dueDate
- status
- barCode
- value