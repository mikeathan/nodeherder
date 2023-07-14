import "moment-timezone";
import { format } from "timeago.js";

export default class ElapsedTimer {
  constructor() {}

  stopElapsedTimer() {
    if (this.__elapsedTimerId != undefined) {
      clearInterval(this.__elapsedTimerId);
      //console.log("[DEBUG] clearInterval " + this.__elapsedTimerId);
    }
  }

  dispose() {
    this.stopElapsedTimer();
  }

  startElapsedTimer(timestamp, callback) {
    this.stopElapsedTimer();

    this.__elapsedTimerId = setInterval(function () {
      this.TimeElapsed = format(timestamp, "en_UK");
      callback(this.TimeElapsed);
    }, 1000); // change that to minutes
  }

  SetTimestamp(timestamp, updaterCallback) {
    this.TimeElapsed = format(timestamp, "en_UK");
    if (isCallback(updaterCallback)) {
      this.startElapsedTimer(timestamp, updaterCallback);
    }
  }
}
function isCallback(callback) {
  return callback && typeof callback == "function";
}
