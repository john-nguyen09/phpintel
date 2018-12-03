import "reflect-metadata";
import * as path from "path";
import * as fs from "fs";
import { RecursiveTraverser } from "../src/treeTraverser";
import { SymbolParser } from "../src/symbol/symbolParser";
import { pathToUri } from "../src/util/uri";
import { PhpDocument } from "../src/symbol/phpDocument";
import { Parser, phraseTypeToString, tokenTypeToString } from "php7parser";
import { TreeNode } from "../src/util/parseTree";
import { dumpAstToDebug } from "../src/testHelper";

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

        dumpAstToDebug(path.join('variable', 'simpleVariable.ast.json'), parseTree);
        
        expect(symbolParser.getTree().toObject()).toMatchSnapshot();
    });
});