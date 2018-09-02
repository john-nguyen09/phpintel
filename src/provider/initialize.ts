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
import { Logger } from "../service/logger";
import { Hasher } from "../service/hasher";
import { inject } from "inversify";
const pjson = require("../../package.json");
const homedir = require('os').homedir();

export class InitializeProvider {
    constructor(
        @inject(BindingIdentifier.LOGGER) private logger: Logger,
        @inject(BindingIdentifier.HASHER) private hasher: Hasher,
        @inject(BindingIdentifier.INDEXER) private indexer: Indexer
    ) { }

    provide(params: InitializeParams): InitializeResult {
        let rootPath: string = '';

        this.logger.info(`node ${process.version}`);
        this.logger.info(`phpintel ${pjson.version} server started`);

        if (params.rootUri != null) {
            rootPath = uriToPath(params.rootUri);
        } else if (params.rootPath != null || params.rootPath != undefined) {
            rootPath = params.rootPath;
        }

        let storagePath = path.join(homedir, '.phpintel', this.hasher.getHash(rootPath));
        Application.initStorage(storagePath);

        this.logger.info(`storagePath: ${storagePath}`);
        let start = process.hrtime();

        this.indexer.indexDir(rootPath)
            .catch((err) => {
                this.logger.error(err);
            }).then(() => {
                this.logger.info(`Finish indexing in ${elapsed(start).toFixed()} ms`);
            });

        return <InitializeResult>{
            capabilities: {
                textDocumentSync: TextDocumentSyncKind.Full,
                hoverProvider: true
            }
        };
    }
}