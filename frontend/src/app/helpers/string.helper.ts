export class StringHelper {
  static insertStr(origin: string, str: string, index: number): string {

    if (index > origin.length || index < 0 || StringHelper.isNullOrEmpty(origin)) {
      return '';
    }

    return origin.slice(0, index) + str + origin.slice(index);
  }

  static isNullOrEmpty(value: string) {
    if (value != null && typeof value === 'number') {
      return false;
    }

    return (value == null || (value === '') || value.length === 0);
  }

  static capitalize(input: string) {
    if (!input) {
      return input;
    }

    return input?.charAt(0).toUpperCase() + input.slice(1);
  }

  static replaceAll(value: string, search: string, replacement: string) {
    return value.replace(new RegExp(search, 'g'), replacement);
  }

  static splitCaseItemName(value: string, separator: string = '|'): string[] {
    let split = value.split(separator);

    if (split.length > 2) {
      const last = split.pop();
      // @ts-ignore
      split = [split.join(` ${separator} `), last];
    }

    return split;
  }

  static handleItemName(item: any): string[] {
    // Possibles names: ItemName
    const name = (item?.NameToken || item?.ItemName || item?.Hash || item.Group).split('|') || [];

    if (name.length > 1) {
      return name;
    }

    return [(item.NameToken ?? item.Hash ?? item.Group ?? item.ItemName)];
  }
}
