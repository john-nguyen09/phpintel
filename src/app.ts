import "reflect-metadata";
import { LevelDatasource } from "./storage/db";
import { LogWriter } from "./service/logWriter";
import { Hasher } from "./service/hasher";
import { Container, interfaces } from "inversify";
import { Indexer } from "./index/indexer";
import { ClassTable } from "./storage/table/class";
import { ClassConstantTable } from "./storage/table/classConstant";
import { ConstantTable } from "./storage/table/constant";
import { FunctionTable } from "./storage/table/function";
import { MethodTable } from "./storage/table/method";
import { PropertyTable } from "./storage/table/property";
import { Traverser } from "./traverser";
import { ReferenceTable } from "./storage/table/reference";
import { PhpDocumentTable } from "./storage/table/phpDoc";
import { IConnection } from "vscode-languageserver";
import { ScopeVarTable } from "./storage/table/scopeVar";

export interface AppOptions {
    storage: string;
}

export class Application {
    protected container: Container = new Container();
    protected options: AppOptions = {
        storage: ''
    };

    constructor(storageLocation: string, connection?: IConnection) {
        this.container.bind<IConnection | undefined>('IConnection').toConstantValue(connection);

        this.initBind();
        this.initStorage(storageLocation);
    }

    public getContainer(): Container {
        return this.container;
    }

    public async clearCache() {
        const db = this.container.get<LevelDatasource>(LevelDatasource).getDb();
        let promises: Promise<void>[] = [];

        return new Promise<void>((resolve, reject) => {
            db.createKeyStream()
                .on('data', (key) => {
                    promises.push(db.del(key));
                })
                .on('end', () => {
                    Promise.all(promises).then(() => { resolve() });
                })
                .on('error', (err: Error) => {
                    reject(err);
                });
        });
    }

    public async shutdown() {
        await this.container.get<LevelDatasource>(LevelDatasource).getDb().close();
    }

    protected initStorage(location: string) {
        this.options.storage = location;

        let datasource = new LevelDatasource(location);
        this.container.bind<LevelDatasource>(LevelDatasource).toConstantValue(datasource);

        // Tables
        this.container.bind<ClassTable>(ClassTable).toSelf().inSingletonScope();
        this.container.bind<ClassConstantTable>(ClassConstantTable).toSelf().inSingletonScope();
        this.container.bind<ConstantTable>(ConstantTable).toSelf().inSingletonScope();
        this.container.bind<FunctionTable>(FunctionTable).toSelf().inSingletonScope();
        this.container.bind<MethodTable>(MethodTable).toSelf().inSingletonScope();
        this.container.bind<PropertyTable>(PropertyTable).toSelf().inSingletonScope();
        this.container.bind<PhpDocumentTable>(PhpDocumentTable).toSelf().inSingletonScope();
        this.container.bind<ReferenceTable>(ReferenceTable).toSelf().inSingletonScope();
        this.container.bind<ScopeVarTable>(ScopeVarTable).toSelf().inSingletonScope();
    }

    protected initBind() {
        this.container.bind<LogWriter>(LogWriter).toSelf().inSingletonScope();
        this.container.bind<Hasher>(Hasher).toSelf().inSingletonScope();
        this.container.bind<Indexer>(Indexer).toSelf().inSingletonScope();
        this.container.bind<Traverser>(Traverser).toSelf().inSingletonScope();
    }
}

export namespace App {
    let app: Application;

    export function init(storageLocation: string, connection?: IConnection) {
        app = new Application(storageLocation, connection);
    }

    export function get<T>(
        identifier: string | symbol | interfaces.Newable<T> | interfaces.Abstract<T>
    ): T {
        return app.getContainer().get(identifier);
    }

    export function set<T>(
        identifier: string,
        value: T
    ): void {
        app.getContainer().bind<T>(identifier).toConstantValue(value);
    }

    export async function clearCache() {
        return await app.clearCache();
    }

    export async function shutdown() {
        return await app.shutdown();
    }
}
