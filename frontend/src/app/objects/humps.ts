interface IHumpsOptions {
  separator?: string;
  split?: RegExp | string;
  process?: any;
}

type THumpsConvert = (key: string, options: IHumpsOptions) => string;

export class Humps {
  static toString = Object.prototype.toString;

  static processKeys<T = any, R = T>(convert: THumpsConvert, obj: any, options?: IHumpsOptions): R {
    if (!Humps.isObject(obj) || Humps.isDate(obj) || Humps.isRegExp(obj) || Humps.isBoolean(obj) || Humps.isFunction(obj)) {
      return obj;
    }

    let output;
    let i = 0;
    let l = 0;

    if (Humps.isArray(obj)) {
      output = [];

      for (l = obj.length; i < l; i++) {
        output.push(Humps.processKeys(convert, obj[i], options));
      }
    } else {
      output = {};

      for (const key in obj) {
        if (Object.prototype.hasOwnProperty.call(obj, key)) {
          // @ts-ignore
          output[convert(key, options)] = Humps.processKeys(convert, obj[key], options);
        }
      }
    }

    // @ts-ignore
    return output;
  }

  static separateWords(str: string, options: IHumpsOptions): string {
    options = options || {};

    const separator = options.separator || '_';
    const split = options.split || /(?=[A-Z])/;

    return str.split(split).join(separator);
  }

  static camelize(str: string): string {
    if (Humps.isNumerical(str)) {
      return str;
    }

    str = str.replace(/[\-_\s]+(.)?/g, (match, chr) =>
      chr ? chr.toUpperCase() : ''
    );

    // Ensure 1st char is always lowercase
    return str.substr(0, 1).toLowerCase() + str.substr(1);
  }

  static pascalize(str: string): string {
    const camelized = Humps.camelize(str);

    // Ensure 1st char is always uppercase
    return camelized.substr(0, 1).toUpperCase() + camelized.substr(1);
  }

  static depascalize(str: string, options?: IHumpsOptions): string {
    // @ts-ignore
    return Humps.separateWords(str, options).toLowerCase();
  }

  static isFunction(obj: any): boolean {
    return typeof (obj) === 'function';
  }

  static isObject(obj: any): boolean {
    return obj === Object(obj);
  }

  static isArray(obj: any): boolean {
    return toString.call(obj) === '[object Array]';
  }

  static isDate(obj: any): boolean {
    return toString.call(obj) === '[object Date]';
  }

  static isRegExp(obj: any): boolean {
    return toString.call(obj) === '[object RegExp]';
  }

  static isBoolean(obj: any): boolean {
    return toString.call(obj) === '[object Boolean]';
  }

  // Performant way to determine if obj coerces to a number
  static isNumerical(obj: any): boolean {
    obj = obj - 0;

    return obj === obj;
  }

  // Sets up function which handles processing keys
  // allowing the convert function to be modified by a callback
  static processor(convert: THumpsConvert, options?: IHumpsOptions): THumpsConvert {
    const callback = options && 'process' in options ? options.process : options;

    if (typeof (callback) !== 'function') {
      return convert;
    }

    return (str, options) => callback(str, convert, options);
  }

  static camelizeKeys<T = any, R = T>(object: T, options?: IHumpsOptions): R {
    return Humps.processKeys(Humps.processor(Humps.camelize, options), object);
  };

  static depascalizeKeys<T = any, R = T>(object: T, options?: IHumpsOptions): R {
    return Humps.processKeys(Humps.processor(Humps.depascalize, options), object, options);
  }

  static pascalizeKeys<T = any, R = T>(object: T, options?: IHumpsOptions): R {
    return Humps.processKeys(Humps.processor(Humps.pascalize, options), object);
  }
}
