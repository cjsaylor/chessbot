# Privacy

ChessBot is a Slack bot that operates for fun. As such, we use as little as possible identifying information in order to make the bot work.

## 1. Information We Collect

We collect the following information from Slack in order to operate the game:

* Team ID, ex: `TXXXXXXX`. This allows the bot to select the correct authorization token in order to interact with your workspace.
* Authorization Token, the token granted by Slack to our bot in order to securely communicate with your workspace.
* User IDs involved with the game, ex: `UXXXXXX`
* Channel IDs for game communication, ex: `CXXXXXX`
* Chess moves made (stored in `PGN` format)

### Logging

Our web server will log information about your request (such as your browser user-agent and IP address). This is necessary for service health monitoring as well as to prevent abuse of the service.

## 2. Security

We store this information in a database that is not accessible from the internet.

## 3. Third party access

We share User IDs and chess moves made in games with Lichess.org in order to import it for your analysis. This is not done automatically, it requires a player to click the "Analyze Game" link at the end of the game.

## 4. Cookies

We do not use cookies.

## 5. Portability

If requested, we will attempt to get all information pertaining to your user ID.