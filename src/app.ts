import "reflect-metadata";
import { LevelDatasource, DbStore } from "./storage/db";
import { IConnection, createConnection, Hover } from "vscode-languageserver";
import { LogWriter } from "./service/logWriter";
import { Hasher } from "./service/hasher";
import { InitializeProvider } from "./provider/initialize";
import { HoverProvider } from "./provider/hover";
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

export namespace App {
    let container: Container = new Container();

    export function run() {
        let connection = createConnection();

        container.bind<IConnection>('IConnection').toConstantValue(connection);
        beforeListen();

        connection.onInitialize(InitializeProvider.provide);
        connection.onHover(HoverProvider.provide);
        connection.listen();
    }

    export function setUpForTest() {
        container.bind<IConnection | undefined>('IConnection').toConstantValue(undefined);
        beforeListen();
        App.initStorage(path.join(__dirname, '..', 'debug', 'storage'));
    }

    export function initStorage(location: string) {
        let datasource = new LevelDatasource(location);

        container.bind<LevelDatasource>(LevelDatasource).toConstantValue(datasource);

        // Tables
        container.bind<ClassTable>(ClassTable).toSelf();
        container.bind<ClassConstantTable>(ClassConstantTable).toSelf();
        container.bind<ConstantTable>(ConstantTable).toSelf();
        container.bind<FunctionTable>(FunctionTable).toSelf();
        container.bind<MethodTable>(MethodTable).toSelf();
        container.bind<PropertyTable>(PropertyTable).toSelf();
        container.bind<PhpDocumentTable>(PhpDocumentTable).toSelf();
        container.bind<ReferenceTable>(ReferenceTable).toSelf();
    }

    export function get<T>(
        identifier: string | symbol | interfaces.Newable<T> | interfaces.Abstract<T>
    ): T {
        return container.get(identifier);
    }

    function beforeListen() {
        container.bind<LogWriter>(LogWriter).toSelf().inSingletonScope();
        container.bind<Hasher>(Hasher).toSelf().inSingletonScope();
        container.bind<TextDocumentStore>(TextDocumentStore).toSelf().inSingletonScope();
        container.bind<Indexer>(Indexer).toSelf().inSingletonScope();
        container.bind<Traverser>(Traverser).toSelf().inSingletonScope();
    }
}
