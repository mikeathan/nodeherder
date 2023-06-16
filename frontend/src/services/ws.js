import store from "../store/store.js";

const socketUri = "ws://localhost:3000/ws"; // used for testing
//const socketUri = "ws://"+document.location.host+"/ws"

console.log("socketUri:" + socketUri);
const ws = new WebSocket(socketUri);

ws.onmessage = (event) => {
  const obj = JSON.parse(event.data);
  if (event.data == undefined) {
    console.log("ws undefined data: " + event);
    return;
  }

  var name = obj.type;
  var payload = obj.payload;

  if (name == "connected") {
    console.log("connected");
    store.commit("init", payload);
  } else if (name == "deviceUpdated") {
    console.log("deviceUpdated");
    store.commit("deviceUpdated", payload);
  } else {
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
