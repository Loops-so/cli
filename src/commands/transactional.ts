import { Command } from "commander";
import { loops } from "../client";

export const transactionalCommand = new Command("transactional").description(
  "Manage transactional emails",
);

transactionalCommand
  .command("list")
  .description("List transactional email templates")
  .option("--per-page <number>", "Results per page (10-50)", "20")
  .option("--cursor <cursor>", "Pagination cursor")
  .action(async (options) => {
    const client = loops();

    const perPage = parseInt(options.perPage, 10);
    if (isNaN(perPage) || perPage < 10 || perPage > 50) {
      console.error("Error: --per-page must be a number between 10 and 50");
      process.exit(1);
    }

    const params: { perPage: number; cursor?: string } = { perPage };
    if (options.cursor) {
      params.cursor = options.cursor;
    }

    const resp = await client.getTransactionalEmails(params);
    console.log(JSON.stringify(resp, null, 2));
  });

transactionalCommand
  .command("send")
  .description("Send a transactional email")
  .requiredOption("--transactional-id <id>", "Transactional email ID")
  .requiredOption("-e, --email <email>", "Recipient email address")
  .option("--add-to-audience", "Create contact if doesn't exist", false)
  .option("--data-variables <json>", "Template variables as JSON")
  .option("--attachments <json>", "Attachments as JSON array")
  .action(async (options) => {
    const client = loops();

    const payload: {
      transactionalId: string;
      email: string;
      addToAudience?: boolean;
      dataVariables?: Record<string, unknown>;
      attachments?: Array<unknown>;
    } = {
      transactionalId: options.transactionalId,
      email: options.email,
    };

    if (options.addToAudience) {
      payload.addToAudience = true;
    }

    if (options.dataVariables) {
      try {
        payload.dataVariables = JSON.parse(options.dataVariables);
      } catch (error) {
        console.error("Error: Invalid JSON for --data-variables");
        process.exit(1);
      }
    }

    if (options.attachments) {
      try {
        const attachments = JSON.parse(options.attachments);
        if (!Array.isArray(attachments)) {
          console.error("Error: --attachments must be a JSON array");
          process.exit(1);
        }
        payload.attachments = attachments;
      } catch (error) {
        console.error("Error: Invalid JSON for --attachments");
        process.exit(1);
      }
    }

    const resp = await client.sendTransactionalEmail(payload);
    console.log(JSON.stringify(resp, null, 2));
  });
