import { IConnection } from "vscode-languageserver";
import { injectable, inject } from "inversify";
import { BindingIdentifier } from "../constant/bindingIdentifier";

@injectable()
export class Messenger {
    private conn: IConnection;

    constructor(@inject(BindingIdentifier.CONNECTION) conn: IConnection) {
        this.conn = conn;
    }

    info(message: string) {
        this.conn.console.info(message);
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

        this.conn.console.error(errMessage);
    }
}