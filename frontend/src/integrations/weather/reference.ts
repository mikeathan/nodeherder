/** Stable weather source identity; values remain in the integration resource. */
export interface WeatherReference {
  kind: 'weather';
  instance_id: string;
  resource_id: string;
}
