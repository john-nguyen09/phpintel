import { DidChangeTextDocumentParams } from "vscode-languageserver";
import { Indexer } from "../index/indexer";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import * as util from "util";

export namespace ChangeNotification {
    export async function provide(params: DidChangeTextDocumentParams) {
        const indexer = App.get<Indexer>(Indexer);

        await PhpDocumentTable.acquireLock(params.textDocument.uri);

        indexer
            .getOrCreatePhpDoc(params.textDocument.uri)
            .then(async (phpDoc) => {
                phpDoc.text = params.contentChanges[0].text;
                await indexer.indexFile(phpDoc);
                PhpDocumentTable.release(params.textDocument.uri);
            });
    }
}