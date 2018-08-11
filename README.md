# ChessBot

[![GoDoc](https://godoc.org/github.com/cjsaylor/chessbot?status.svg)](https://godoc.org/github.com/cjsaylor/chessbot)

> Currently under development

This is a Slack bot that allows players in a channel to challenge each other to a game of Chess.

![](./doc/screenshot.png)

## Quick Start

```
go run cmd/web/web.go
```

## Configuration

| Environment Variable | Default | Description
| --- | --- | ---
| PORT | `8080` | Port that the web server will listen
| HOSTNAME | `localhost:8080` | Used for generating links to render the game board state images
| SIGNINGKEY | N/A | Key used to sign the signature for board rendering URLs
| SQLITEPATH | N/A | Path to a sqlite3 database file. If not included, falls back to memory store.
| SLACKBOTTOKEN | N/A | The app bot's oauth token. This can be retrieved from the slack app basic info screen
| SLACKVERIFICATIONTOKEN | N/A | (deprecated) This is used to verify `POST`s to `/slack` and `/slack/action` originate from slack

## Endpoints

```
GET /board?fen=&signature=&from=&to=&check=
```

Renders the game board based on the state of a game by FEN.

```
POST /slack
```

All slack event subscription callbacks flow through this.

* This is used for all typed commands mentioning `@ChessBot` in the channel.

```
POST /slack/action
```

All slack interactive component callbacks flow through this.

* This is used for accepting/rejecting challenges.

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