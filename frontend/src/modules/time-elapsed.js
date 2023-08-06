import "moment-timezone";
import moment from "moment-timezone";
import { format, render } from "timeago.js";
const SEC_ARRAY = [
  60, // 60 seconds in 1 min
  60, // 60 mins in 1 hour
  24, // 24 hours in 1 day
  7, // 7 days in 1 week
  365 / 7 / 12, // 4.345238095238096 weeks in 1 month
  12, // 12 months in 1 year
];
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

    const diff = diffSec(timestamp);
    var interval = nextInterval(diff) * 1000;
    console.log(interval);

    this.__elapsedTimerId = setInterval(function () {
      this.TimeElapsed = format(timestamp, "en_UK");
      var diff = diffSec(timestamp);
      interval = nextInterval(diff) * 1000;
      console.log(interval);

      callback(this.TimeElapsed);
    }, interval);
  }

  SetTimestamp(timestamp, updaterCallback) {
    this.TimeElapsed = format(timestamp, "en_UK");
    if (isCallback(updaterCallback)) {
      this.startElapsedTimer(timestamp, updaterCallback);
    }
  }

  // Format(timestamp, htmlElement) {

  // }
}

export function diffSec(date) {
  const relDate = new Date();
  return (+relDate - +moment(date)) / 1000;
}
function nextInterval(diff) {
  let rst = 1,
    i = 0,
    d = Math.abs(diff);
  for (; diff >= SEC_ARRAY[i] && i < SEC_ARRAY.length; i++) {
    diff /= SEC_ARRAY[i];
    rst *= SEC_ARRAY[i];
  }
  d = d % rst;
  d = d ? rst - d : rst;
  return Math.ceil(d);
}

function isCallback(callback) {
  return callback && typeof callback == "function";
}
