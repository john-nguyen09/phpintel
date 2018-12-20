import "reflect-metadata";
import * as path from "path";
import * as fs from "fs";
import { Traverser } from "../src/traverser";
import { SymbolParser } from "../src/symbol/symbolParser";
import { pathToUri } from "../src/util/uri";
import { PhpDocument } from "../src/symbol/phpDocument";
import { Parser } from "php7parser";
import { dumpAstToDebug } from "../src/testHelper";

describe('variable', () => {
    it('simple variable', () => {
        let filePath = path.resolve(__dirname, '..', 'case', 'variable', 'simpleVariable.php');
        let treeTraverser = new Traverser();

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