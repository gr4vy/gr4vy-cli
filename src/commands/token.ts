import { Client, JWTScope } from "@gr4vy/node";
import { Flags } from "@oclif/core";
import { BaseCommand } from "../base";
import { decodeJWT } from "../helpers/decode";

export default class Token extends BaseCommand {
  static summary = "Generate a bearer token for server-to-server API calls.";

  static description = `This token should be used with care as it is not
restricted to any specific frontend scopes only.
`;
  static usage = "token expiresIn=10d --scope=buyers.read --scope=buyers.write";

  static flags = {
    expiresIn: Flags.string({
      char: "e",
      summary: "The expiry of the token",
      description:
        "The expiration expressed in seconds or a string describing a time span vercel/ms.",
      multiple: false,
      default: "1h",
      required: false,
    }),

    scope: Flags.string({
      char: "s",
      summary: "A scope to add to this flag",
      description: "A single scope to add to this JWT",
      multiple: true,
      options: [
        "all.read",
        "all.write",
        "*.read",
        "*.write",
        "anti-fraud-service-definitions.read",
        "anti-fraud-service-definitions.write",
        "anti-fraud-services.read",
        "anti-fraud-services.write",
        "buyers.read",
        "buyers.write",
        "buyers.billing-details.read",
        "buyers.billing-details.write",
        "connections.read",
        "connections.write",
        "digital-wallets.read",
        "digital-wallets.write",
        "flows.read",
        "flows.write",
        "payment-methods.read",
        "payment-methods.write",
        "payment-options.read",
        "payment-options.write",
        "payment-service-definitions.read",
        "payment-service-definitions.write",
        "payment-services.read",
        "payment-services.write",
        "reports.read",
        "reports.write",
        "roles.read",
        "roles.write",
        "transactions.read",
        "transactions.write",
        "audit-logs.read",
        "audit-logs.write",
        "checkout-sessions.read",
        "checkout-sessions.write",
        "card-scheme-definitions.read",
        "card-scheme-definitions.write",
        "payment-method-definitions.read",
        "payment-method-definitions.write",
        "reset.read",
        "reset.write",
        "merchant-accounts.read",
        "merchant-accounts.write",
      ],
      parse: async (input) => input.replace("all.", "*."),
      default: ["*.read", "*.write"],
      required: false,
    }),

    debug: Flags.boolean({
      summary: "Returns the raw header and claim for the token",
      description:
        "Returns the decoded header and claim from the JWT token without the signature",
    }),
  };

  public async run(): Promise<void> {
    const { flags } = await this.parse(Token);

    const client = new Client(this.clientConfig as any);
    const token = await client.getBearerToken(
      flags.scope as JWTScope[],
      flags.expiresIn
    );

    if (flags.debug) {
      const data = decodeJWT(token);
      this.log(JSON.stringify(data, null, 2));
    } else {
      this.log(token);
    }
  }
}
