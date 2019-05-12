import { TextDocumentPositionParams, Location as LspLocation } from "vscode-languageserver";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { RefResolver } from "./refResolver";
import { Formatter } from "./formatter";
import { Location } from "../symbol/meta/location";
import { isLocatable } from "../symbol/symbol";

export namespace DefinitionProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<LspLocation | LspLocation[] | null> {
        const phpDocTable: PhpDocumentTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        let phpDoc = await phpDocTable.get(params.textDocument.uri);
        let result: Location[] = [];

        if (phpDoc === null) {
            return null;
        }

        let ref = await phpDoc.findRefAt(phpDoc.getOffset(params.position.line, params.position.character));

        if (ref === null) {
            return null;
        }

        let symbols = await RefResolver.getSymbolsByReference(phpDoc, ref);
        for (let symbol of symbols) {
            if (isLocatable(symbol)) {
                result.push(symbol.location);
            }
        }

        if (result.length === 0) {
            return null;
        } else if (result.length === 1) {
            if (result[0].uri === undefined) {
                return null;
            }

            let defDoc = await phpDocTable.get(result[0].uri);

            if (defDoc === null) {
                return null;
            }

            return Formatter.toLspLocation(defDoc, result[0]);
        } else {
            let lspLocs: LspLocation[] = [];

            for (let loc of result) {
                if (loc.uri === undefined) {
                    continue;
                }

                let defDoc = await phpDocTable.get(loc.uri);
                if (defDoc === null) {
                    continue;
                }

                lspLocs.push(Formatter.toLspLocation(defDoc, loc));
            }

            return lspLocs;
        }
    }
}