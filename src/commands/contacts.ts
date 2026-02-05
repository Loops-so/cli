import { Command } from "commander";
import { loops } from "../client";

export const contactsCommand = new Command("contacts").description(
  "Manage contacts",
);

contactsCommand
  .command("find")
  .description("Find a contact by email or user ID")
  .option("-e, --email <email>", "Find by email address")
  .option("-u, --user-id <userId>", "Find by user ID")
  .action(async (options) => {
    const client = loops();

    if (!options.email && !options.userId) {
      console.error("Error: Either --email or --user-id is required");
      process.exit(1);
    }

    if (options.email && options.userId) {
      console.error("Error: Cannot specify both --email and --user-id");
      process.exit(1);
    }

    const query = options.email
      ? { email: options.email }
      : { userId: options.userId };

    const resp = await client.findContact(query);
    console.log(JSON.stringify(resp, null, 2));
  });

contactsCommand
  .command("create")
  .description("Create a new contact")
  .requiredOption("-e, --email <email>", "Contact email address")
  .option("--properties <json>", "Contact properties as JSON")
  .option("--mailing-lists <json>", "Mailing list subscriptions as JSON")
  .action(async (options) => {
    const client = loops();

    const payload: {
      email: string;
      properties?: Record<string, unknown>;
      mailingLists?: Record<string, boolean>;
    } = {
      email: options.email,
    };

    if (options.properties) {
      try {
        payload.properties = JSON.parse(options.properties);
      } catch (error) {
        console.error("Error: Invalid JSON for --properties");
        process.exit(1);
      }
    }

    if (options.mailingLists) {
      try {
        payload.mailingLists = JSON.parse(options.mailingLists);
      } catch (error) {
        console.error("Error: Invalid JSON for --mailing-lists");
        process.exit(1);
      }
    }

    const resp = await client.createContact(payload);
    console.log(JSON.stringify(resp, null, 2));
  });

contactsCommand
  .command("delete")
  .description("Delete a contact by email or user ID")
  .option("-e, --email <email>", "Delete by email address")
  .option("-u, --user-id <userId>", "Delete by user ID")
  .action(async (options) => {
    const client = loops();

    if (!options.email && !options.userId) {
      console.error("Error: Either --email or --user-id is required");
      process.exit(1);
    }

    if (options.email && options.userId) {
      console.error("Error: Cannot specify both --email and --user-id");
      process.exit(1);
    }

    const query = options.email
      ? { email: options.email }
      : { userId: options.userId };

    const resp = await client.deleteContact(query);
    console.log(JSON.stringify(resp, null, 2));
  });

contactsCommand
  .command("update")
  .description("Update a contact by email or user ID")
  .option("-e, --email <email>", "Update by email address")
  .option("-u, --user-id <userId>", "Update by user ID")
  .option("--properties <json>", "Contact properties as JSON")
  .option("--mailing-lists <json>", "Mailing list subscriptions as JSON")
  .action(async (options) => {
    const client = loops();

    if (!options.email && !options.userId) {
      console.error("Error: Either --email or --user-id is required");
      process.exit(1);
    }

    if (options.email && options.userId) {
      console.error("Error: Cannot specify both --email and --user-id");
      process.exit(1);
    }

    const payload: {
      email?: string;
      userId?: string;
      properties?: Record<string, unknown>;
      mailingLists?: Record<string, boolean>;
    } = {};

    if (options.email) {
      payload.email = options.email;
    }

    if (options.userId) {
      payload.userId = options.userId;
    }

    if (options.properties) {
      try {
        payload.properties = JSON.parse(options.properties);
      } catch (error) {
        console.error("Error: Invalid JSON for --properties");
        process.exit(1);
      }
    }

    if (options.mailingLists) {
      try {
        payload.mailingLists = JSON.parse(options.mailingLists);
      } catch (error) {
        console.error("Error: Invalid JSON for --mailing-lists");
        process.exit(1);
      }
    }

    const resp = await client.updateContact(payload);
    console.log(JSON.stringify(resp, null, 2));
  });
