select a.id, a.name, tc.code, a.account_type_id, a."order"
from accounts a
         join public.account_meta am on a.id = am.account_id and am.name = 'currency_id'
         join public.transaction_currencies tc on replace(am.data, '"', '')::int = tc.id
where a.account_type_id not in (6, 13, 10, 2)
order by a."order" asc
