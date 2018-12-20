import { PhpDocument } from "./symbol/phpDocument";
import { pathToUri } from "./util/uri";
import * as fs from "fs";
import { SymbolParser } from "./symbol/symbolParser";
import { Traverser } from "./traverser";
import { TreeNode } from "./util/parseTree";
import { Parser, Phrase, phraseTypeToString, tokenTypeToString } from "php7parser";
import * as path from "path";
import { inspect } from "util";

export function indexFiles(filePaths: string[]): PhpDocument[] {
    let phpDocs: PhpDocument[] = [];
    let treeTraverser = new Traverser();

    for (let filePath of filePaths) {
        const fileUri = pathToUri(filePath);
        const fileContent = fs.readFileSync(filePath).toString();
        const symbolParser = new SymbolParser(new PhpDocument(fileUri, fileContent));
        const parseTree = Parser.parse(fileContent);
        
        treeTraverser.traverse(parseTree, [symbolParser]);

        phpDocs.push(symbolParser.getTree());
    }

    return phpDocs;
}

export function getCaseDir(): string {
    return path.resolve(__dirname, "..", "case");
}

export function getDebugDir(): string {
    return path.resolve(__dirname, '..', 'debug');
}

export function dumpToDebug(name: string, object: any, depth?: number): void {
    fs.writeFile(
        path.resolve(__dirname, '..', 'debug', name),
        inspect(object, {
            depth: depth
        }),
        (err) => {
            console.error(err);
        }
    );
}

export function dumpAstToDebug(name: string, parseTree: Phrase): void {
    fs.writeFile(
        path.resolve(__dirname, '..', 'debug', name),
        JSON.stringify(parseTree, (key, value) => {
            if (key == 'modeStack') {
                return undefined;
            }

            if (key == 'phraseType') {
                return phraseTypeToString(value);
            }

            if (key == 'tokenType') {
                return tokenTypeToString(value);
            }

            return value;
        }, 2),
        (err) => {
            if (err) {
                console.log(err);
            }
        }
    );
}