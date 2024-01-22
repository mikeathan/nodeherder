const devSocketUri = "ws://localhost:3000/ws";
const productionSocketUri = "ws://" + document.location.host + "/ws";

const maxNumberOfAttempts = 10;
const intervalTimeMs = 200;
var socketUri = getSocketUri();

export function getSocketUri() {
    if (process.env.NODE_ENV == "development") {
        console.info("Enviroment:", process.env.NODE_ENV);
        return devSocketUri;
    }

    return productionSocketUri;
}

export function send() {
    todo
}