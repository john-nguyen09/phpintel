import { TextDocument } from "./textDocument";

export class TextDocumentStore {
    private textDocuments: {[key: string]: TextDocument} = {};

    add(uri: string, textDoc: TextDocument): void {
        this.textDocuments[uri] = textDoc;
    }

    remove(uri: string): void {
        if (uri in this.textDocuments) {
            delete this.textDocuments[uri];
        }
    }

    get(uri: string): TextDocument | undefined {
        return this.textDocuments[uri];
    }
}