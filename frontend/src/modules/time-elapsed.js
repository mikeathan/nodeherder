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
    var timeout = 1000;
    this.__elapsedTimerId = setInterval(function () {
      this.TimeElapsed = format(timestamp, "en_UK");
      if (
        this.TimeElapsed.includes("seconds") ||
        this.TimeElapsed.includes("now")
      ) {
        timeout = 1000;
      } else if (this.TimeElapsed.includes("minutes")) {
        timeout = 60000;
      } else {
        timeout = 3600000;
      }
      callback(this.TimeElapsed);
    }, timeout);
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
