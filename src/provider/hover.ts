import { TextDocumentPositionParams, Hover } from "vscode-languageserver";
import { TextDocumentStore } from "../textDocumentStore";
import { LogWriter } from "../service/logWriter";
import { App } from "../app";
import { ReferenceTable } from "../storage/table/referenceTable";

export namespace HoverProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<Hover> {
        const textDocumentStore = App.get<TextDocumentStore>(TextDocumentStore);
        const logger = App.get<LogWriter>(LogWriter);
        const referenceTable = App.get<ReferenceTable>(ReferenceTable);

        let uri = params.textDocument.uri;
        let textDocument = textDocumentStore.get(uri);

        if (typeof textDocument !== 'undefined') {
            let ref = referenceTable.findAt(
                uri,
                textDocument.getOffset(params.position.line, params.position.character)
            );

            logger.info(JSON.stringify(ref));
        }

        return {
            contents: ''
        }
    }
}