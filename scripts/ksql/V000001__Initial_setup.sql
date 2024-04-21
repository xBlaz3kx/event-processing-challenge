CREATE STREAM casino_events_stream (
    id INT,
    player_id INT,
    game_id INT,
    type VARCHAR,
    amount INT,
    currency VARCHAR,
    amount_eur INT,
    created_at VARCHAR,
    player STRUCT<email VARCHAR,last_signed_in_at VARCHAR>
)
WITH (KAFKA_TOPIC='casino-event-materialize', VALUE_FORMAT='JSON', PARTITIONS=1);

CREATE TABLE casino_player_stats AS
SELECT player_id,
       SUM(CASE WHEN type = 'bet' THEN 1 ELSE 0 END)              AS player_bets,
       SUM(CASE WHEN type = 'win' THEN 1 ELSE 0 END)              AS player_wins,
       SUM(CASE WHEN type = 'deposit' THEN amount_eur ELSE 0 END) AS player_deposits
FROM casino_events_stream
    WINDOW TUMBLING (SIZE 1 MINUTE)
    GROUP BY player_id;