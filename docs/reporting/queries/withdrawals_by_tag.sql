select coalesce(tg.name, 'UNK'), sum(t.source_amount_in_base_currency)
from transactions t
         left join lateral unnest(t.tag_ids) as tag_id on true
         left join tags tg on tg.id = tag_id
where transaction_date_only >= '2025-06-01'
  and t.transaction_type = 3
group by tg.name;
