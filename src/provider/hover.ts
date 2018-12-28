import { TextDocumentPositionParams, Hover, MarkedString, LogMessageNotification } from "vscode-languageserver";
import { App } from "../app";
import { ReferenceTable } from "../storage/table/referenceTable";
import { RefKind } from "../symbol/reference";
import { Range } from "../symbol/meta/range";
import { Formatter } from "./formatter";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { RefResolver } from "./refResolver";

export namespace HoverProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<Hover> {
        const referenceTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        let uri = params.textDocument.uri;
        let phpDoc = await phpDocTable.get(uri);
        let contents: MarkedString[] = [];
        let range: Range | undefined = undefined;

        if (phpDoc !== null) {
            let ref = await referenceTable.findAt(
                uri,
                phpDoc.getOffset(params.position.line, params.position.character)
            );

            if (ref !== null) {
                if (ref.refKind === RefKind.FunctionCall) {
                    range = ref.location.range;
                    let funcs = await RefResolver.getFuncSymbols(phpDoc, ref);

                    for (let func of funcs) {
                        contents.push(Formatter.funcDef(phpDoc, func));
                    }
                }
            }
        }

        return {
            contents,
            range
        };
    }
}