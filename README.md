# ChessBot

> Currently under development

This is a Slack bot that allows players in a channel to challenge each other to a game of Chess.

![](./doc/screenshot.png)

## Quick Start

```
go run cmd/web/web.go
```

> See `Configuration` for all environment vars

## Endpoints

```
GET /board?game_id=
```

Renders the game board based on the state of a game by ID (slack `thread_ts`)

```
POST /slack
```

All slack event subscription callbacks flow through this.

## Testing the Chess Engine

```
go run cmd/repl/main.go
```

```
λ go run cmd/repl/main.go
Game REPL
Note the chess board is rendered backwords (white = black) :(

 A B C D E F G H
8♜ ♞ ♝ ♛ ♚ ♝ ♞ ♜
7♟ ♟ ♟ ♟ ♟ ♟ ♟ ♟
6- - - - - - - -
5- - - - - - - -
4- - - - - - - -
3- - - - - - - -
2♙ ♙ ♙ ♙ ♙ ♙ ♙ ♙
1♖ ♘ ♗ ♕ ♔ ♗ ♘ ♖

player1's turn (White)

> d2d4

 A B C D E F G H
8♜ ♞ ♝ ♛ ♚ ♝ ♞ ♜
7♟ ♟ ♟ ♟ ♟ ♟ ♟ ♟
6- - - - - - - -
5- - - - - - - -
4- - - ♙ - - - -
3- - - - - - - -
2♙ ♙ ♙ - ♙ ♙ ♙ ♙
1♖ ♘ ♗ ♕ ♔ ♗ ♘ ♖

player2's turn (Black)

> fen
rnbqkbnr/pppppppp/8/8/3P4/8/PPP1PPPP/RNBQKBNR b KQkq d3 0 1

>
```

## Why not use Slack's RTM API?

There are two reason:

1. We have to implement a web server anyways for serving the game board as a PNG to be unfurled by Slack.
2. RTM messages don't support attachments yet.

I will reconsider RTM if the messaging ability improves and I will split off the gameboard rendering webserver
to its own process at that time.