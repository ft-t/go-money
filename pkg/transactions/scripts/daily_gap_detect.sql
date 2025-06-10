WITH date_series AS (SELECT generate_series(
                                    ?::date,
                                    ?::DATE,
                                    '1 day'::INTERVAL
                            ) ::DATE AS date)
select 1 as rec
from date_series s
         left join daily_stat d on d.account_id = ? and d.date = s.date
where d.account_id is null limit 1
