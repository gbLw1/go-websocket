const { createApp, ref } = Vue;

createApp({
  data() {
    return {
      nickname: "",
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
        content: this.message,
      };
      this.ws.send(JSON.stringify(msg));
      this.message = "";

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
    },

    connect() {
      if (!this.nickname) {
        this.nickname = `Guest${Math.floor(Math.random() * 1000)}`;
      }

      if (this.nickname.toUpperCase() === "SERVER") {
        alert("Nickname is not allowed");
        return;
      }

      this.ws = new WebSocket(
        `ws://localhost:3000/ws?nickname=${this.nickname}`,
      );
      this.ws.onopen = this.onOpen;
      this.ws.onmessage = this.onMessage;
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
        const res = await fetch("http://localhost:3000/clients");
        const data = await res.json();
        this.clients = data;
      } catch (error) {
        console.log(error);
      }
    },
  },
}).mount("#app");
