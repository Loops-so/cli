import { Command } from "commander";
import { loops } from "../client";

export const contactPropertiesCommand = new Command(
  "contact-properties",
).description("Manage contact properties");

contactPropertiesCommand
  .command("list")
  .description("List contact properties")
  .option("--filter <type>", "Filter properties: all or custom", "all")
  .action(async (options) => {
    const client = loops();

    if (options.filter !== "all" && options.filter !== "custom") {
      console.error('Error: --filter must be either "all" or "custom"');
      process.exit(1);
    }

    const resp = await client.getCustomProperties(options.filter);
    console.log(JSON.stringify(resp, null, 2));
  });

contactPropertiesCommand
  .command("create")
  .description("Create a custom contact property")
  .requiredOption("-n, --name <name>", "Property name (camelCase)")
  .requiredOption(
    "-t, --type <type>",
    "Property type: string, number, boolean, or date",
  )
  .action(async (options) => {
    const client = loops();

    const validTypes = ["string", "number", "boolean", "date"];
    if (!validTypes.includes(options.type)) {
      console.error(
        `Error: --type must be one of: ${validTypes.join(", ")}`,
      );
      process.exit(1);
    }

    const resp = await client.createContactProperty(
      options.name,
      options.type,
    );
    console.log(JSON.stringify(resp, null, 2));
  });
