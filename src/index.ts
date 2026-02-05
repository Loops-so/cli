#!/usr/bin/env node

import { Command, Option } from "commander";
import { config as loadDotenv } from "dotenv";
import { loops, setConfig } from "./client";
import { contactsCommand } from "./commands/contacts";
import { contactPropertiesCommand } from "./commands/contactProperties";
import { mailingListsCommand } from "./commands/mailingLists";
import { eventsCommand } from "./commands/events";

// putting LOOPS_ENDPOINT_URL in .env means that we need to load the file asap
let dotenvPath: string | undefined;

const dotenvArgIndex = process.argv.indexOf("--dotenv");
if (dotenvArgIndex !== -1 && process.argv[dotenvArgIndex + 1]) {
  dotenvPath = process.argv[dotenvArgIndex + 1];
} else {
  const dotenvEquals = process.argv.find((arg) => arg.startsWith("--dotenv="));
  if (dotenvEquals) {
    dotenvPath = dotenvEquals.split("=")[1];
  }
}

if (!dotenvPath) {
  dotenvPath = process.env.LOOPS_DOTENV;
}

if (dotenvPath) {
  loadDotenv({ path: dotenvPath, quiet: true });
}

const program = new Command();

program
  .name("loops")
  .description("CLI client for Loops API")
  .version("0.0.1")
  .addOption(
    new Option("--endpoint-url <url>", "API endpoint URL")
      .env("LOOPS_ENDPOINT_URL")
      .default("https://app.loops.so/api/"),
  )
  .addOption(
    new Option("--dotenv <path>", "Path to .env file").env("LOOPS_DOTENV"),
  )
  .hook("preAction", (cmd) => {
    const opts = cmd.optsWithGlobals();
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
program.addCommand(contactPropertiesCommand);
program.addCommand(mailingListsCommand);
program.addCommand(eventsCommand);

program.parse();
