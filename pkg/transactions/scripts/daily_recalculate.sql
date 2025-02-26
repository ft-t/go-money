WITH date_series AS (SELECT generate_series(
                                    (SELECT MIN(transaction_date_only) FROM transactions where wallet_id = 1),
                                    NOW()::DATE,
                                    '1 day'::INTERVAL
                            ) ::DATE AS date), daily_sums as (
select coalesce (sum (coalesce (amount, 0)), 0) as amount, transaction_date_only as tx_date
from transactions
where wallet_id = 1
group by transaction_date_only),
    running as (
select d.date, 1, sum (s.amount) over (
    ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
    )
from date_series d
    left join daily_sums s
on s.tx_date = d.date
order by d.date asc)
select *
from running;
