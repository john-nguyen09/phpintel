import {
    InitializeResult,
    InitializeParams,
    TextDocumentSyncKind
} from "vscode-languageserver";
import { Indexer } from "../index/indexer";
import { uriToPath } from "../util/uri";
import { Logger } from "../util/logger";
import * as path from "path";
import { Hasher } from "../util/hash";
import { DB } from "../storage/db";
import { elapsed } from "../util/hrtime";
const pjson = require("../../package.json");
const homedir = require('os').homedir();

export namespace InitializeProvider {
    export function provide(params: InitializeParams): InitializeResult {
        let indexer = new Indexer();
        let rootPath: string = '';

        Logger.info(`node ${process.version}`);
        Logger.info(`phpintel ${pjson.version} server started`);

        if (params.rootUri != null) {
            rootPath = uriToPath(params.rootUri);
        } else if (params.rootPath != null || params.rootPath != undefined) {
            rootPath = params.rootPath;
        }

        let storagePath = path.join(homedir, '.phpintel', Hasher.getHash(rootPath));
        DB.init(storagePath);

        Logger.info(`storagePath: ${storagePath}`);
        let start = process.hrtime();

        indexer.indexDir(rootPath)
            .catch((err) => {
                Logger.error(err);
            }).then(() => {
                Logger.info(`Finish indexing in ${elapsed(start).toFixed()} ms`);
            });

        return <InitializeResult>{
            capabilities: {
                textDocumentSync: TextDocumentSyncKind.Full,
                hoverProvider: true
            }
        };
    }
}