# Event processing challenge

A casino has a few different types of events coming in that we would like to
process to get an insight into casino performance and player activity.

For this assignment, you will need to implement a set of components that:

1. publish generated events,
2. subscribe to these events,
3. enrich them using data from various sources (HTTP API, database, and in-memory mapping),
4. materialize various aggregates,
5. output events as logs.

## Setup

Clone the repository and run `make`. This will start the `database` service, which is a Postgres server, and run the
needed DB migrations.

Optionally, you can run `make generator` to see how the generator works. It will run for 5 seconds, logging the
generated events, then exit.

You are now ready to start building the components.

## Event structure

You can find the event structure in [./internal/casino/event.go](./internal/casino/event.go).

## Publish

A `generator` service has been provided that generates random events. Add the ability to publish these events. Feel free
to decide which transport and/or technology to use.

## Subscribe

Implement a service that will subscribe to the published events. Again, feel free to decide which technology to use.

## Enrich

Implement 3 components (services) that will receive the events and enrich them:

### Common currency

To be able to analyze `bet` and `deposit` events, we need them in a common currency.

If `Currency` is not `EUR`, we need to convert the `Amount` to `EUR`. For this, use API
endpoint https://api.exchangerate.host/latest. Store the `EUR` amount into `AmountEUR` field.

If `Currency` is already `EUR`, set `AmountEUR` to the same value as `Amount`.

API results may be cached for up to 1 minute. Feel free to decide on the kind of caching technology you want to use.

### Player data

We have a Postgres database (`database` service) where you can find a `players` table with some of the players inserted.

We need to look up a player for each event and store their data into `Player` field. If there is no data for a player,
log that it's missing and leave the field as zero-value.

DB results may not be cached.

### Human-friendly description

We need to represent each event with a human-friendly description. Examples:

```json
{
  "id": 1,
  "player_id": 10,
  "game_id": 100,
  "type": "game_start",
  "created_at": "2022-01-10T12:34:56.789+00"
}
```

```
Player #10 started playing a game "Rocket Dice" on January 10th, 2022 at 12:34 UTC.
```

```json
{
  "id": 2,
  "player_id": 11,
  "game_id": 101,
  "type": "bet",
  "amount": 500,
  "currency": "USD",
  "amount_eur": 468,
  "created_at": "2022-02-02T23:45:67.89+00",
  "player": {
    "email": "john@example.com",
    "last_signed_in_at": "2022-02-02T23:01:02.03+00"
  }
}
```

```
Player #11 (john@example.com) placed a bet of 5 USD (4.68 EUR) on a game "It's bananas!" on February 2nd, 2022 at 23:45 UTC.
```

```json
{
  "id": 3,
  "player_id": 12,
  "type": "deposit",
  "amount": 10000,
  "currency": "EUR",
  "created_at": "2022-03-03T12:12:12+00"
}
```

```
Player #12 made a deposit of 100 EUR on February 3rd, 2022 at 12:12 UTC.
```

You can find the mapping of game titles in [./internal/casino/game.go](./internal/casino/game.go).

## Materialize

We would like to materialize the following data:

- total number of events,
- number of events per minute,
- events per second as a moving average in the last minute,
- top player by number of bets,
- top player by number of wins,
- top player by sum of deposits in EUR.

Feel free to decide on the algorithm, technology or library you want to use. This data should not be persisted and
should only be available while the components are running.

Data should be available via an HTTP API in the following form:

```
GET http://localhost/materialized
```

```json
{
  "events_total": 12345,
  "events_per_minute": 123.45,
  "events_per_second_moving_average": 3.12,
  "top_player_bets": {
    "id": 10,
    "count": 150
  },
  "top_player_wins": {
    "id": 11,
    "count": 50
  },
  "top_player_deposits": {
    "id": 12,
    "count": 15000
  }
}
```

## Output

Log the events in their final form to standard output. Logs should be in JSON format and use the same keys as
the `Event` type.

## Solution

I wrapped the event generator in a struct, just because I like the "Object oriented" approach a bit more.

Tech stack:

- Go
- GORM (for Postgres)
- Redis (for caching)
- Gin (HTTP Framework)
- Docker (with multistage image build and go cache for faster build times) and Docker Compose
- Benthos (service (framework?) for event processing/transformations/enrichment)
- Grafana LGTM stack (observability, because why not)
- Traefik (local API Gateway)

### Components and explanation

I chose to use an existing solution (`Benthos`) for data enrichment and transformation instead of implementing it from
scratch. The logging and materialization part could also be done in Benthos, but I wanted to display my understanding
of the problem and the ability to write code as well as usage of external tools, creating a hybrid solution.

The Player data, currency and description enrichment are processed in Benthos - all pipelines call the service's HTTP
API to get the enrichment data and then forward it to the next Kafka topic in the following order:
`casino-event -> casino-event-currency -> casino-event-player-data -> casino-event-description`.

The last topic's consumer is implemented in the service, so we can perform materialization and logging of the event.

### Features

- Added CI/CD steps (lint, test and build)
- Added observability (just because)
- Added health checks for various components, available at `/healthz`
- Added local API Gateway (Traefik)

### Possible improvements

- Unit tests
- Tracing and metrics via OTel for better observability
- The whole task could be done using Benthos (mappings, even caching, etc.) but I wanted to show my programing skills
  instead of configuration skills
- Adding proper documentation (architecture, components, flows, etc.)