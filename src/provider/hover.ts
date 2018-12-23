import { TextDocumentPositionParams, Hover, MarkedString, LogMessageNotification } from "vscode-languageserver";
import { TextDocumentStore } from "../textDocumentStore";
import { App } from "../app";
import { ReferenceTable } from "../storage/table/referenceTable";
import { RefKind, Reference } from "../symbol/reference";
import { Range } from "../symbol/meta/range";
import { FunctionTable } from "../storage/table/function";
import { Formatter } from "./formatter";
import { TypeComposite } from "../type/composite";
import { LogWriter } from "../service/logWriter";
import { inspect } from "util";

export namespace HoverProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<Hover> {
        const textDocumentStore = App.get<TextDocumentStore>(TextDocumentStore);
        const referenceTable = App.get<ReferenceTable>(ReferenceTable);
        const functionTable = App.get<FunctionTable>(FunctionTable);

        let uri = params.textDocument.uri;
        let textDocument = textDocumentStore.get(uri);
        let contents: string[] = [];
        let range: Range | undefined = undefined;

        if (typeof textDocument !== 'undefined') {
            let ref = await referenceTable.findAt(
                uri,
                textDocument.getOffset(params.position.line, params.position.character)
            );

            if (ref !== null) {
                if (ref.refKind === RefKind.FunctionCall) {
                    contents = await getFuncDefConts(ref, functionTable);
                    range = ref.location.range;
                }
            }
        }

        let finalContents: MarkedString[] = [];

        if (contents.length !== 0) {
            for (let content of contents) {
                finalContents.push(Formatter.beautifyPhpContent(content));
            }
        }

        return {
            contents: finalContents,
            range
        };
    }

    async function getFuncDefConts(ref: Reference, functionTable: FunctionTable): Promise<string[]> {
        if (ref.type instanceof TypeComposite) {
            return [];
        }

        let symbols = await functionTable.get(ref.type.getName());
        let contents: string[] = [];

        for (let symbol of symbols) {
            contents.push(Formatter.funcDef(symbol));
        }

        return contents;
    }
}