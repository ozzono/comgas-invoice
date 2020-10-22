# COMGAS query flow
##### _Read it in english [here](https://github.com/ozzono/comgas_invoice/blob/master/README.md)._
Esse pacote usa o [chromedp](https://github.com/chromedp/chromedp) para navegar pelo página de usuário da [comgas](https://virtual.comgas.com.br/#/comgasvirtual/historicoFaturas).

Esse pacote exige:
- cpf
- código do usuário
- nome

E retorna:
- vencimento _(dueDate)_
- status _(status)_
- código de barras_(barCode)_
- valor _(value)_