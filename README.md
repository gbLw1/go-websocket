# Go WebSocket

## Overview

This is a simple websocket chat application.

The server and client were written in [Go](https://golang.org) and [VueJS](https://vuejs.org).

## Example

![example](./docs/example_darkmode.png)

## Dependencies

The websocket packaged used in this project is:

- [x] [nhooyr.io/websocket](https://github.com/nhooyr/websocket)

## Run

### Testing locally:

1. change the `./public/script.js` file to use the local server on connect() method:

```javascript
this.ws = new WebSocket(
    `ws://localhost:3000/ws?nickname=${this.nickname}&room=${this.room}`
);
```

2. also change the updateConnectedClients() method to use the local server:

```javascript
const res = await fetch(
    `https://localhost:3000/clients?room=${this.room}`,
);
```

```bash
go run ./main.go
```

