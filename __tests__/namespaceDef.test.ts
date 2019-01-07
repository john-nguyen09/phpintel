import "reflect-metadata";
import { SymbolParser } from "../src/symbol/symbolParser";
import { PhpDocument } from "../src/symbol/phpDocument";
import { pathToUri } from "../src/util/uri";
import * as path from 'path';
import * as fs from 'fs';
import { Parser, phraseTypeToString, tokenTypeToString } from "php7parser";
import { Traverser } from "../src/traverser";

describe('namespaceDef', () => {
    it('should assign namespace to phpDocument', () => {
        let workspaceDir = path.resolve(__dirname, '..', 'case', 'namespaceDef');
        let files = fs.readdirSync(workspaceDir);
        let treeTraverser = new Traverser();

        for (let file of files) {
            let filePath = path.join(workspaceDir, file);
            let fileUri = pathToUri(filePath);

            if (file.endsWith('.php')) {
                const fileContent = fs.readFileSync(filePath).toString();
                let symbolParser = new SymbolParser(new PhpDocument(
                    fileUri,
                    fileContent
                ));
                let parseTree = Parser.parse(fileContent);

                treeTraverser.traverse(parseTree, [
                    symbolParser
                ]);

                expect(symbolParser.getPhpDoc().toObject()).toMatchSnapshot();
            }
        }
    });
});