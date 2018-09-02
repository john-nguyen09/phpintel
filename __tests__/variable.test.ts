import "reflect-metadata";
import * as path from "path";
import * as fs from "fs";
import { RecursiveTraverser } from "../src/treeTraverser/recursive";
import { SymbolParser } from "../src/symbol/symbolParser";
import { pathToUri } from "../src/util/uri";
import { PhpDocument } from "../src/symbol/phpDocument";
import { Parser, phraseTypeToString, tokenTypeToString } from "php7parser";
import { TreeNode } from "../src/util/parseTree";

describe('variable', () => {
    it('simple variable', () => {
        let filePath = path.resolve(__dirname, '..', 'case', 'variable', 'simpleVariable.php');
        let treeTraverser = new RecursiveTraverser<TreeNode>();

        const fileUri = pathToUri(filePath);
        const fileContent = fs.readFileSync(filePath).toString();
        let symbolParser = new SymbolParser(new PhpDocument(
            fileUri,
            fileContent
        ));
        let parseTree = Parser.parse(fileContent);

        treeTraverser.traverse(parseTree, [
            symbolParser
        ]);

        fs.writeFile(
            path.resolve(__dirname, '..', 'debug', 'variable', 'simpleVariable.ast.json'),
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
        
        expect(symbolParser.getTree().toObject()).toMatchSnapshot();
    });
});