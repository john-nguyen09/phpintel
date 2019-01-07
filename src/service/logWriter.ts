import "reflect-metadata";
import { IConnection } from "vscode-languageserver";
import { injectable, inject } from "inversify";

@injectable()
export class LogWriter {
    private conn: IConnection | undefined;

    constructor(@inject('IConnection') conn?: IConnection) {
        this.conn = conn;
    }

    info(message: string) {
        if (typeof this.conn !== 'undefined') {
            this.conn.console.info(message);
        } else {
            console.info(message);
        }
    }

    error(err: any) {
        let errMessage: string = '';

        if (err instanceof Error) {
            errMessage = `${err.message}\n${err.stack}`;
        } else if (err == null) {
            errMessage = 'Potential coding error, since error is null';
        } else {
            errMessage = err.toString();
        }

        if (typeof this.conn !== 'undefined') {
            this.conn.console.error(errMessage);
        } else {
            console.error(errMessage);
        }
    }
}