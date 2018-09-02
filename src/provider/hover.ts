import { TextDocumentPositionParams, Hover } from "vscode-languageserver";
import { PositionIndex } from "../index/positionIndex";
import { inject, injectable } from "inversify";
import { TextDocumentStore } from "../textDocumentStore";
import { Messenger } from "../service/messenger";
import { BindingIdentifier } from "../constant/bindingIdentifier";
import { Application } from "../app";

export namespace HoverProvider {
    export async function provide(params: TextDocumentPositionParams): Promise<Hover> {
        const textDocumentStore = Application.get<TextDocumentStore>(BindingIdentifier.TEXT_DOCUMENT_STORE);
        const positionIndex = Application.get<PositionIndex>(BindingIdentifier.POSITION_INDEX);
        const logger = Application.get<Messenger>(BindingIdentifier.MESSENGER);

        let uri = params.textDocument.uri;
        let textDocument = textDocumentStore.get(uri);
        
        if (textDocument != undefined) {
            try {
                let offset = textDocument.getOffset(params.position.line, params.position.character);
                logger.info(offset.toString());
                let symbol = await positionIndex.find(uri, offset);
    
                logger.info(JSON.stringify(symbol));
            } catch(err) {
                logger.error(err);
            }
        }

        return {
            contents: ''
        }
    }
}