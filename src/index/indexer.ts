import * as fs from "fs";
import * as path from "path";
import { promisify } from "util";
import { pathToUri } from "../util/uri";
import { SymbolParser } from "../symbol/symbolParser";
import { PhpDocument } from "../symbol/phpDocument";
import { Parser } from "php7parser";
import { TreeTraverser } from "../treeTraverser/structures";
import { TreeNode } from "../util/parseTree";
import { RecursiveTraverser } from "../treeTraverser/recursive";
import { isIdentifiable, Symbol } from "../symbol/symbol";
import { IdentifierMatchIndex } from "./identifierMatch";
import { UriMatchIndex } from "./uriMatch";
import { LocationMatchIndex } from "./locationMatch";

const readdir = promisify(fs.readdir);
const readFile = promisify(fs.readFile);
const stat = promisify(fs.stat);

export class Indexer {
    static readonly separator = '#';

    private treeTraverser: TreeTraverser<TreeNode>;
    private identifierMatch: IdentifierMatchIndex;
    private uriMatch: UriMatchIndex;
    private locationMatch: LocationMatchIndex;

    constructor() {
        this.treeTraverser = new RecursiveTraverser();
        this.identifierMatch = new IdentifierMatchIndex();
        this.uriMatch = new UriMatchIndex(this.identifierMatch);
        this.locationMatch = new LocationMatchIndex(this.identifierMatch);
    }

    async indexDir(directory: string): Promise<void> {
        let files = await readdir(directory);

        for (let file of files) {
            let filePath = path.join(directory, file);
            let fileUri = pathToUri(filePath);

            if (file.endsWith('.php')) {
                let fileContent = (await readFile(filePath)).toString();
                let symbolParser = new SymbolParser(new PhpDocument(fileUri, fileContent));
                let parseTree = Parser.parse(fileContent);

                this.treeTraverser.traverse(parseTree, [symbolParser]);
                await this.indexPhpDocument(symbolParser.getTree());
            } else {
                let fstat = await stat(filePath);

                if (fstat.isDirectory()) {
                    await this.indexDir(filePath);
                }
            }
        }
    }

    private async indexSymbol(symbol: Symbol, uri: string): Promise<void> {
        if (!isIdentifiable(symbol)) {
            return;
        }

        await [
            this.identifierMatch.put(symbol, uri),
            this.uriMatch.put(uri, symbol.getIdentifier()),
            this.locationMatch.put(symbol)
        ];
    }

    private async removeIndexes(uri: string): Promise<void> {
        await [
            this.uriMatch.delete(uri),
            this.locationMatch.delete(uri)
        ];
    }

    private async indexPhpDocument(doc: PhpDocument): Promise<void> {
        // Symbol name index
        this.removeIndexes(doc.uri);
        for (let symbol of doc.symbols) {
            await this.indexSymbol(symbol, doc.uri);
        }
    }
}