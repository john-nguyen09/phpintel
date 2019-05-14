import { App } from "./app";
import { createConnection, InitializeParams, InitializeResult, TextDocumentSyncKind, DocumentSymbol } from "vscode-languageserver";
import { LogWriter } from "./service/logWriter";
import { Hasher } from "./service/hasher";
import { uriToPath } from "./util/uri";
import * as path from "path";
import { Indexer } from "./index/indexer";
import { elapsed } from "./util/hrtime";
import { HoverProvider } from "./handler/hover";
import { DefinitionProvider } from "./handler/definition";
import { CompletionProvider } from "./handler/completion";
import { SignatureHelpProvider } from "./handler/signatureHelp";
import { NotificationHandler } from "./handler/notificationHandler";
import { DocumentSymbolProvider } from "./handler/documentSymbol";
const pjson = require("../package.json");
const homedir = require('os').homedir();

const connection = createConnection();
const hasher = new Hasher();

connection.onHover(HoverProvider.provide);
connection.onDefinition(DefinitionProvider.provide);
connection.onCompletion(CompletionProvider.provide);
connection.onSignatureHelp(SignatureHelpProvider.provide);
connection.onDocumentSymbol(DocumentSymbolProvider.provide);

connection.onDidChangeTextDocument(NotificationHandler.change);
connection.onDidOpenTextDocument(NotificationHandler.open);
connection.onDidCloseTextDocument(NotificationHandler.close);

connection.onInitialize((params: InitializeParams): InitializeResult => {
    let rootPath: string = '';

    if (params.rootUri != null) {
        rootPath = uriToPath(params.rootUri);
    } else if (params.rootPath != null || params.rootPath != undefined) {
        rootPath = params.rootPath;
    }

    const storagePath = path.join(homedir, '.phpintel', hasher.getHash(rootPath));
    App.init(storagePath, connection);

    const logger = App.get<LogWriter>(LogWriter);

    logger.info(`node ${process.version}`);
    logger.info(`phpintel ${pjson.version} server started`);
    logger.info(`storagePath: ${storagePath}`);

    let start = process.hrtime();

    let indexer = App.get<Indexer>(Indexer);
    indexer.indexWorkspace(rootPath)
        .catch((err: Error) => {
            logger.error(err);
        }).then(() => {
            logger.info(`Finish indexing in ${elapsed(start).toFixed()} ms`);
        });

    return <InitializeResult>{
        capabilities: {
            textDocumentSync: TextDocumentSyncKind.Full,
            hoverProvider: true,
            definitionProvider: true,
            completionProvider: {
                triggerCharacters: [
                    '$', '>', ':', //php
                    '.', '<', '/' //html/js
                ]
            },
            signatureHelpProvider: {
                triggerCharacters: [
                    '(', ')', ','
                ],
            },
            documentSymbolProvider: true,
        }
    };
});

connection.listen();
