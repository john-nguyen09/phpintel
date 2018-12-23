import "reflect-metadata";
import { LevelDatasource, DbStore } from "./storage/db";
import { IConnection, createConnection, Hover } from "vscode-languageserver";
import { LogWriter } from "./service/logWriter";
import { Hasher } from "./service/hasher";
import { Container, interfaces } from "inversify";
import { Indexer } from "./index/indexer";
import { TextDocumentStore } from "./textDocumentStore";
import { ClassTable } from "./storage/table/class";
import { ClassConstantTable } from "./storage/table/classConstant";
import { ConstantTable } from "./storage/table/constant";
import { FunctionTable } from "./storage/table/function";
import { MethodTable } from "./storage/table/method";
import { PropertyTable } from "./storage/table/property";
import { Traverser } from "./traverser";
import { ReferenceTable } from "./storage/table/referenceTable";
import * as path from "path";
import { PhpDocumentTable } from "./storage/table/phpDoc";
import { InitializeProvider } from "./provider/initialize";
import { HoverProvider } from "./provider/hover";

export interface IApp {
    run(): void;
    initStorage(location: string): void;
    get<T>(
        identifier: string | symbol | interfaces.Newable<T> | interfaces.Abstract<T>
    ): T;
}

export abstract class Application {
    protected container: Container = new Container();

    constructor() {
        this.initBind();
    }

    public initStorage(location: string) {
        let datasource = new LevelDatasource(location);
        this.container.bind<LevelDatasource>(LevelDatasource).toConstantValue(datasource);

        // Tables
        this.container.bind<ClassTable>(ClassTable).toSelf();
        this.container.bind<ClassConstantTable>(ClassConstantTable).toSelf();
        this.container.bind<ConstantTable>(ConstantTable).toSelf();
        this.container.bind<FunctionTable>(FunctionTable).toSelf();
        this.container.bind<MethodTable>(MethodTable).toSelf();
        this.container.bind<PropertyTable>(PropertyTable).toSelf();
        this.container.bind<PhpDocumentTable>(PhpDocumentTable).toSelf();
        this.container.bind<ReferenceTable>(ReferenceTable).toSelf();
    }

    public get<T>(
        identifier: string | symbol | interfaces.Newable<T> | interfaces.Abstract<T>
    ): T {
        return this.container.get(identifier);
    }

    protected initBind() {
        this.container.bind<LogWriter>(LogWriter).toSelf().inSingletonScope();
        this.container.bind<Hasher>(Hasher).toSelf().inSingletonScope();
        this.container.bind<TextDocumentStore>(TextDocumentStore).toSelf().inSingletonScope();
        this.container.bind<Indexer>(Indexer).toSelf().inSingletonScope();
        this.container.bind<Traverser>(Traverser).toSelf().inSingletonScope();
    }
}

class LspApplication extends Application implements IApp {
    private connection: IConnection;

    constructor() {
        super();

        this.connection = createConnection();
        this.container.bind<IConnection>('IConnection').toConstantValue(this.connection);
    }

    public run() {
        this.connection.listen();
    }
}

export namespace App {
    const app: IApp = new LspApplication();

    export function run(initProviders: (connection: IConnection) => void) {
        initProviders(app.get<IConnection>('IConnection'));
        app.run();
    }

    export function getApp() {
        return app;
    }

    export function get<T>(
        identifier: string | symbol | interfaces.Newable<T> | interfaces.Abstract<T>
    ): T {
        return app.get(identifier);
    }
}
