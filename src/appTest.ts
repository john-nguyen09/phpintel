import { Application, IApp } from "./app";
import { IConnection } from "vscode-languageserver";
import { interfaces } from "inversify";
import * as path from "path";

class TestApplication extends Application implements IApp {
    public run() { }
}

export namespace App {
    const app: IApp = new TestApplication();

    app.initStorage(path.join(__dirname, '..', 'debug', 'storage'));
    
    export function run() {
        app.run();
    }

    export function get<T>(
        identifier: string | symbol | interfaces.Newable<T> | interfaces.Abstract<T>
    ): T {
        return app.get(identifier);
    }
}
