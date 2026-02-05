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
