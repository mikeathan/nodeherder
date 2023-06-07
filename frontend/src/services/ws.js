import store from "../store/store.js";

const socketUri = "ws://localhost:3000/ws"; // used for testing
//const socketUri = "ws://"+document.location.host+"/ws"

console.log("socketUri:" + socketUri);
const ws = new WebSocket(socketUri);

ws.onmessage = (event) => {
  const obj = JSON.parse(event.data);
  if (event.data == undefined) {
    return;
  }
  store.commit("deviceUpdated", obj);
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
