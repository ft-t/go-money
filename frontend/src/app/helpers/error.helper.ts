import { ConnectError } from '@connectrpc/connect';

export class ErrorHelper {
    static GetMessage(error: any): string {
        if (error instanceof ConnectError) {
            return error.rawMessage;
        }

        return error.toString()
    }
}
