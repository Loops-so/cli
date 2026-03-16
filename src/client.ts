import { LoopsClient } from "loops";
import { getApiKeyForEndpoint } from "./config";

export class CustomLoopsClient extends LoopsClient {
  constructor(apiKey: string, endpointUrl?: string) {
    super(apiKey);

    if (endpointUrl) {
      this.apiRoot = endpointUrl.endsWith("/")
        ? endpointUrl
        : `${endpointUrl}/`;
    }
  }
}

type ClientConfig = {
  endpointUrl?: string;
  apiKey?: string;
};

let client: CustomLoopsClient | null = null;
let config: ClientConfig | null = null;

export function setConfig(cfg: ClientConfig): void {
  config = cfg;
}

export function loops(): CustomLoopsClient {
  if (!client) {
    const endpointUrl = config?.endpointUrl;

    const apiKeyFromOption = config?.apiKey;
    const apiKeyFromEnv = process.env.LOOPS_API_KEY;
    const apiKeyFromConfig = getApiKeyForEndpoint(endpointUrl);

    const apiKey = apiKeyFromOption || apiKeyFromEnv || apiKeyFromConfig;

    if (!apiKey) {
      throw new Error(
        "No API key found. Pass --api-key, set LOOPS_API_KEY, or run `loops login`.",
      );
    }

    client = new CustomLoopsClient(apiKey, endpointUrl);
  }
  return client;
}
