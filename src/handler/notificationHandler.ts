import { DidOpenTextDocumentParams, DidCloseTextDocumentParams, DidChangeTextDocumentParams } from "vscode-languageserver";
import { Indexer } from "../index/indexer";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";

export namespace NotificationHandler {
    export async function change(params: DidChangeTextDocumentParams) {
        const indexer = App.get<Indexer>(Indexer);

        PhpDocumentTable.acquireLock(params.textDocument.uri, async () => {
            const phpDoc = await indexer.getOrCreatePhpDoc(params.textDocument.uri);
            phpDoc.text = params.contentChanges[0].text;
            phpDoc.refresh();
            await indexer.indexFile(phpDoc);
        });
    }

    export async function open(params: DidOpenTextDocumentParams) {
        const indexer = App.get<Indexer>(Indexer);

        await indexer.open(params.textDocument.uri);
    }

    export async function close(params: DidCloseTextDocumentParams) {
        const indexer = App.get<Indexer>(Indexer);

        indexer.close(params.textDocument.uri);        
    }
}