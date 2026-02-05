#!/usr/bin/env node

import { Command } from "commander";

const program = new Command();

program.name("loops").description("CLI client for Loops API").version("0.0.1");

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
