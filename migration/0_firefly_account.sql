select a.name,
       tc.code                           as currency,
       case a.account_type_id
           when 11 then 4 -- liability
           else
               1 -- regular
           end                           as type,
       (select n.text
        from notes n
        where n.noteable_id = a.id
          and n.noteable_type = 'FireflyIII\Models\Account'
                                            limit 1)                         as note,
       (select case
                   WHEN n.text ~ '^\s*[\{\[]' THEN (SELECT jsonb_object_agg(key, value)
                                                    FROM jsonb_each_text(n.text::jsonb))
                   ELSE jsonb_build_object('note', n.text) end
        from notes n
        where n.noteable_id = a.id
          and n.noteable_type = 'FireflyIII\Models\Account'
        limit 1)                         as extra,
       a.iban,
       (select replace(am.data, '"', '')
        from account_meta m
        where m.account_id = a.id
          and m.name = 'account_number') as account_number
from accounts a
    join public.account_meta am on a.id = am.account_id and am.name = 'currency_id'
    join public.transaction_currencies tc on replace(am.data, '"', '')::int = tc.id
where a.account_type_id not in (6, 13, 10, 2)
order by a."order" asc
