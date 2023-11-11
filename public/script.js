const { createApp, ref } = Vue;

createApp({
  onMounted() {
    // Set room from url params
    const urlParams = new URLSearchParams(window.location.search);
    this.room = urlParams.get("room");

    // Scroll to bottom of chat
    const chatMessages = document.querySelector(".chat-messages");
    chatMessages.addEventListener("MutationObserver", () => {
      chatMessages.scrollTop = chatMessages.scrollHeight;
    });
  },
  data() {
    return {
      nickname: "",
      room: "",
      connected: false,
      ws: null,
      message: "",
      messages: [],
      clients: [],
    };
  },

  methods: {
    sendMessage() {
      const msg = {
        from: this.nickname,
        to: this.room,
        content: this.message,
      };
      this.ws.send(JSON.stringify(msg));
      this.message = "";
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
    },

    connect() {
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

      this.ws = new WebSocket(
        `wss://go-websocket-production.up.railway.app/ws?nickname=${this.nickname}&room=${this.room}`,
      );
      this.ws.onopen = this.onOpen;
      this.ws.onmessage = this.onMessage;

      window.location.href = `?room=${this.room}`;
    },

    disconnect() {
      this.ws.close();
      this.connected = false;
      this.ws = null;
      this.message = "";
      this.messages = [];
      this.clients = [];
    },

    async updateConnectedClients() {
      try {
        const res = await fetch(
          `https://go-websocket-production.up.railway.app/clients?room=${this.room}`,
        );
        const data = await res.json();
        this.clients = data;
      } catch (error) {
        console.log(error);
      }
    },
  },
}).mount("#app");
