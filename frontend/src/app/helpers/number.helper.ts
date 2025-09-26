
export class NumberHelper {
    static toNegativeNumber(value: string | undefined): string | undefined {
        if (!value) return value;

        let num = parseFloat(value);
        if (!num) return value;

        return (-Math.abs(num)).toString();
    }

    static toPositiveNumber(value: string | undefined): string | undefined {
        if (!value) return value;

        let num = parseFloat(value);
        if (!num) return value;

        return Math.abs(num).toString();
    }
}
