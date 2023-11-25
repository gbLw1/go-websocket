# Go WebSocket

## Overview

This is a simple websocket chat application.

The server and client were written in [Go](https://golang.org) and [VueJS](https://vuejs.org).

See the [Features](https://github.com/gbLw1/go-websocket/releases)

## Example

![example](./docs/example_darkmode.png)

## Dependencies

The websocket packaged used in this project is:

- [x] [nhooyr.io/websocket](https://github.com/nhooyr/websocket)

## Run

### Testing the app locally

The `./public/script.js` file has a `ENVIRONMENTS` const that can be used to
switch between the local server and the production server.

1. find and replace the `ENVIRONMENTS` const in `./public/script.js`:

   search for `.PROD.` and replace it with `.DEV.`

   ```bash
   :%s/\.PROD/.DEV/g
   ```

2. run the server:

   ```bash
   go run ./main.go
   ```

### Connect to the server only

To open a new WebSocket connection to the server, you have to pass the following query parameters:

- `nickname`: the nickname of the user
- `room`: the room name (optional, default: `general`)

Example:

1. Connect to the `general` room with the nickname `John`:

   ```javascript
   const ws = new WebSocket(
     `ws://localhost:3000/ws?nickname=John&room=general`
   );
   ```

2. Subscribe to see the messages from the ws:

   ```javascript
   ws.onmessage = (event) => {
     console.log(event.data);
   };
   ```

   if you don't want to get spammed with the notifications, you can filter the messages:

   ```javascript
   ws.onmessage = (event) => {
     const data = JSON.parse(event.data);
     if (data.type === "message") {
       console.log(data);
     }
   };
   ```

3. Send a message to the server:

   You can send two types of messages to the server: `message` and `notification`.

   - `message`: a message that will be displayed in the chat
   - `notification`: the sender will be added to a list of who is typing (displayed below the chat)

    3.1. Payload message format:

   ```javascript
   {
       "type": "message", // string (required): accepted values = "message", "notification"
       "from": {
           "nickname": "John", // string (required): your nickname
           "color": "#000000", // string (optional): hexadecimal color (default: #000000)
       },
       content: "Hello world!", // string (required for "message" type): the message content
       isTyping: true, // boolean (required for "notification" type): notifies other connected clients that you are typing
   }
   ```

   3.2. Sending a Hello world! message:

   ```javascript
   ws.send(
     JSON.stringify({
       type: "message",
       from: {
         nickname: "John",
       },
       content: "Hello world!",
     })
   );
   ```
