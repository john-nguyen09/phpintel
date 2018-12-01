import {
    InitializeResult,
    InitializeParams,
    TextDocumentSyncKind
} from "vscode-languageserver";
import { Indexer } from "../index/indexer";
import { uriToPath } from "../util/uri";
import * as path from "path";
import { elapsed } from "../util/hrtime";
import { Application } from "../app";
import { LogWriter } from "../service/logWriter";
import { Hasher } from "../service/hasher";
import { BindingIdentifier } from "../constant/bindingIdentifier";
const pjson = require("../../package.json");
const homedir = require('os').homedir();

export namespace InitializeProvider {
    export function provide(params: InitializeParams): InitializeResult {
        const logger = Application.get<LogWriter>(BindingIdentifier.MESSENGER);
        const hasher = Application.get<Hasher>(BindingIdentifier.HASHER);
        let rootPath: string = '';

        logger.info(`node ${process.version}`);
        logger.info(`phpintel ${pjson.version} server started`);

        if (params.rootUri != null) {
            rootPath = uriToPath(params.rootUri);
        } else if (params.rootPath != null || params.rootPath != undefined) {
            rootPath = params.rootPath;
        }

        let storagePath = path.join(homedir, '.phpintel', hasher.getHash(rootPath));
        Application.initStorage(storagePath);

        logger.info(`storagePath: ${storagePath}`);
        let start = process.hrtime();

        let indexer = Application.get<Indexer>(BindingIdentifier.INDEXER);
        indexer.indexDir(rootPath)
            .catch((err) => {
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