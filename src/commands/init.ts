import { Args, Command } from "@oclif/core";
import fs from "fs";
import path from "path";
import os from 'os';

export default class Token extends Command {
  static summary = "Store configuration into a .gr4vyrc.json file";

  static description = `Generates a config file that can be used to generate the token.
`;
  static usage = "init acme sandbox private_key.pem";

  static args = {
    gr4vyId: Args.string({
      description:
        "The ID of your instance.",
      required: true
    }),

    environment: Args.string({
      description:
        "The environment of your instance.",
      required: true,
      options: ['production', 'sandbox']
    }),

    privateKey: Args.file({
      description:
        "The filename of the private key to add to the config.",
      required: true,
      parse: async (input) => {
        const file = path.resolve(input);
        return fs.readFileSync(file).toString()
      }
    }),
  };

  public async run(): Promise<void> {
    const { args } = await this.parse(Token);

    const file = path.resolve(os.homedir(), ".gr4vyrc.json");

    fs.writeFileSync(
      file,
      JSON.stringify({
        gr4vyId: args.gr4vyId,
        environment: args.environment,
        privateKey: args.privateKey,
      }, null, 2)
    );

    this.log("Successfully created .gr4vyrc.json file")
  }
}
