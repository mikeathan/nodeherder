const SERVICE_HOME: Record<string, string> = {
  'geocoding-api.open-meteo.com': 'https://open-meteo.com/en/docs/geocoding-api',
  'api.open-meteo.com': 'https://open-meteo.com/',
  'air-quality-api.open-meteo.com': 'https://open-meteo.com/en/docs/air-quality-api',
  'api.postcodes.io': 'https://postcodes.io/',
  'api.weather.gov': 'https://www.weather.gov/documentation/services-web-api',
  'data.hub.metoffice.gov.uk': 'https://datahub.metoffice.gov.uk/docs/getting-started',
  'www.fuel-finder.service.gov.uk': 'https://www.developer.fuel-finder.service.gov.uk/fuel-finder/public-api',
  'www.developer.fuel-finder.service.gov.uk': 'https://www.developer.fuel-finder.service.gov.uk/fuel-finder/public-api',
  'whenisbins.com': 'https://whenisbins.com/',
};

/** Resolve a raw provider API base to a human-facing home or documentation page. */
export function serviceHomeURL(raw: string): string {
  if (!raw) return '';
  try {
    const url = new URL(raw);
    return SERVICE_HOME[url.host] ?? `${url.protocol}//${url.host}/`;
  } catch {
    return raw;
  }
}
