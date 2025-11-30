export class TimestampHelper {
    static timestampToDate(ts: { seconds: number | string | bigint; nanos: number }): Date {
        const seconds = typeof ts.seconds === 'bigint'
            ? ts.seconds
            : BigInt(typeof ts.seconds === 'string' ? parseInt(ts.seconds, 10) : ts.seconds);

        const millis = seconds * 1000n + BigInt(Math.floor(ts.nanos / 1e6));
        const utcDate = new Date(Number(millis));

        return new Date(
            utcDate.getUTCFullYear(),
            utcDate.getUTCMonth(),
            utcDate.getUTCDate(),
            utcDate.getUTCHours(),
            utcDate.getUTCMinutes(),
            utcDate.getUTCSeconds(),
            utcDate.getUTCMilliseconds()
        );
    }

    static dateToTimestamp(date: Date): { seconds: bigint; nanos: number } {
        const utcMillis = Date.UTC(
            date.getFullYear(),
            date.getMonth(),
            date.getDate(),
            date.getHours(),
            date.getMinutes(),
            date.getSeconds(),
            date.getMilliseconds()
        );
        return {
            seconds: BigInt(Math.floor(utcMillis / 1000)),
            nanos: (date.getMilliseconds() % 1000) * 1_000_000
        };
    }
}
