export class TimestampHelper {
    static timestampToDate(ts: { seconds: number | string | bigint; nanos: number }): Date {
        const seconds = typeof ts.seconds === 'bigint'
            ? ts.seconds
            : BigInt(typeof ts.seconds === 'string' ? parseInt(ts.seconds, 10) : ts.seconds);

        const millis = seconds * 1000n + BigInt(Math.floor(ts.nanos / 1e6));
        return new Date(Number(millis));
    }
}
