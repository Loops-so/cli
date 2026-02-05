import { LoopsClient } from 'loops';

export class CustomLoopsClient extends LoopsClient {
  constructor(apiKey: string, endpointUrl?: string) {
    super(apiKey);

    if (endpointUrl) {
      this.apiRoot = endpointUrl.endsWith('/') ? endpointUrl : `${endpointUrl}/`;
    }
  }
}

let client: CustomLoopsClient | null = null;
let config: { endpointUrl?: string } | null = null;

export function setConfig(cfg: { endpointUrl?: string }): void {
  config = cfg;
}

export function loops(): CustomLoopsClient {
  if (!client) {
    const apiKey = process.env.LOOPS_API_KEY;
    if (!apiKey) {
      throw new Error('LOOPS_API_KEY environment variable is required');
    }
    client = new CustomLoopsClient(apiKey, config?.endpointUrl);
  }
  return client;
}
