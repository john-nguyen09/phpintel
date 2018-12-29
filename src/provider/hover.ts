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
                range = ref.location.range;

                if (ref.refKind === RefKind.FunctionCall) {
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
                } else if (ref.refKind === RefKind.MethodCall) {
                    let methods = await RefResolver.getMethodSymbols(phpDoc, ref);

                    for (let method of methods) {
                        contents.push(Formatter.methodDef(phpDoc, method));
                    }
                } else if (ref.refKind === RefKind.Property) {
                    let props = await RefResolver.getPropSymbols(phpDoc, ref);

                    for (let prop of props) {
                        contents.push(Formatter.propDef(phpDoc, prop));
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