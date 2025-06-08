// rabbit
import { CustomEnumError } from './custom-enum-error';

export class CustomError<T = any> {
  // @ts-ignore
  public ErrorCode: keyof CustomEnumError | number;
  // @ts-ignore
  public ErrorMessage: string;
  // @ts-ignore
  public OriginalError: string;
  public IsTranslationNeed = false;
  // @ts-ignore
  public Data: T;

  constructor(obj?: Partial<CustomError>) {
    if (obj) {
      Object.assign(this, obj);
    }
  }

  public toString(): string {
    return this.ErrorMessage;
  }
}
