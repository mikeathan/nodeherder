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

  // todo:
  // first minute update very sec and after than every min
  // same for 1 hour, update eveyr hour
  startElapsedTimer(timestamp, callback) {
    this.stopElapsedTimer();

    this.__elapsedTimerId = setInterval(function () {
      this.TimeElapsed = format(timestamp, "en_UK");
      callback(this.TimeElapsed);
    }, 60000);
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
