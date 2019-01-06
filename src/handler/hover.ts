import { TextDocumentPositionParams, Hover, MarkedString } from "vscode-languageserver";
import { App } from "../app";
import { ReferenceTable } from "../storage/table/referenceTable";
import { RefKind } from "../symbol/reference";
import { Range as LspRange } from "vscode-languageserver";
import { Formatter } from "./formatter";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { RefResolver } from "./refResolver";
import { Range } from "../symbol/meta/range";

export namespace HoverProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<Hover> {
        const refTable = App.get<ReferenceTable>(ReferenceTable);
        const phpDocTable = App.get<PhpDocumentTable>(PhpDocumentTable);

        let uri = params.textDocument.uri;
        let phpDoc = await phpDocTable.get(uri);
        let contents: MarkedString[] = [];
        let range: Range | undefined = undefined;

        if (phpDoc !== null) {
            let ref = await refTable.findAt(
                uri,
                phpDoc.getOffset(params.position.line, params.position.character)
            );

            if (ref !== null) {
                range = ref.location.range;

                if (ref.refKind === RefKind.Function) {
                    let funcs = await RefResolver.getFuncSymbols(phpDoc, ref);

                    for (let func of funcs) {
                        contents.push(Formatter.funcDef(phpDoc, func));
                    }
                } else if (ref.refKind === RefKind.ClassTypeDesignator) {
                    let constructors = await RefResolver.getMethodSymbols(phpDoc, ref);

                    if (constructors.length === 0) {
                        let classes = await RefResolver.getClassSymbols(phpDoc, ref);
    
                        for (let theClass of classes) {
                            contents.push(Formatter.classDef(phpDoc, theClass));
                        }
                    } else {
                        for (let constructor of constructors) {
                            contents.push(Formatter.methodDef(phpDoc, constructor));
                        }
                    }
                } else if (ref.refKind === RefKind.Class) {
                    let classes = await RefResolver.getClassSymbols(phpDoc, ref);

                    for (let theClass of classes) {
                        contents.push(Formatter.classDef(phpDoc, theClass));
                    }
                } else if (ref.refKind === RefKind.Method) {
                    let methods = await RefResolver.getMethodSymbols(phpDoc, ref);

                    for (let method of methods) {
                        contents.push(Formatter.methodDef(phpDoc, method));
                    }
                } else if (ref.refKind === RefKind.Property) {
                    let props = await RefResolver.getPropSymbols(phpDoc, ref);

                    for (let prop of props) {
                        contents.push(Formatter.propDef(phpDoc, prop));
                    }
                } else if (ref.refKind === RefKind.ClassConst) {
                    let classConsts = await RefResolver.getClassConstSymbols(phpDoc, ref);

                    for (let classConst of classConsts) {
                        contents.push(Formatter.classConstDef(phpDoc, classConst));
                    }
                } else if (ref.refKind === RefKind.Variable) {
                    contents.push(Formatter.varRef(phpDoc, ref));
                } else if (ref.refKind === RefKind.ConstantAccess) {
                    let consts = await RefResolver.getConstSymbols(phpDoc, ref);

                    for (let constant of consts) {
                        contents.push(Formatter.constDef(phpDoc, constant));
                    }
                }
            }
        }

        let lspRange: LspRange | undefined = undefined;

        if (phpDoc !== null && range !== undefined) {
            lspRange = Formatter.toLspRange(phpDoc, range);
        }

        return {
            contents,
            range: lspRange
        };
    }
}