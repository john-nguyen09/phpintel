import { DidChangeTextDocumentParams } from "vscode-languageserver";
import { Indexer } from "../index/indexer";
import { App } from "../app";

export namespace ChangeNotification {
    export async function provide(params: DidChangeTextDocumentParams) {
        const indexer = App.get<Indexer>(Indexer);

        let phpDoc = await indexer.getOrCreatePhpDoc(params.textDocument.uri);
        
        phpDoc.text = params.contentChanges[0].text;

        await indexer.indexFile(phpDoc);
    }
}