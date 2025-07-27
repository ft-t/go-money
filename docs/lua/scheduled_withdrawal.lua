local acc = helpers:getAccountByID(2)

tx:transactionType(3)
tx:sourceAccountID(acc.ID)
tx:sourceAmount(-1)
tx:sourceCurrency(acc.Currency)
