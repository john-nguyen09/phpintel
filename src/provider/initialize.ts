import {
    InitializeResult,
    InitializeParams,
    TextDocumentSyncKind
} from "vscode-languageserver";
import { Indexer } from "../index/indexer";
import { pathToUri } from "../util/uri";

export namespace InitializeProvider {
    export function provide(params: InitializeParams): InitializeResult {
        let indexer = new Indexer();
        let rootUri = '';

        if (params.rootPath != null || params.rootPath != undefined) {
            rootUri = pathToUri(params.rootPath);
        }

        if (params.rootUri != null) {
            rootUri = params.rootUri;
        }

        indexer.indexDir(rootUri);

        return <InitializeResult>{
            capabilities: {
                textDocumentSync: TextDocumentSyncKind.Full,
                hoverProvider: true
            }
        };
    }
}