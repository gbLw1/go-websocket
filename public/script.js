const { createApp, ref } = Vue;

createApp({
  data() {
    return {
      nickname: localStorage.getItem("nickname") || "",
      room: "",
      connected: false,
      ws: null,
      message: "",
      messages: [],
      clients: [],
      unreadMessages: 0,
    };
  },

  created() {
    this.verifyRoomFromQuery();
    this.handleVisibilityChange();
  },

  methods: {
    handleVisibilityChange() {
      document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "visible") {
          this.unreadMessages = 0;
          document.title = `Chat - ${this.room}`;
        }
      });
    },

    createSlug(str) {
      // remove accents
      let slug = str?.normalize("NFD").replace(/[\u0300-\u036f]/g, "");

      // remove special characters
      slug = slug?.replace(/[^\w\s-]/g, "").toLowerCase();

      // replace spaces with dash
      slug = slug?.replace(/\s+/g, "-");

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
      const msg = {
        from: this.nickname,
        to: this.room,
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
      this.messages.push(data);

      if (this.messages.length > 50) {
        this.messages.shift();
      }

      this.updateConnectedClients();

      if (window.document.visibilityState === "visible") {
        this.unreadMessages = 0;
        document.title = `Chat - ${this.room}`;
      } else {
        this.unreadMessages++;
        document.title = `Chat (${
          this.unreadMessages > 99 ? "+99" : this.unreadMessages
        }) - ${this.room}`;
      }
    },

    connect() {
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

      this.ws = new WebSocket(
        `wss://go-websocket-production.up.railway.app/ws?nickname=${this.nickname}&room=${this.room}`, // production
        // `ws://localhost:3000/ws?nickname=${this.nickname}&room=${this.room}`, // local
      );
      this.ws.onopen = this.onOpen;
      this.ws.onmessage = this.onMessage;

      history.pushState({}, "", `/?room=${this.room || "general"}`);
    },

    disconnect() {
      this.ws.close();
      this.connected = false;
      this.ws = null;
      this.message = "";
      this.messages = [];
      this.clients = [];

      history.pushState({}, "", "/");
      localStorage.removeItem("nickname");
    },

    async updateConnectedClients() {
      try {
        const res = await fetch(
          `https://go-websocket-production.up.railway.app/clients?room=${this.room}`, // production
          // `http://localhost:3000/clients?room=${this.room}`, // local
        );

        const data = await res.json();
        this.clients = data;
      } catch (error) {
        console.log(error);
      }
    },
  },
}).mount("#app");
