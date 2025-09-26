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
                        end as sourceInBase
             from transactions t
                      left join currencies sourceCurrency on sourceCurrency.id = t.source_currency
                      left join currencies destinationCurrency on destinationCurrency.id = t.destination_currency
             where ((@specificTxIDs)::bigint[] IS NULL
                OR t.id = ANY ((@specificTxIDs)::bigint[]))
               and t.deleted_at IS NULL)
UPDATE transactions
SET destination_amount_in_base_currency = abs(upd.sourceInBase),
    source_amount_in_base_currency      = -abs(upd.sourceInBase)
FROM upd
WHERE upd.id = transactions.id
returning transactions.id, destination_amount_in_base_currency, source_amount_in_base_currency
