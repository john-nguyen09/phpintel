import { IdentifierIndex } from "./index/identifierIndex";
import { PositionIndex } from "./index/positionIndex";
import { TimestampIndex } from "./index/timestampIndex";
import { LevelDatasource, DbStore } from "./storage/db";
import { UriIndex } from "./index/uriIndex";
import { IConnection, createConnection } from "vscode-languageserver";
import { Logger } from "./service/logger";
import { Hasher } from "./service/hasher";
import { InitializeProvider } from "./provider/initialize";
import { HoverProvider } from "./provider/hover";
import { Container } from "inversify";
import { IndexNames, IndexId, IndexVersion } from "./constant/index";
import { TreeTraverser } from "./treeTraverser/structures";
import { TreeNode } from "./util/parseTree";
import { RecursiveTraverser } from "./treeTraverser/recursive";
import { Indexer } from "./index/indexer";

export namespace Application {
    let container: Container;

    export function run() {
        beforeConnection();
        initConnection();
        afterConnection();
    }

    export function initStorage(location: string) {
        let datasource = new LevelDatasource(location, {
            encodingValue: 'json'
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

        container.bind<IdentifierIndex>(BindingIdentifier.IDENTIFIER_INDEX).to(IdentifierIndex);
        container.bind<PositionIndex>(BindingIdentifier.POSITION_INDEX).to(PositionIndex);
        container.bind<TimestampIndex>(BindingIdentifier.TIMESTAMP_INDEX).to(TimestampIndex);
        container.bind<UriIndex>(BindingIdentifier.URI_INDEX).to(UriIndex);
    }

    function beforeConnection() {
        container.bind<InitializeProvider>(BindingIdentifier.INITIALIZE_PROVIDER)
            .to(InitializeProvider);
        container.bind<HoverProvider>(BindingIdentifier.HOVER_PROVIDER)
            .to(HoverProvider);
    }

    function afterConnection() {
        container.bind<Logger>(BindingIdentifier.LOGGER).to(Logger);
        container.bind<Hasher>(BindingIdentifier.HASHER).to(Hasher);
        container.bind<TreeTraverser<TreeNode>>(BindingIdentifier.TREE_NODE_TRAVERSER)
            .to(RecursiveTraverser);
        container.bind<Indexer>(BindingIdentifier.INDEXER).to(Indexer);
    }

    function initConnection() {
        let connection = createConnection();

        connection.onInitialize(
            container.get<InitializeProvider>(BindingIdentifier.INITIALIZE_PROVIDER).provide
        );
        connection.onHover(
            container.get<HoverProvider>(BindingIdentifier.HOVER_PROVIDER).provide
        );
        connection.listen();

        container.bind<IConnection>(BindingIdentifier.CONNECTION).toConstantValue(connection);
    }
}