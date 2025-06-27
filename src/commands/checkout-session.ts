import { Gr4vy, withToken } from "@gr4vy/sdk";
import { Flags } from "@oclif/core";
import { BaseCommand } from "../base";
import { decodeJWT } from "../helpers/decode";

export default class CheckoutSession extends BaseCommand {
  static summary = "Generate a checkout session ID.";

  static description = `This ID can be used with Secure Fields and our native mobile SDKs.
`;
  static usage = "checkout-session --merchant-account-id=default";

  static strict = false;

  static flags = {
     mid: Flags.string({
      char: "m",
      summary: "The merchant account ID to generate a checkout session for.",
      description:  "This default to no value if not provided.",
    }),
    debug: Flags.boolean({
      summary: "Returns the raw header and claim for the token",
      description:
        "Returns the decoded header and claim from the JWT token without the signature",
    }),
  };

  public async run(): Promise<void> {
    const { flags } = await this.parse(CheckoutSession);

    const gr4vy = new Gr4vy({
        merchantAccountId: flags.mid as string ?? undefined,
        server: this.clientConfig.environment == "production" ? "production" : "sandbox",
        id: this.clientConfig.gr4vyId,
        bearerAuth: withToken({
            privateKey: this.clientConfig.privateKey,
        }),
    });

    const checkoutSession = await gr4vy.checkoutSessions.create();

    if (flags.debug) {
      this.log(JSON.stringify(checkoutSession, null, 2));
    } else {
      this.log(checkoutSession.id);
    }
  }
}
