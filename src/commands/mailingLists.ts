import { Command } from "commander";
import { loops } from "../client";

export const mailingListsCommand = new Command("mailing-lists").description(
  "Manage mailing lists",
);

mailingListsCommand
  .command("list")
  .description("List all mailing lists")
  .action(async () => {
    const client = loops();
    const resp = await client.getMailingLists();
    console.log(JSON.stringify(resp, null, 2));
  });
