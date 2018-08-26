import { IConnection } from "vscode-languageserver";

export namespace Logger {
    let conn: IConnection;

    export function init(connection: IConnection) {
        conn = connection;
    }

    export function info(message: string) {
        conn.console.info(message);
    }

    export function error(err: any) {
        let errMessage: string = '';

        if (err instanceof Error) {
            errMessage = `${err.message}\n${err.stack}`;
        } else {
            errMessage = err.toString();
        }

        conn.console.error(errMessage);
    }
}