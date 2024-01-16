// {
// "id": "0x00124b00146c31cd",
// "friendly_name": "Motion sensor 1",
// "description": "Motion sensor",
// "connection_type": "mqtt",
// "power_source": "Battery",
// "exposes": {
//   "occupancy": {
//     "name": "occupancy",
//     "description": "Indicates whether the device detected occupancy",
//     "data": null,
//     "type": "binary",
//     "attributes": {
//       "off": false,
//       "on": true
//     }
//   },
//   "temperature": {
//     "name": "temperature",
//     "description": "Measured temperature value",
//     "unit": "°C",
//     "data": null,
//     "type": "numeric"
//   }
// },
// "properties": {
//   "availability": "offline"
// }
// }

type Device = {
  id: string;
  friendly_name: string;
  description: string;
  connection_type: string;
  power_source: string;
  exposes: { [key: string]: Expose };
  properties: { [key: string]: any };
};

type Expose = {
  name: string;
  description: string;
  unit: string;
  data: any | null;
  type: string | null;
  attributes: { [key: string]: any } | null;
  presets: { [key: string]: any } | null;
  properties: { [key: string]: any } | null;
};
