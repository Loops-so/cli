import { Command } from "commander";
import readline from "node:readline";
import { setApiKeyForEndpoint } from "../config";

export const loginCommand = new Command("login")
  .description(
    "Authenticate with Loops by storing an API key in a local config file (chmod 600).",
  )
  .option(
    "--endpoint-url <url>",
    "API endpoint URL to associate with this key (defaults to current global endpoint URL)",
  )
  .action(async (options, command) => {
    const parent = command.parent;
    const parentOpts =
      typeof parent?.optsWithGlobals === "function"
        ? parent.optsWithGlobals()
        : {};

    const endpointUrl: string =
      options.endpointUrl ||
      parentOpts.endpointUrl ||
      process.env.LOOPS_ENDPOINT_URL ||
      "https://app.loops.so/api/";

    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
    });

    const apiKey: string = await new Promise((resolve) => {
      rl.question("Enter your Loops API key: ", (answer) => {
        rl.close();
        resolve(answer.trim());
      });
    });

    if (!apiKey) {
      console.error("No API key provided.");
      process.exitCode = 1;
      return;
    }

    setApiKeyForEndpoint(endpointUrl, apiKey);
    // The config module enforces chmod 600 on the config file.
    console.log(
      `API key saved for endpoint ${endpointUrl}. You can override it with the LOOPS_API_KEY environment variable if needed.`,
    );
  });

