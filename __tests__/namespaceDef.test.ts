import "reflect-metadata";
import { SymbolParser } from "../src/symbol/symbolParser";
import { PhpDocument } from "../src/symbol/phpDocument";
import { pathToUri } from "../src/util/uri";
import * as path from 'path';
import * as fs from 'fs';
import { Parser, phraseTypeToString, tokenTypeToString } from "php7parser";
import { RecursiveTraverser } from "../src/treeTraverser/recursive";

describe('namespaceDef', () => {
    it('should assign namespace to phpDocument', () => {
        let workspaceDir = path.resolve(__dirname, '..', 'case', 'namespaceDef');
        let files = fs.readdirSync(workspaceDir);
        let treeTraverser = new RecursiveTraverser();
        
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
                
                fs.writeFile(
                    path.resolve(__dirname, '..', 'debug', file + '.ast.json'),
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

                treeTraverser.traverse(parseTree, [
                    symbolParser
                ]);

                expect(symbolParser.getTree().toObject()).toMatchSnapshot();
            }
        }
    });
});