const { createApp, ref } = Vue;

const ENVIRONMENTS = {
  PROD: {
    BASE_WS_URL: "wss://go-websocket-production.up.railway.app",
    BASE_HTTP_URL: "https://go-websocket-production.up.railway.app",
  },
  DEV: {
    BASE_WS_URL: "ws://localhost:3000",
    BASE_HTTP_URL: "http://localhost:3000",
  },
};

createApp({
  data() {
    return {
      nickname: localStorage.getItem("nickname") || "",
      color: this.generateRandomHexColor(),
      room: "",
      connected: false,
      ws: null,
      message: "",
      messages: [],
      clients: [],
      unreadMessages: 0,
      isTyping: false,
      clientsTyping: [],
    };
  },

  created() {
    this.verifyRoomFromQuery();
    this.handleVisibilityChange();
  },

  methods: {
    generateRandomHexColor() {
      return (
        "#" +
        Math.floor(Math.random() * 16777215)
          .toString(16)
          .padStart(6, "0")
      );
    },

    handleVisibilityChange() {
      document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "visible") {
          this.resetChatNotifications();
        }
      });
    },

    resetChatNotifications() {
      this.unreadMessages = 0;
      document.title = `Chat - ${this.room || "Welcome"}`;
    },

    createSlug(str) {
      let slug = str?.normalize("NFD").replace(/[\u0300-\u036f]/g, ""); // remove accents
      slug = slug?.replace(/[^\w\s-]/g, "").toLowerCase(); // remove special characters
      slug = slug?.replace(/\s+/g, "-"); // replace spaces with dash
      return slug;
    },

    verifyRoomFromQuery() {
      const params = new URLSearchParams(window.location.search);
      const room = params.get("room");

      if (room) {
        this.room = room;
        this.connect();
      }
    },

    sendMessage() {
      if (!this.ws.readyState === this.ws.OPEN) {
        console.log("Connection lost, reconnecting...");
        this.connect();
        return;
      }

      if (this.isTyping) {
        this.isTyping = false;

        this.ws.send(
          JSON.stringify({
            type: "notification",
            from: {
              nickname: this.nickname,
            },
            isTyping: false,
          }),
        );

        this.clientsTyping = this.clientsTyping.filter(
          (c) => c !== this.nickname,
        );
      }

      const msg = {
        type: "message",
        from: {
          nickname: this.nickname,
          color: this.color,
        },
        content: this.message,
      };

      this.ws.send(JSON.stringify(msg));
      this.message = "";

      // Scroll to bottom of chat
      const chatMessages = document.querySelector(".chat-messages");
      chatMessages.addEventListener("DOMSubtreeModified", () => {
        chatMessages.scrollTop = chatMessages.scrollHeight;
      });
    },

    onOpen(event) {
      this.connected = true;
    },

    onMessage(event) {
      const data = JSON.parse(event.data);

      switch (data.type) {
        case "message":
          this.handleMessage(data);
          break;
        case "notification":
          this.handleNotification(data);
          break;
        default:
          break;
      }
    },

    handleMessage(data) {
      this.messages.push(data);

      if (this.messages.length > 100) {
        this.messages.shift();
      }

      this.updateConnectedClients();

      if (window.document.visibilityState === "visible") {
        this.resetChatNotifications();
      } else {
        this.unreadMessages++;

        document.title = `Chat (${
          this.unreadMessages > 99 ? "+99" : this.unreadMessages
        }) - ${this.room}`;

        var audio = new Audio("notification.mp3");
        audio.play();
      }
    },

    handleNotification(data) {
      const { from, isTyping } = data;

      if (isTyping) {
        if (
          this.clientsTyping.includes(from.nickname) ||
          from.nickname === this.nickname
        ) {
          return;
        }

        this.clientsTyping.push(from.nickname);
      } else {
        this.clientsTyping = this.clientsTyping.filter(
          (c) => c !== from.nickname,
        );
      }
    },

    handleInput(event) {
      setTimeout(() => {
        const inputHasValue = event.target.value.trim().length > 0;

        if (this.isTyping !== inputHasValue) {
          this.isTyping = inputHasValue;
          this.ws.send(
            JSON.stringify({
              type: "notification",
              from: {
                nickname: this.nickname,
              },
              isTyping: inputHasValue,
            }),
          );
        }
      }, 200);
    },

    async connect() {
      this.room = this.createSlug(this.room);

      if (!this.room) {
        this.room = "general";
      }

      if (!this.nickname) {
        this.nickname = `Guest${Math.floor(Math.random() * 1000)}`;
      }

      if (this.nickname.toUpperCase() === "SERVER") {
        alert("Nickname is not allowed");
        return;
      }

      if (!this.nickname.toLowerCase().includes("guest")) {
        localStorage.setItem("nickname", this.nickname);
      }

      // Check if nickname is already in use
      const res = await fetch(
        `${ENVIRONMENTS.PROD.BASE_HTTP_URL}/clients?room=${this.room}`,
      );
      const connectedClients = await res.json();

      if (connectedClients?.some((c) => c.nickname === this.nickname)) {
        alert("Nickname is already in use");
        return;
      }

      this.ws = new WebSocket(
        `${ENVIRONMENTS.PROD.BASE_WS_URL}/ws?nickname=${this.nickname}&room=${this.room}`,
      );
      this.ws.onopen = this.onOpen;
      this.ws.onmessage = this.onMessage;
      this.ws.onclose = (e) =>
        console.log("Connection closed at ", new Date().toLocaleString(), e);

      history.pushState({}, "", `/?room=${this.room || "general"}`);
    },

    disconnect() {
      this.ws.close();
      this.connected = false;
      this.message = "";
      this.messages = [];
      this.clients = [];
      this.room = "";

      history.pushState({}, "", "/");
      localStorage.removeItem("nickname");
      document.title = "Chat - Welcome";
    },

    async updateConnectedClients() {
      try {
        const res = await fetch(
          `${ENVIRONMENTS.PROD.BASE_HTTP_URL}/clients?room=${this.room}`,
        );

        const data = await res.json();
        this.clients = data;
      } catch (error) {
        console.log(error);
      }
    },
  },
}).mount("#app");
