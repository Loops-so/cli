import { Command } from "commander";
import { loops } from "../client";

export const eventsCommand = new Command("events").description(
  "Manage events",
);

eventsCommand
  .command("send")
  .description("Send an event")
  .requiredOption("--event-name <name>", "Event name")
  .option("-e, --email <email>", "Contact email address")
  .option("-u, --user-id <userId>", "Contact user ID")
  .option("--contact-properties <json>", "Contact properties as JSON")
  .option("--event-properties <json>", "Event properties as JSON")
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
      eventName: string;
      contactProperties?: Record<string, unknown>;
      eventProperties?: Record<string, unknown>;
      mailingLists?: Record<string, boolean>;
    } = {
      eventName: options.eventName,
    };

    if (options.email) {
      payload.email = options.email;
    }

    if (options.userId) {
      payload.userId = options.userId;
    }

    if (options.contactProperties) {
      try {
        payload.contactProperties = JSON.parse(options.contactProperties);
      } catch (error) {
        console.error("Error: Invalid JSON for --contact-properties");
        process.exit(1);
      }
    }

    if (options.eventProperties) {
      try {
        payload.eventProperties = JSON.parse(options.eventProperties);
      } catch (error) {
        console.error("Error: Invalid JSON for --event-properties");
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

    const resp = await client.sendEvent(payload);
    console.log(JSON.stringify(resp, null, 2));
  });
