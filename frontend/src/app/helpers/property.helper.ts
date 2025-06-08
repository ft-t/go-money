import { SimpleChanges } from '@angular/core';
import { BehaviorSubject, merge, Observable, of } from 'rxjs';
import { Humps } from '../objects/humps';
import { UntypedFormGroup } from '@angular/forms';
import { debounceTime, map, take } from 'rxjs/operators';

export class PropertyHelper {
  public static snakelizeObject(data: any): any {
    return Humps.depascalizeKeys(data);
  }

  public static snakelizeString(data: string): string {
    return Humps.depascalize(data);
  }

  public static pascalizeObject<T>(response: any): T {
    return Humps.pascalizeKeys(response);
  }

  public static pascalizeString(response: string): string {
    return Humps.pascalize(response);
  }

  public static getClassProps<T>(obj: T): string[] {
    const props: string[] = [];

    for (const prop in obj) {
      props.push(prop);
    }

    return props;
  }

  public static isFunction(fn: any): boolean {
    return typeof fn === 'function';
  }

  public static hasChanges(changes: SimpleChanges, propName: string) {
    // tslint:disable-next-line:triple-equals
    return changes.hasOwnProperty(propName) && changes[propName].previousValue != changes[propName].currentValue;
  }

  /*
   * flat and map object include arrays into following format:
   * {"0.data.prop": "1", "0.data.prop": "prop"}
   */
  // public static flatObjectWithPrefixes(obj: any, isArray: boolean = true): {[key: string]: string | number | null | undefined} {
  //   const getEntries = (obj: any, prefix = ''): any =>
  //     Object.entries(obj).flatMap(([prop, nestedObj]) => {
  //       if (nestedObj instanceof Array && nestedObj.length &&
  //         (typeof nestedObj[0] == 'string' || typeof nestedObj[0] == 'number' || typeof nestedObj[0] == 'boolean' ||
  //           typeof nestedObj[0] == 'bigint' || typeof nestedObj[0] == 'symbol' ||  typeof nestedObj[0] == null ||
  //           typeof nestedObj[0] == undefined)) {
  //         return [[`${prefix}${prop}`, nestedObj.join(', ')]];
  //       }
  //       return Object(nestedObj) === nestedObj
  //         ? getEntries(nestedObj, `${prefix}${prop}.`)
  //         : [[`${prefix}${prop}`, nestedObj]];
  //     });
  //
  //   const entries = getEntries(obj);
  //
  //   if (isArray) {
  //     return entries;
  //   }
  //
  //   // @ts-ignore
  //   const result = Object.fromEntries(getEntries(obj));
  //
  //   return result;
  // }

  /**
   *
   * @param obj - object which props are checking
   * @param complexPropName - name of property object should have
   * @param returnLastProp - if true returns prop value, if false returns object
   * @param addNonExistedProp - is need to add prop
   * @param valueOfLastProp - is the value of lat property
   * @param defaultValue - value to return if has no prop
   */
  public static getComplexProp(obj: any,
                               complexPropName: string,
                               returnLastProp: boolean = true,
                               addNonExistedProp: boolean = false,
                               valueOfLastProp?: any,
                               defaultValue: any = undefined): any {
    if (!obj || typeof obj !== 'object' || !Object.keys(obj).length) {
      return obj;
    }

    const split = complexPropName.split('.');
    let tempObj = obj;
    let prevProp: any = defaultValue;

    const splitLength = split.length;

    split.forEach((propName, index) => {
      const isLastIndex = index === splitLength - 1;
      const propExists = tempObj.hasOwnProperty(propName) || (tempObj.hasOwnProperty('hasAttribute') && tempObj.hasAttribute(propName));
      prevProp = index === 0 ? prevProp : tempObj;

      if (propExists) {
        if (isLastIndex && (addNonExistedProp || valueOfLastProp)) {
          tempObj[propName] = valueOfLastProp;
        }
        tempObj = tempObj[propName];
        return;
      }

      if (!addNonExistedProp) {
        tempObj = defaultValue;
        return;
      }

      tempObj[propName] = isLastIndex ? valueOfLastProp : {};
      tempObj = tempObj[propName];
    });

    return returnLastProp ? tempObj : prevProp;
  }

