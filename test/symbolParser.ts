import { describe, it } from 'mocha';
import { PhpDocument } from '../src/phpDocument';
import { pathToUri } from '../src/util/uri';
import * as path from 'path';
import * as fs from 'fs';
import { SymbolParser } from '../src/symbolParser';
import { phraseTypeToString, tokenTypeToString } from '../node_modules/php7parser';

describe('symbolParser', () => {
    it('shall return symbol tree', () => {
        let filePath = path.join(__dirname, 'case', 'global_symbols.php');
        let text = fs.readFileSync(filePath).toString();
        let doc = new PhpDocument(pathToUri(filePath), text);
        let tree = doc.getTree();
        let symbolParser = new SymbolParser(doc);

        symbolParser.traverse(tree);

        console.dir(symbolParser.getTree(), {
            depth: 3
        });

        fs.writeFile(
            path.resolve(__dirname, '..', 'ast.json'),
            JSON.stringify(tree, (key, value) => {
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

        // fs.writeFile(
        //     path.resolve(__dirname, '..', 'test.json'),
        //     JSON.stringify(symbolParser.getTree(), (key, value) => {
        //         if (key == 'node') {
        //             return undefined;
        //         }
    
        //         return value;
        //     }),
        //     (err) => {
        //         if (err) {
        //             console.log(err);
        //         }
        //     }
        // );
    });
});