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
        const lspLocs: LspLocation[] = [];

        await PhpDocumentTable.acquireLock(params.textDocument.uri, async () => {
            let phpDoc = await phpDocTable.get(params.textDocument.uri);
            let result: Location[] = [];
    
            if (phpDoc === null) {
                return;
            }
    
            let ref = phpDoc.findRefAt(phpDoc.getOffset(params.position.line, params.position.character));
    
            if (ref === null) {
                return;
            }
    
            let symbols = await RefResolver.getSymbolsByReference(phpDoc, ref);
            for (let symbol of symbols) {
                if (isLocatable(symbol)) {
                    result.push(symbol.location);
                }
            }
    
            if (result.length === 0) {
                return;
            }
    
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
        });

        if (lspLocs.length === 0) {
            return null;
        } else if (lspLocs.length === 1) {
            return lspLocs[0];
        } else {
            return lspLocs;
        }
    }
}