import store from "../store/store.js";

const socketUri = "ws://localhost:3000/ws"; // used for testing
//const socketUri = "ws://" + document.location.host + "/ws";

console.log("socketUri:" + socketUri);
const ws = new WebSocket(socketUri);

export function wssend(event, payload) {
  var message = JSON.stringify({ type: event, payload: payload });
  ws.send(message);
}

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
  console.log("Open: " + event);
};
ws.onclose = function (event) {
  console.log("Close");
};
ws.onerror = function (event) {
  console.log("Error: " + event.data);
};

export default ws;
