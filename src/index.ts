#!/usr/bin/env node

import { Command, Option } from "commander";
import { loops, setConfig } from "./client";
import { contactsCommand } from "./commands/contacts";
import { contactPropertiesCommand } from "./commands/contactProperties";
import { mailingListsCommand } from "./commands/mailingLists";
import { eventsCommand } from "./commands/events";
import { transactionalCommand } from "./commands/transactional";
import { loginCommand } from "./commands/login";

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
    new Option("--api-key <key>", "API key (overrides stored and env keys)"),
  )
  .hook("preAction", (cmd) => {
    const opts = cmd.optsWithGlobals();
    setConfig({ endpointUrl: opts.endpointUrl, apiKey: opts.apiKey });
  });

program
  .command("api-key")
  .description("Test API key")
  .action(async () => {
    const client = loops();
    const resp = await client.testApiKey();
    console.log(resp);
  });

program.addCommand(loginCommand);
program.addCommand(contactsCommand);
program.addCommand(contactPropertiesCommand);
program.addCommand(mailingListsCommand);
program.addCommand(eventsCommand);
program.addCommand(transactionalCommand);

program.parse();
