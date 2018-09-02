import { TextDocumentPositionParams, Hover } from "vscode-languageserver";
import { PositionIndex } from "../index/positionIndex";
import { inject } from "inversify";

export class HoverProvider {
    constructor(
        @inject(BindingIdentifier.POSITION_INDEX) private positionIndex: PositionIndex
    ) { }

    async provide(params: TextDocumentPositionParams): Promise<Hover> {

        return {
            contents: ''
        }
    }
}