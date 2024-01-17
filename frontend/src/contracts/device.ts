export type Device = {
  id: string;
  friendly_name: string;
  description: string;
  connection_type: string;
  power_source: string;
  exposes: { [key: string]: Expose };
  properties: { [key: string]: any };
};

export type Expose = {
  name: string;
  description: string;
  unit: string;
  data: any | null;
  type: string | null;
  attributes: { [key: string]: any } | null;
  presets: { [key: string]: any } | null;
  properties: { [key: string]: any } | null;
};
