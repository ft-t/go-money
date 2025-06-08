import { StringHelper } from '../../helpers/string.helper';

export enum Environment {
  Development = 0,
  Production = 1,
}

export class InitialConfiguration {
  public Environment: Environment;

  constructor(environment: string) {
    // @ts-ignore
    this.Environment = Environment[StringHelper.capitalize(environment.toLowerCase())] as any;
  }
}

export class Configuration extends InitialConfiguration {
  public readonly SiteHost!: string;
  // @ts-ignore
  public ServerHost: string;
  // @ts-ignore
  public UsersService: GrpcConfiguration;
}
