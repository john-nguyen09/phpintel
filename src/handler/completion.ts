import { CompletionParams, CompletionItem, CompletionList } from "vscode-languageserver";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { RefResolver } from "./refResolver";
import { ReferenceTable } from "../storage/table/reference";
import { Function } from "../symbol/function/function";
import { Formatter } from "./formatter";
import { Class } from "../symbol/class/class";
import { Constant } from "../symbol/constant/constant";
import { Variable } from "../symbol/variable/variable";
import { Property } from "../symbol/variable/property";
import { Method } from "../symbol/function/method";
import { ClassConstant } from "../symbol/constant/classConstant";

export namespace CompletionProvider {
    export async function provide(params: CompletionParams):
        Promise<CompletionItem[] | CompletionList | null | undefined> {
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        let items: CompletionItem[] = [];

        await PhpDocumentTable.acquireLock(params.textDocument.uri, async () => {
            const phpDoc = await phpDocTable.get(params.textDocument.uri);
            if (phpDoc === null) {
                return;
            }

            const offset = phpDoc.getOffset(params.position.line, params.position.character);
            const ref = await refTable.findAt(params.textDocument.uri, offset);
            if (ref === null) {
                return;
            }

            const symbols = await RefResolver.searchSymbolsForReference(phpDoc, ref, offset);

            for (let symbol of symbols) {
                if (symbol instanceof Function) {
                    items.push(Formatter.getFunctionCompletion(phpDoc, symbol));
                } else if (symbol instanceof Class) {
                    items.push(Formatter.getClassCompletion(phpDoc, symbol));
                } else if (symbol instanceof Constant) {
                    items.push(Formatter.getConstantCompletion(phpDoc, symbol));
                } else if (symbol instanceof Variable) {
                    items.push(Formatter.getVariableCompletion(symbol));
                } else if (symbol instanceof Property) {
                    items.push(Formatter.getPropertyCompletion(symbol));
                } else if (symbol instanceof Method) {
                    items.push(Formatter.getMethodCompletion(symbol));
                } else if (symbol instanceof ClassConstant) {
                    items.push(Formatter.getClassConstantCompletion(symbol));
                }
            }
        });

        if (items.length === 0) {
            return null;
        }

        return {
            isIncomplete: true,
            items
        };
    }
}