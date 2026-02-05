#!/usr/bin/env node

import { Command } from "commander";
import { loops, setConfig } from "./client";

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

program
  .command("contacts")
  .description("Manage contacts")
  .action(() => {
    console.log("Contacts command - to be implemented");
  });

program
  .command("events")
  .description("Manage events")
  .action(() => {
    console.log("Events command - to be implemented");
  });

program.parse();