  /**
   * update or init BehaviourSubject context[subjectPropName] with context[inputPropName]
   * @param context - must contain inputPropName and subjectPropName
   * @param inputPropName - prop value that set to context[subjectPropName]
   * @param subjectPropName - name of BehaviourSubject that will be updated or initialized with context[inputPropName]
   * subjectPropName could be undefined, in that case it will be equal subjectPropName with '$' sign in the end
   */
  public static updateSubjectFromInputProp(context: any, inputPropName: string,
                                           subjectPropName: string = `${inputPropName}$`): void {
    PropertyHelper.checkUpdatedBehaviourSubjectProps(context, inputPropName, subjectPropName);

    if (!context[subjectPropName]) {
      (context[subjectPropName] as BehaviorSubject<any>) = new BehaviorSubject<any>(context[inputPropName]);
    }

    if (!(context[subjectPropName] as BehaviorSubject<any>).closed) {
      context[subjectPropName].next(context[inputPropName]);
      return;
    }

    throw new Error(`Subject: ${context[subjectPropName]} closed here`);
    // dangerous logic below, test before uncomment
    // const observers = (context[subjectPropName] as BehaviorSubject<any>).observers;
    // observers.forEach((observer) => observer.closed = false);
    // (context[subjectPropName] as BehaviorSubject<any>) = new BehaviorSubject<any>(context[inputPropName]);
    // (context[subjectPropName] as BehaviorSubject<any>).observers = observers;
  }

  /**
   * check updateSubjectFromInputProp
   * @param context
   * @param inputPropName
   * @param subjectPropName
   */
  public static checkUpdatedBehaviourSubjectProps(context: any,
                                                  inputPropName: string,
                                                  subjectPropName: string = `${inputPropName}$`): void | never {
    if (!context.hasOwnProperty(inputPropName)) {
      throw new Error(`inputPropName: "${inputPropName}" not exists in context: "${context}"`);
    }

    if (!context.hasOwnProperty(subjectPropName)) {
      throw new Error(`subjectPropName: "${subjectPropName}" not exists in context: "${context}"`);
    }

    if (!context[subjectPropName].subscribe) {
      throw new Error(`subjectPropName: "${subjectPropName}" must be instanceof Subject`);
    }
  }

  /**
   * check updateSubjectFromInputProp
   * @param changes
   * @param context
   * @param inputPropName
   * @param subjectPropName
   */
  public static updateSubjectFromInputPropFromChanges(changes: SimpleChanges,
                                                      context: any,
                                                      inputPropName: string,
                                                      subjectPropName: string = `${inputPropName}$`): boolean {
    PropertyHelper.checkUpdatedBehaviourSubjectProps(context, inputPropName, subjectPropName);

    if (!PropertyHelper.hasChanges(changes, inputPropName)) {
      return false;
    }

    PropertyHelper.updateSubjectFromInputProp(context, inputPropName, subjectPropName);

    return true;
  }

  /**
   * delete or set equal properties to specific value
   * @param obj1 - old object
   * @param obj2 - new object that will be mutated
   * @param deleteSameProps - will delete equal props from obj2
   * @param valueToSet - will set equal props to value in obj2
   * @param ignoreProps
   */
  static setEqualPropertiesToValue(obj1: object,
                                   obj2: object,
                                   deleteSameProps: boolean = true,
                                   valueToSet: any = null,
                                   ignoreProps: string[] = []): void {
    const obj2Keys = Object.keys(obj2);

    obj2Keys.forEach((obj2Key) => {
      // @ts-ignore
      if (!obj1.hasOwnProperty(obj2Key) || obj1[obj2Key] !== obj2[obj2Key] || ignoreProps.includes(obj2Key)) {
        return;
      }

      if (deleteSameProps) {
        // @ts-ignore
        delete obj2[obj2Key];

        return;
      }

      // @ts-ignore
      obj2[obj2Key] = valueToSet;
    });
  }

  static isAllFormControlsPristine(formGroup: UntypedFormGroup): Observable<boolean> {
    return merge(of(true).pipe(take(1)), formGroup.valueChanges, formGroup.statusChanges).pipe(
      debounceTime(100),
      map(() => {
        for (const controlKey in formGroup.controls) {
          const control = formGroup.controls[controlKey];

          if (!control.pristine) {
            return false;
          }
        }

        return true;
      })
    );
  }

  static shuffleArrayWithAlgorithm<T>(arr: T[]): T[] {
    // case 0: % 2

    // max i = 158
    // max div = 15

    // const divider = i > 15 ? i - ((Math.floor(i / 15))) * 15 : i;
    //
    // arr.forEach((item, index) => {
    //   if(index % divider == 0){
    //     array.push(item)
    //   }
    // })

    const array = [...arr];

    for (let i = array.length - 1; i > 0; i--) {
      const j = i > 15 ? i - ((Math.floor(i / 15))) * 15 : i;
      [array[i], array[j]] = [array[j], array[i]];
    }

    return array;
  }

  static shuffleArray<T>(arr: T[]): T[] {
    const array = [...arr];

    for (let i = array.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [array[i], array[j]] = [array[j], array[i]];
    }

    return array;
  }

  static getRandomInt(min: number, max: number) {
    min = Math.ceil(min);
    max = Math.floor(max);

    return Math.floor(Math.random() * (max - min)) + min;
  }
}
