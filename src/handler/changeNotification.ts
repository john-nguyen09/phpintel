import { DidChangeTextDocumentParams } from "vscode-languageserver";
import { Indexer } from "../index/indexer";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";

export namespace ChangeNotification {
    export async function provide(params: DidChangeTextDocumentParams) {
        const indexer = App.get<Indexer>(Indexer);

        PhpDocumentTable.acquireLock(params.textDocument.uri, async () => {
            const phpDoc = await indexer.getOrCreatePhpDoc(params.textDocument.uri);
            phpDoc.text = params.contentChanges[0].text;
            await indexer.indexFile(phpDoc);
        });
    }
}