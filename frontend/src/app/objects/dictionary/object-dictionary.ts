export interface IObjectDictionary<T> {
  [key: string]: T;
}

export class ObjectDictionary<T = any> {
  private dictionary: IObjectDictionary<T> = {};

  constructor(obj?: Partial<ObjectDictionary<T>>, objectDictionary?: IObjectDictionary<T>) {
    if (obj) {
      Object.assign(this, obj);
    }

    if (objectDictionary) {
      this.dictionary = {...this.dictionary, ...objectDictionary};
    }
  }

  public get length(): number {
    return this.keys().length;
  }

  public setValue(key: string, value: T): T {
    if (typeof key !== 'string') {
      console.warn('use only strings for keys');
    }

    this.dictionary[key] = value;

    return this.dictionary[key];
  }

  public getValue(key: string): T | null {
    return this.dictionary[key] || null;
  }

  public remove(key: string): void {
    if (this.dictionary[key]) {
      delete this.dictionary[key];
    }
  }

  /**
   * use carefully, low performance
   * @param value => delete by
   */
  public removeByValue(value: T): void {
    const stringifiedValue = JSON.stringify(value);

    for (const key in this.dictionary) {
      if (JSON.stringify(this.dictionary[key]) === stringifiedValue) {
        delete this.dictionary[key];
      }
    }
  }

  public removeByRef(value: T): void {
    for (const key in this.dictionary) {
      if (this.dictionary[key] === value) {
        delete this.dictionary[key];
      }
    }
  }

  public removeAll(): void {
    this.dictionary = {};
  }

  public values(): T[] {
    const values: T[] = [];

    for (const key in this.dictionary) {
      values.push(this.dictionary[key]);
    }

    return values;
  }

  public getFirst(): T {
    const values = this.values();

    // @ts-ignore
    return values.length ? values[0] : null;
  }

  public getLast(): T {
    const values = this.values();

    // @ts-ignore
    return values.length ? values[values.length - 1] : null;
  }

  public keys(): string[] {
    const keys: string[] = [];

    for (const key in this.dictionary) {
      keys.push(key);
    }

    return keys;
  }

  public entries(): Array<[string, T]> {
    const entries: Array<[string, T]> = [];

    for (const key in this.dictionary) {
      entries.push([key, this.dictionary[key]]);
    }

    return entries;
  }

  public getRawDictionary(): IObjectDictionary<T> {
    return this.dictionary;
  }
}
