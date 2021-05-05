# Buildwatch

This is a discord bot that will listen on `*:5436/build` for a web hook `POST` and send an embed message to a specified channel in a Discord server.

## Installation

Coming soon. Maybe this will work...
```
go get github.com/ChrisB92/tc-build-bot
```

## Usage

Create a `config.json` file in the `bot` folder
```json
{
    "token": "",
    "mainChannelId": "",
    "authToken": ""
}
```

You can then `go build` and run the bot calling the executable `bot`

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
[MIT](https://choosealicense.com/licenses/mit/)