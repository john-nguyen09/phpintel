import { DocumentSymbolParams, SymbolInformation } from "vscode-languageserver";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { App } from "../app";
import { ClassTable } from "../storage/table/class";
import { ClassConstantTable } from "../storage/table/classConstant";
import { ConstantTable } from "../storage/table/constant";
import { FunctionTable } from "../storage/table/function";
import { MethodTable } from "../storage/table/method";
import { PropertyTable } from "../storage/table/property";
import { Symbol } from "../symbol/symbol";
import { Formatter } from "./formatter";

export namespace DocumentSymbolProvider {
    export async function provide(params: DocumentSymbolParams) {
        const phpDocTable: PhpDocumentTable = App.get<PhpDocumentTable>(PhpDocumentTable);
        const classTable: ClassTable = App.get<ClassTable>(ClassTable);
        const classConstantTable: ClassConstantTable = App.get<ClassConstantTable>(ClassConstantTable);
        const constantTable: ConstantTable = App.get<ConstantTable>(ConstantTable);
        const functionTable: FunctionTable = App.get<FunctionTable>(FunctionTable);
        const methodTable: MethodTable = App.get<MethodTable>(MethodTable);
        const propertyTable: PropertyTable = App.get<PropertyTable>(PropertyTable);
        const symbolInfos: SymbolInformation[] = [];

        await PhpDocumentTable.acquireLock(params.textDocument.uri, async () => {
            const phpDoc = await phpDocTable.get(params.textDocument.uri);
            const promises: Promise<Symbol[]>[] = [];

            if (phpDoc === null) {
                return;
            }

            promises.push(classTable.getByDoc(phpDoc));
            promises.push(classConstantTable.getByDoc(phpDoc));
            promises.push(constantTable.getByDoc(phpDoc));
            promises.push(functionTable.getByDoc(phpDoc));
            promises.push(methodTable.getByDoc(phpDoc));
            promises.push(propertyTable.getByDoc(phpDoc));

            const jobs = await Promise.all(promises);

            for (const symbols of jobs) {
                for (const symbol of symbols) {
                    const symbolInfo = Formatter.getSymbolInfo(phpDoc, symbol);

                    if (symbolInfo === null) {
                        continue;
                    }

                    symbolInfos.push(symbolInfo);
                }
            }
        });

        symbolInfos.sort((first, second): number => {
            if (first.location.range.start.line == second.location.range.start.line) {
                return first.location.range.start.character - second.location.range.start.character;
            }

            return first.location.range.start.line - second.location.range.start.line;
        })

        return symbolInfos;
    }
}