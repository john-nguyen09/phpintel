import { TextDocumentPositionParams, Location as LspLocation } from "vscode-languageserver";
import { ReferenceTable } from "../storage/table/referenceTable";
import { App } from "../app";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { RefKind } from "../symbol/reference";
import { RefResolver } from "./refResolver";
import { inspect } from "util";
import { Formatter } from "./formatter";
import { Location } from "../symbol/meta/location";

export namespace DefinitionProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<LspLocation | LspLocation[] | null> {
        const refTable: ReferenceTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable: PhpDocumentTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        let phpDoc = await phpDocTable.get(params.textDocument.uri);
        let result: Location[] = [];

        if (phpDoc === null) {
            return null;
        }

        let ref = await refTable.findAt(
            phpDoc.uri,
            phpDoc.getOffset(params.position.line, params.position.character)
        );

        if (ref === null) {
            return null;
        }

        switch (ref.refKind) {
            case RefKind.Function:
                let funcs = await RefResolver.getFuncSymbols(phpDoc, ref);

                for (let func of funcs) {
                    result.push(func.location);
                }
                break;
            case RefKind.ClassTypeDesignator:
                let constructors = await RefResolver.getMethodSymbols(phpDoc, ref);

                if (constructors.length === 0) {
                    let classes = await RefResolver.getClassSymbols(phpDoc, ref);

                    for (let theClass of classes) {
                        result.push(theClass.location);
                    }
                } else {
                    for (let constructor of constructors) {
                        result.push(constructor.location);
                    }
                }
                break;
            case RefKind.Class:
                let classes = await RefResolver.getClassSymbols(phpDoc, ref);

                for (let theClass of classes) {
                    result.push(theClass.location);
                }
                break;
            case RefKind.Method:
                let methods = await RefResolver.getMethodSymbols(phpDoc, ref);

                for (let method of methods) {
                    result.push(method.location);
                }
                break;
            case RefKind.Property:
                let props = await RefResolver.getPropSymbols(phpDoc, ref);

                for (let prop of props) {
                    result.push(prop.location);
                }
                break;
            case RefKind.ClassConst:
                let classConsts = await RefResolver.getClassConstSymbols(phpDoc, ref);

                for (let classConst of classConsts) {
                    result.push(classConst.location);
                }
                break;
            case RefKind.ConstantAccess:
                let consts = await RefResolver.getConstSymbols(phpDoc, ref);

                for (let constant of consts) {
                    result.push(constant.location);
                }
                break;
        }

        if (result.length === 0) {
            return null;
        } else if (result.length === 1) {
            let defDoc = await phpDocTable.get(result[0].uri);

            if (defDoc === null) {
                return null;
            }

            return Formatter.toLspLocation(defDoc, result[0]);
        } else {
            let lspLocs: LspLocation[] = [];

            for (let loc of result) {
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