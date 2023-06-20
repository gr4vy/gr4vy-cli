import fs = require("fs");
import os = require("os");
import path = require("path");
import { Command } from "@oclif/core";

const CONFIG_FILENAME = ".gr4vyrc.json";

export abstract class BaseCommand extends Command {
  public clientConfig: Record<string, string> = {};

  private load = () => {
    try {
      const file = path.resolve(os.homedir(), CONFIG_FILENAME);
      this.clientConfig = JSON.parse(String(fs.readFileSync(file)));
    } catch {
      this.error(`Could not load configuration file "${CONFIG_FILENAME}"`);
    }
  };

  public async init(): Promise<void> {
    await super.init();
    this.load();
  }
}
