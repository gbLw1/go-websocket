# Go WebSocket

## Overview

This is a simple websocket chat application.

The server and client were written in [Go](https://golang.org) and [VueJS](https://vuejs.org).

## Features

- [x] Choose a nickname
- [x] Choose a room (multiple chat rooms support)
- [x] Send messages
- [x] See who is online in the room
- [x] Connected clients list updated in real time
- [x] Disconnect from the room
- [x] Share room directly with URL link
- [x] Notifications from the server when someone joins or leaves the room
- [x] Notifications of unread messages in the browser tab + sound
- [x] Connected room displayed in the browser tab
- [x] Scroll down to the last message when a new message is received
- [x] Nickname saved in the browser local storage
- [x] Dark theme
- [x] Responsive design
- [ ] See who is typing
- [ ] Different colors for each user
- [ ] Send links
- [ ] Emoji support

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

3. run the server:

```bash
go run ./main.go
```

