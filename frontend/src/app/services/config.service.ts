import { Injectable } from '@angular/core';
import { Configuration, Environment, InitialConfiguration } from '../objects/configuration/сonfiguration';

@Injectable()
export class ConfigService {
  private readonly CONFIG: Configuration;

  constructor(
    initial: InitialConfiguration
  ) {
    let tempConfigName = null;

    let environment = initial.Environment;

    switch (environment) {
      case Environment.Development :
        // @ts-ignore
        tempConfigName = require('../configuration/dev.json');
        break;
      case Environment.Production :
        // @ts-ignore
        tempConfigName = require('../configuration/prod.json');
        break;
      default:
        tempConfigName = '../configuration/common.json';
        console.error(`Can not find specified config file. ${environment}`);
        break;
    }

    const common = tempConfigName;

    const conf = common as Configuration;

    conf.Environment = environment;

    this.CONFIG = common;
  }

  public getConfig(): Configuration {
    return this.CONFIG;
  }
}
