import {
    InitializeResult,
    InitializeParams,
    TextDocumentSyncKind
} from "vscode-languageserver";
import { Indexer } from "../index/indexer";
import { uriToPath } from "../util/uri";
import * as path from "path";
import { elapsed } from "../util/hrtime";
import { LogWriter } from "../service/logWriter";
import { Hasher } from "../service/hasher";
import { App } from "../app";
const pjson = require("../../package.json");
const homedir = require('os').homedir();

export namespace InitializeProvider {
    export function provide(params: InitializeParams): InitializeResult {
        const logger = App.get<LogWriter>(LogWriter);
        const hasher = App.get<Hasher>(Hasher);
        let rootPath: string = '';

        logger.info(`node ${process.version}`);
        logger.info(`phpintel ${pjson.version} server started`);

        if (params.rootUri != null) {
            rootPath = uriToPath(params.rootUri);
        } else if (params.rootPath != null || params.rootPath != undefined) {
            rootPath = params.rootPath;
        }

        let storagePath = path.join(homedir, '.phpintel', hasher.getHash(rootPath));
        App.getApp().initStorage(storagePath);

        logger.info(`storagePath: ${storagePath}`);
        let start = process.hrtime();

        let indexer = App.get<Indexer>(Indexer);
        indexer.indexDir(rootPath)
            .catch((err: Error) => {
                logger.error(err);
            }).then(() => {
                logger.info(`Finish indexing in ${elapsed(start).toFixed()} ms`);
            });

        return <InitializeResult>{
            capabilities: {
                textDocumentSync: TextDocumentSyncKind.Full,
                hoverProvider: true
            }
        };
    }
}