if string.match(tx:title(), "CARD_PAYMENT.XTB") 
 and tx:sourceCurrency() == "USD" then
    local destinationAccount = helpers:getAccountByID(116)

    tx:transactionType(1) -- convert to transfer

    tx:destinationAccountID(destinationAccount.ID)
    tx:destinationCurrency(destinationAccount.Currency)
    tx:destinationAmount(
        math.abs(helpers:convertCurrency(
            tx:sourceCurrency(),
            destinationAccount.Currency,
            tx:sourceAmount()
        ))
    )
end
