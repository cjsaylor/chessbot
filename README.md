# ChessBot

> Currently under development (not even Alpha)

This is a Slack bot that allows players in a channel to challenge each other to a game of Chess.

## Quick Start

```
go run cmd/web/web.go
```

## Endpoints

```
GET /board?fen=&last_move=&checked_tile=
```

When games start getting persisted, the above will probably be replaced by a game ID to pull the fen/last_move/checked_title

## Why not use Slack's RTM API?

There are two reason:

1. We have to implement a web server anyways for serving the game board as a PNG to be unfurled by Slack.
2. RTM messages don't support attachments yet.

I will reconsider RTM if the messaging ability improves and I will split off the gameboard rendering webserver
to its own process at that time.