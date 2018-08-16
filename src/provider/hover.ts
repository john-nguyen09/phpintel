import { TextDocumentPositionParams, Hover } from "vscode-languageserver";

export namespace HoverProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<Hover> {
        

        return {
            contents: ''
        }
    }
}