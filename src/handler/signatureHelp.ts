import { TextDocumentPositionParams, SignatureHelp } from "vscode-languageserver";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { ReferenceTable } from "../storage/table/reference";
import { RefResolver } from "./refResolver";
import { Reference } from "../symbol/reference";

export namespace SignatureHelpProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<SignatureHelp | null> {
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        let signatureHelp: SignatureHelp | null = null;


        await PhpDocumentTable.acquireLock(params.textDocument.uri, async () => {
            const phpDoc = await phpDocTable.get(params.textDocument.uri);

            if (phpDoc === null) {
                return;
            }

            const offset = phpDoc.getOffset(params.position.line, params.position.character);
            const ref = await refTable.findAt(phpDoc.uri, offset);

            if (ref === null) {
                return;
            }

            signatureHelp = await RefResolver.getSignatureHelp(phpDoc, ref, offset);
        });

        return signatureHelp;
    }
}