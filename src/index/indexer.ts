import * as fs from "fs";
import * as path from "path";
import { promisify } from "util";
import { pathToUri } from "../util/uri";
import { SymbolParser } from "../symbol/symbolParser";
import { PhpDocument } from "../symbol/phpDocument";
import { Parser } from "php7parser";
import { TreeTraverser } from "../treeTraverser/structures";
import { TreeNode } from "../util/parseTree";
import { isIdentifiable, Symbol, isLocatable } from "../symbol/symbol";
import { IdentifierIndex } from "./identifierIndex";
import { UriIndex } from "./uriIndex";
import { PositionIndex } from "./positionIndex";
import { TimestampIndex } from "./timestampIndex";
import { inject, injectable } from "inversify";
import { TextDocumentStore } from "../textDocumentStore";
import { BindingIdentifier } from "../constant/bindingIdentifier";
import { Messenger } from "../service/messenger";

const readdir = promisify(fs.readdir);
const readFile = promisify(fs.readFile);
const stat = promisify(fs.stat);

@injectable()
export class Indexer {
    static readonly separator = '#';

    constructor(
        @inject(BindingIdentifier.TREE_NODE_TRAVERSER) private treeTraverser: TreeTraverser<TreeNode>,
        @inject(BindingIdentifier.IDENTIFIER_INDEX) private identifierIndex: IdentifierIndex,
        @inject(BindingIdentifier.URI_INDEX) private uriIndex: UriIndex,
        @inject(BindingIdentifier.POSITION_INDEX) private positionIndex: PositionIndex,
        @inject(BindingIdentifier.TIMESTAMP_INDEX) private timestampIndex: TimestampIndex,
        @inject(BindingIdentifier.TEXT_DOCUMENT_STORE) private textDocumentStore: TextDocumentStore
    ) { }

    async indexDir(directory: string): Promise<void> {
        let files = await readdir(directory);

        for (let file of files) {
            let filePath = path.join(directory, file);
            let fileUri = pathToUri(filePath);
            let fstat = await stat(filePath);
            let lastIndexTime = await this.timestampIndex.get(fileUri);

            if (file.endsWith('.php')) {
                let fileContent = (await readFile(filePath)).toString();
                let phpDoc = new PhpDocument(fileUri, fileContent);

                this.textDocumentStore.add(fileUri, phpDoc.textDocument);

                if (fstat.mtimeMs != lastIndexTime) {
                    let symbolParser = new SymbolParser(new PhpDocument(fileUri, fileContent));
                    let parseTree = Parser.parse(fileContent);

                    this.treeTraverser.traverse(parseTree, [symbolParser]);
                    await this.indexPhpDocument(symbolParser.getTree());
                    await this.timestampIndex.put(fileUri, fstat.mtimeMs);
                }
            } else if (fstat.isDirectory()) {
                await this.indexDir(filePath);
            }
        }
    }

    private async indexBranchSymbol(symbol: Symbol, uri: string): Promise<void> {
        if (!isIdentifiable(symbol)) {
            return;
        }

        await [
            this.identifierIndex.put(symbol, uri),
            this.uriIndex.put(uri, symbol.getIdentifier()),
        ];
    }

    private async indexSymbol(symbol: Symbol): Promise<void> {
        if (!isLocatable(symbol)) {
            return;
        }

        await this.positionIndex.put(symbol);
    }

    private async removeIndexes(uri: string): Promise<void> {
        await [
            this.uriIndex.delete(uri),
            this.positionIndex.delete(uri)
        ];
    }

    private async indexPhpDocument(doc: PhpDocument): Promise<void> {
        // Symbol name index
        this.removeIndexes(doc.uri);
        for (let branchSymbol of doc.branchSymbols) {
            await this.indexBranchSymbol(branchSymbol, doc.uri);
        }
        for (let symbol of doc.symbols) {
            await this.indexSymbol(symbol);
        }
    }
}