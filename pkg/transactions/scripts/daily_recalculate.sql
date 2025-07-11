WITH minDate as (select least(coalesce(
                                (select (@startDate)::date -- if we dont have any daily_stat its an edge case, we should fallback to latest date 
                                 from daily_stat st2
                                 where st2.account_id = @accountID
                                   and st2.date < (@startDate)::date
                                 order by date desc
                                 limit 1)::date, (select min(transaction_date_only)
                                                  from transactions
                                                  where source_account_id = @accountID
                                                     or destination_account_id = @accountID
                                                  limit 1)::date, (@startDate)::date),(@startDate)::date) as minDate), -- last fallback to startDate from backend
     date_series AS (SELECT generate_series(
                                    (select * from minDate),
                                    GREATEST(NOW()::DATE, (select max(transaction_date_only)
                                                           from transactions
                                                           where source_account_id = @accountID
                                                              or destination_account_id = @accountID)) +
                                    1, -- 1 day to get current
                                    '1 day'::INTERVAL
                            ) ::DATE AS date),
     daily_sums as (select coalesce(sum(coalesce(
             case when source_account_id = @accountID then source_amount else destination_amount end, 0)), 0) as amount,
                           transaction_date_only                                                              as tx_date
                    from transactions
                    where (source_account_id = @accountID
                        or destination_account_id = @accountID)
                      and transaction_date_only in (select * from date_series)
                    group by transaction_date_only),
     lastestValue as (select st2.amount
                      from daily_stat st2
                      where st2.account_id = @accountID
                        and st2.date < (select * from minDate)
                      order by date desc
                      limit 1),
     initialValue as (select (select min(date) from date_series d)          as date,
                             (select coalesce(amount, 0) from lastestValue) as amount),
     running as (select d.date,
                        1,
                        sum(coalesce(s.amount, 0) + coalesce(initial.amount, 0)) over (
                            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
                            ) as amount
                 from date_series d
                          left join initialValue initial on initial.date = d.date
                          left join daily_sums s
                                    on s.tx_date = d.date
                 order by d.date asc),
     lastestRunning as (select coalesce(amount, 0) as amount, date
                        from running
                        where date = (select max(date) from running)),
     udpatedCurrentBalance as (update accounts set
         current_balance = coalesce((select amount from lastestRunning), 0),
         last_updated_at = timezone('utc', now())
         where id = @accountID
         returning current_balance)
insert
into daily_stat(account_id, date, amount)
select @accountID, date, coalesce(amount, 0)
from running
on conflict ON CONSTRAINT daily_stat_pk do update set amount = excluded.amount
