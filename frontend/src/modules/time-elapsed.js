import 'moment-timezone';
import moment from 'moment-timezone';
import { format } from 'timeago.js';

const SEC_ARRAY = [
  60, // 60 seconds in 1 min
  60, // 60 mins in 1 hour
  24, // 24 hours in 1 day
  7, // 7 days in 1 week
  365 / 7 / 12, // 4.345238095238096 weeks in 1 month
  12, // 12 months in 1 year
];

export default class ElapsedTimer {
  constructor(element) {
    this.__element = element;
  }

  stopElapsedTimer() {
    if (this.__elapsedTimerId != undefined) {
      clearInterval(this.__elapsedTimerId);
      //console.log("[DEBUG] clearInterval " + this.__elapsedTimerId);
    }
  }

  dispose() {
    this.stopElapsedTimer();
  }

  startElapsedTimer(timestamp) {
    if (!timestamp) {
      this.TimeElapsed = 'NA';
      return;
    }

    this.stopElapsedTimer();

    const diff = diffSec(timestamp);
    const interval = nextInterval(diff) * 1000;

    this.TimeElapsed = format(timestamp, 'en_UK');
    this.__element.innerText = this.TimeElapsed;
    this.__elapsedTimerId = setInterval(
      function () {
        this.startElapsedTimer(timestamp);
      }.bind(this),
      interval
    );
  }

  Format(timestamp) {
    this.startElapsedTimer(timestamp);
  }

  // Format(timestamp, updaterCallback) {
  //   if (isCallback(updaterCallback)) {
  //     this.updaterCallback = updaterCallback;
  //     this.startElapsedTimer(timestamp);
  //   } else {
  //     this.TimeElapsed = format(timestamp, "en_UK");
  //   }
  // }
}

function diffSec(date) {
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
