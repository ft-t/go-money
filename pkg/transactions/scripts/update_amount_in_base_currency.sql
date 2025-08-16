with upd as (select t.id,
                    case
                        when t.source_currency = @baseCurrency then
                            t.source_amount
                        when t.transaction_type = 3 and t.fx_source_currency = @baseCurrency and
                             t.fx_source_amount is not null then
                            t.fx_source_amount
                        when t.source_amount is not null and t.destination_amount is not null and
                             t.destination_currency =
                             @baseCurrency -- if other side already in base currency, no need to convert
                            then
                            t.destination_amount
                        when t.source_currency != @baseCurrency then
                            round(t.source_amount / sourceCurrency.rate, sourceCurrency.decimal_places)
                        else t.source_amount
                        end as sourceInBase,
                    case
                        when t.destination_currency = @baseCurrency then
                            t.destination_amount
                        when t.transaction_type = 3 and t.fx_source_currency = @baseCurrency and
                             t.fx_source_amount is not null then
                            t.fx_source_amount
                        when t.destination_amount is not null and t.source_amount is not null and
                             t.source_currency = @baseCurrency
                            then -- if other side already in base currency, no need to convert
                            t.source_amount
                        when t.destination_currency != @baseCurrency then
                            round(t.destination_amount / destinationCurrency.rate, destinationCurrency.decimal_places)
                        else t.destination_amount
                        end as destinationInBase
             from transactions t
                      left join currencies sourceCurrency on sourceCurrency.id = t.source_currency
                      left join currencies destinationCurrency on destinationCurrency.id = t.destination_currency
             where (@specificTxIDs)::bigint[] IS NULL
                OR t.id = ANY ((@specificTxIDs)::bigint[]))
UPDATE transactions
SET destination_amount_in_base_currency = case
                                              when transaction_type = 1
                                                  then abs(upd.sourceInBase) -- for transaction lets use same value for both operations but ensure its > 0
                                              when transaction_type = 3 and destination_account_id = 0 -- if withdrawal and no destination account
                                                  then null
                                              when transaction_type = 3 and destination_account_id != 0
                                                  then abs(upd.sourceInBase) -- left side should be same as right side
                                              else upd.destinationInBase end,
    source_amount_in_base_currency      = case
                                              when transaction_type = 1
                                                  then abs(upd.sourceInBase) * -1 -- ensure that source < 0 in transfers 
                                              when transaction_type = 3
                                                  then -abs(upd.sourceInBase)
                                              else upd.sourceInBase end
FROM upd
WHERE upd.id = transactions.id
returning transactions.id, destination_amount_in_base_currency, source_amount_in_base_currency
