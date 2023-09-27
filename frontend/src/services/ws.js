import store from "../store/store.js";

const socketUri = "ws://localhost:3000/ws"; // used for testing
//const socketUri = "ws://" + document.location.host + "/ws";
var ws = create();
var messageQueue = [];

function create() {
  console.log("ws create");
  console.log("socketUri:" + socketUri);

  return new WebSocket(socketUri);
}

export function emit(event, message = "") {
  var payload = JSON.stringify({ type: event, payload: message });
  console.log("ws emit");
  if (ws.readyState !== 1) {
    messageQueue.push(payload);
  } else {
    ws.send(payload);
  }
}

export default function connect() {
  console.log("ws connect");

  ws.onmessage = (event) => {
    if (event == undefined) {
      console.log("ws undefined event: " + event);
      return;
    }
    if (event.data == undefined) {
      console.log("ws undefined data: " + event.data);
      return;
    }

    const obj = JSON.parse(event.data);
    console.log("ws message received:", obj.type);

    switch (obj.type) {
      case "connected":
        store.commit("init", obj.payload);
        break;
      case "deviceUpdated":
        store.commit("deviceUpdated", obj.payload);
        break;
      case "automations":
        store.commit("initAutomations", obj.payload);
        break;
      default:
        console.log("ws unhandled type: ", event.data);
    }
  };

  ws.onopen = function (event) {
    console.log("ws open");
    while (messageQueue.length > 0) {
      ws.send(messageQueue.pop());
    }
  };
  ws.onclose = function (event) {
    console.log("ws close");
  };
  ws.onerror = function (event) {
    console.log("ws error: " + event.data);
  };

  return ws;
}
