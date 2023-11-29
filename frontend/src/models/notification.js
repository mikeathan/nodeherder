export class Notification {
  constructor() {
    this.id = "";
    this.type = "";
    this.message = "";
    this.autoClose = false;
    this.duration = 1;
  }
}

export function createNotification() {
  return new Notification();
}
