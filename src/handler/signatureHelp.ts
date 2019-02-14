import { TextDocumentPositionParams, SignatureHelp } from "vscode-languageserver";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { ReferenceTable } from "../storage/table/reference";
import { RefResolver } from "./refResolver";

export namespace SignatureHelpProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<SignatureHelp | null> {
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const refTable = App.get<ReferenceTable>(ReferenceTable);

        const phpDoc = await phpDocTable.get(params.textDocument.uri);

        if (phpDoc === null) {
            return null;
        }

        const offset = phpDoc.getOffset(params.position.line, params.position.character);
        const ref = await refTable.findAt(phpDoc.uri, offset);

        if (ref === null) {
            return null;
        }

        return RefResolver.getSignatureHelp(phpDoc, ref, offset);
    }
}