function formatLastSeen(payload) {
  if (payload["last_seen"] != undefined) {
    return "";
  }

  var sensorLastSeen = payload["last_seen"];
  var lastSeen = moment(sensorLastSeen);
  var diff = moment().diff(lastSeen);
  var duration = moment.duration(diff);

  // TODO: handle days() > 0
  if (duration.hours() > 0) {
    return duration.hours() + " hours ago";
  }
  if (duration.minutes() > 0) {
    return duration.minutes() + " minutes ago";
  }
  if (duration.seconds() > 5) {
    return duration.seconds() + " seconds ago";
  }

  return "just now";
}

export function formatDeviceInfo(payload) {}

function formatLinkQuality(payload) {
  if (payload["linkquality"] != undefined) {
    return "";
  }

  //<i className="fa fa-signal fa-fw" /> {linkquality} LQI
  var linkQuality = payload["linkquality"];

  return linkQuality + " LQI";
}
