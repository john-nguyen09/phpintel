import { SymbolParser } from "../src/symbolParser";
import { PhpDocument } from "../src/symbol/phpDocument";
import { pathToUri } from "../src/util/uri";
import * as path from 'path';
import * as fs from 'fs';
import { Parser } from "php7parser";
import { inspect } from "util";

describe('namespaceDef', () => {
    it('should assign namespace to phpDocument', () => {
        let workspaceDir = path.resolve(__dirname, '..', 'case', 'namespaceDef');
        let files = fs.readdirSync(workspaceDir);
        
        for (let file of files) {
            let filePath = path.join(workspaceDir, file);
            let fileUri = pathToUri(filePath);

            if (file.endsWith('.php')) {
                const fileContent = fs.readFileSync(filePath).toString();
                let symbolParser = new SymbolParser(new PhpDocument(
                    fileUri,
                    fileContent
                ));

                symbolParser.traverse(Parser.parse(fileContent));

                expect(symbolParser.getTree().toObject()).toMatchSnapshot();
            }
        }
    });
});