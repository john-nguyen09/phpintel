import "reflect-metadata";
import { IdentifierIndex } from "./index/identifierIndex";
import { PositionIndex } from "./index/positionIndex";
import { TimestampIndex } from "./index/timestampIndex";
import { LevelDatasource, DbStore } from "./storage/db";
import { UriIndex } from "./index/uriIndex";
import { IConnection, createConnection, Hover } from "vscode-languageserver";
import { LogWriter } from "./service/logWriter";
import { Hasher } from "./service/hasher";
import { InitializeProvider } from "./provider/initialize";
import { HoverProvider } from "./provider/hover";
import { Container } from "inversify";
import { IndexNames, IndexId, IndexVersion } from "./constant/indexId";
import { TreeTraverser } from "./treeTraverser/structures";
import { TreeNode } from "./util/parseTree";
import { RecursiveTraverser } from "./treeTraverser/recursive";
import { Indexer } from "./index/indexer";
import { TextDocumentStore } from "./textDocumentStore";
import { BindingIdentifier } from "./constant/bindingIdentifier";

export namespace Application {
    let container: Container = new Container();

    export function run() {
        let connection = createConnection();

        container.bind<IConnection>(BindingIdentifier.CONNECTION).toConstantValue(connection);
        beforeListen();

        connection.onInitialize(InitializeProvider.provide);
        connection.onHover(HoverProvider.provide);
        connection.listen();
    }

    export function initStorage(location: string) {
        let datasource = new LevelDatasource(location, {
            valueEncoding: 'json'
        });

        container.bind<LevelDatasource>(BindingIdentifier.DATASOURCE)
            .toConstantValue(datasource);

        IndexNames.alternatives.map((indexName) => {
            container.bind<DbStore>(BindingIdentifier.DB_STORE)
                .toConstantValue(new DbStore(datasource, {
                    name: IndexId[indexName.value],
                    version: IndexVersion[indexName.value]
                }))
                .whenTargetNamed(IndexId[indexName.value]);
        });

        container.bind<IdentifierIndex>(BindingIdentifier.IDENTIFIER_INDEX)
            .to(IdentifierIndex)
            .inSingletonScope();
        container.bind<PositionIndex>(BindingIdentifier.POSITION_INDEX)
            .to(PositionIndex)
            .inSingletonScope();
        container.bind<TimestampIndex>(BindingIdentifier.TIMESTAMP_INDEX)
            .to(TimestampIndex)
            .inSingletonScope();
        container.bind<UriIndex>(BindingIdentifier.URI_INDEX)
            .to(UriIndex)
            .inSingletonScope();
    }

    export function get<T>(identifier: string): T {
        return container.get(identifier);
    }

    function beforeListen() {
        container.bind<LogWriter>(BindingIdentifier.MESSENGER)
            .to(LogWriter)
            .inSingletonScope();
        container.bind<Hasher>(BindingIdentifier.HASHER)
            .to(Hasher)
            .inSingletonScope();
        container.bind<TreeTraverser<TreeNode>>(BindingIdentifier.TREE_NODE_TRAVERSER)
            .to(RecursiveTraverser)
            .inSingletonScope();
        container.bind<TextDocumentStore>(BindingIdentifier.TEXT_DOCUMENT_STORE)
            .to(TextDocumentStore)
            .inSingletonScope();
        container.bind<Indexer>(BindingIdentifier.INDEXER)
            .to(Indexer)
            .inSingletonScope();
    }
}

Application.run();