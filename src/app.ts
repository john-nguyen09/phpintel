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
import { ReferenceTable } from "./storage/table/referenceTable";
import { PhpDocumentTable } from "./storage/table/phpDoc";
import { IConnection } from "vscode-languageserver";

export class Application {
    protected container: Container = new Container();

    constructor(storageLocation: string, connection?: IConnection) {
        this.container.bind<IConnection | undefined>('IConnection').toConstantValue(connection);

        this.initBind();
        this.initStorage(storageLocation);
    }

    public getContainer(): Container {
        return this.container;
    }

    protected initStorage(location: string) {
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
}
