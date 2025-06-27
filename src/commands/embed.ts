import { getEmbedToken } from "@gr4vy/sdk";
import { Args, Flags } from "@oclif/core";
import { BaseCommand } from "../base";
import { decodeJWT } from "../helpers/decode";
import { parseEmbedParams } from "../helpers/parse-embed-data";

export default class Token extends BaseCommand {
  static summary = "Generate a token for use with Gr4vy Embed.";

  static description = `This token can be used with Embed as it is
restricted to frontend scopes only.

It accepts any number of key=value pairs as additional data to be 
pinned in the token.
`;
  static usage = "embed 1299 USD buyer_external_identifier=user-123";

  static strict = false;

  static args = {
    amount: Args.integer({
      description:
        "The amount to generate a token for. This amount needs to be in the smallest denomination for the currency, e.g. 1299 for $12.99",
      required: true,
    }),

    currency: Args.string({
      description: "The 3 digit currency code to generate a token for.",
      example: "USD",
      required: true,
    }),
  };

  static flags = {
    debug: Flags.boolean({
      summary: "Returns the raw header and claim for the token",
      description:
        "Returns the decoded header and claim from the JWT token without the signature",
    }),
  };

  public async run(): Promise<void> {
    const { flags, args } = await this.parse(Token);
    const embedParams = parseEmbedParams(args.amount, args.currency, this.argv);

    const token = await getEmbedToken({
      privateKey: this.clientConfig.privateKey,
      embedParams
    });

    if (flags.debug) {
      const data = decodeJWT(token);
      this.log(JSON.stringify(data, null, 2));
    } else {
      this.log(token);
    }
  }
}
