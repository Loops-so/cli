#!/usr/bin/env node

import { Command } from "commander";
import { config as loadDotenv } from "dotenv";
import { loops, setConfig } from "./client";
import { contactsCommand } from "./commands/contacts";

const program = new Command();

program
  .name("loops")
  .description("CLI client for Loops API")
  .version("0.0.1")
  .option(
    "--endpoint-url <url>",
    "API endpoint URL",
    "https://app.loops.so/api/",
  )
  .option("--dotenv <path>", "Path to .env file")
  .hook("preAction", (cmd) => {
    const opts = cmd.optsWithGlobals();

    if (opts.dotenv) {
      loadDotenv({ path: opts.dotenv, quiet: true });
    }

    setConfig({ endpointUrl: opts.endpointUrl });
  });

program
  .command("api-key")
  .description("Test API key")
  .action(async () => {
    const client = loops();
    const resp = await client.testApiKey();
    console.log(resp);
  });

program.addCommand(contactsCommand);

program
  .command("events")
  .description("Manage events")
  .action(() => {
    console.log("Events command - to be implemented");
  });

program.parse();
