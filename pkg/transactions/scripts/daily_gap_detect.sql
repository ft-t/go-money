WITH date_series AS (SELECT generate_series(
                                    ?::date, -- source
                                    ?::DATE, -- to
                                    '1 day'::INTERVAL
                            ) ::DATE AS date)
select 1
from date_series s
         left join daily_stat d on d.account_id = 1 and d.date = s.date
where d.account_id is null limit 1
