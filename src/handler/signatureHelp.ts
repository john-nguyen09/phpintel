import { TextDocumentPositionParams, SignatureHelp } from "vscode-languageserver";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { RefResolver } from "./refResolver";

export namespace SignatureHelpProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<SignatureHelp | null> {
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        let signatureHelp: SignatureHelp | null = null;


        await PhpDocumentTable.acquireLock(params.textDocument.uri, async () => {
            const phpDoc = await phpDocTable.get(params.textDocument.uri);

            if (phpDoc === null) {
                return;
            }

            const offset = phpDoc.getOffset(params.position.line, params.position.character);
            const argumentList = await phpDoc.findArgumentListAt(offset);

            if (argumentList === null) {
                return;
            }

            signatureHelp = await RefResolver.getSignatureHelp(phpDoc, argumentList, offset);
        });

        return signatureHelp;
    }
}