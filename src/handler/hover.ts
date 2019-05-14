import { TextDocumentPositionParams, Hover, MarkedString } from "vscode-languageserver";
import { App } from "../app";
import { Range as LspRange } from "vscode-languageserver";
import { Formatter } from "./formatter";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { RefResolver } from "./refResolver";
import { Range } from "../symbol/meta/range";
import { Class } from "../symbol/class/class";
import { Function } from "../symbol/function/function";
import { Constant } from "../symbol/constant/constant";
import { Method } from "../symbol/function/method";
import { Property } from "../symbol/variable/property";
import { ClassConstant } from "../symbol/constant/classConstant";
import { DefineConstant } from "../symbol/constant/defineConstant";

export namespace HoverProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<Hover> {
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        let uri = params.textDocument.uri;
        let contents: MarkedString[] = [];
        let lspRange: LspRange | undefined = undefined;

        await PhpDocumentTable.acquireLock(params.textDocument.uri, async () => {
            let phpDoc = await phpDocTable.get(uri);
            let range: Range | undefined = undefined;

            if (phpDoc !== null) {
                let ref = await phpDoc.findRefAt(phpDoc.getOffset(params.position.line, params.position.character));

                if (ref !== null) {
                    range = ref.location.range;

                    let symbols = await RefResolver.getSymbolsByReference(phpDoc, ref);

                    for (let symbol of symbols) {
                        if (symbol instanceof Class) {
                            contents.push(Formatter.classDef(phpDoc, symbol));
                        } else if (symbol instanceof Function) {
                            contents.push(Formatter.funcDef(phpDoc, symbol));
                        } else if (symbol instanceof Constant || symbol instanceof DefineConstant) {
                            contents.push(Formatter.constDef(phpDoc, symbol));
                        } else if (symbol instanceof Method) {
                            contents.push(Formatter.methodDef(phpDoc, symbol));
                        } else if (symbol instanceof Property) {
                            contents.push(Formatter.propDef(phpDoc, symbol));
                        } else if (symbol instanceof ClassConstant) {
                            contents.push(Formatter.classConstDef(phpDoc, symbol));
                        }
                    }
                }
                if (range !== undefined) {
                    lspRange = Formatter.toLspRange(phpDoc, range);
                }
            }
        });

        return {
            contents,
            range: lspRange
        };
    }
}