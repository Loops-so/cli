import fs from "node:fs";
import os from "node:os";
import path from "node:path";

const CONFIG_DIR = path.join(os.homedir(), ".loops");
const CONFIG_PATH = path.join(CONFIG_DIR, "config.json");

type EndpointConfig = {
  apiKey: string;
};

type CliConfig = {
  endpoints: Record<string, EndpointConfig>;
  defaultEndpoint?: string;
};

function loadConfig(): CliConfig {
  try {
    const raw = fs.readFileSync(CONFIG_PATH, "utf8");
    const parsed = JSON.parse(raw) as Partial<CliConfig>;
    if (!parsed || typeof parsed !== "object") {
      return { endpoints: {} };
    }
    if (!parsed.endpoints || typeof parsed.endpoints !== "object") {
      parsed.endpoints = {};
    }
    return parsed as CliConfig;
  } catch {
    return { endpoints: {} };
  }
}

function saveConfig(config: CliConfig): void {
  if (!fs.existsSync(CONFIG_DIR)) {
    fs.mkdirSync(CONFIG_DIR, { recursive: true, mode: 0o700 });
  }

  const tmpPath = `${CONFIG_PATH}.tmp`;
  const contents = JSON.stringify(config, null, 2);

  fs.writeFileSync(tmpPath, contents, { mode: 0o600 });
  fs.renameSync(tmpPath, CONFIG_PATH);
  fs.chmodSync(CONFIG_PATH, 0o600);
}

export function setApiKeyForEndpoint(endpointUrl: string, apiKey: string): void {
  const cfg = loadConfig();

  if (!cfg.endpoints) {
    cfg.endpoints = {};
  }

  cfg.endpoints[endpointUrl] = { apiKey };

  if (!cfg.defaultEndpoint) {
    cfg.defaultEndpoint = endpointUrl;
  }

  saveConfig(cfg);
}

export function getApiKeyForEndpoint(endpointUrl?: string): string | null {
  const cfg = loadConfig();

  const effectiveEndpoint =
    endpointUrl || cfg.defaultEndpoint;

  if (!effectiveEndpoint) {
    return null;
  }

  const entry = cfg.endpoints[effectiveEndpoint];
  return entry?.apiKey ?? null;
}

export function getConfigPath(): string {
  return CONFIG_PATH;
}

