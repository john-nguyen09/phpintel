import { PhpFile } from "./phpFile";
import * as path from "path";
import * as fs from "fs";
import { TreeAnalyser } from "./treeAnalyser";
import { FileInfo } from "./indexer";
import { Formatter } from "./util/formatter";
import { inspect } from "util";

describe('test parsing php file', () => {
    it('should return a reflection of php file', async () => {
        const phpFile = await PhpFile.create(
            await FileInfo.create(path.join(__dirname, '..', 'case', 'global_symbols.php'))
        );
        const funcDeclareFile = await PhpFile.create(
            await FileInfo.create(path.join(__dirname, '..', 'case', 'function_declare.php'))
        );

        TreeAnalyser.analyse(phpFile);
        TreeAnalyser.analyse(funcDeclareFile);

        const astString = Formatter.treeSitterOutput(funcDeclareFile.getTree().rootNode.toString());
        const debugDir = path.join(__dirname, '..', 'debug');
        fs.writeFile(path.join(debugDir, path.basename(funcDeclareFile.path) + '.ast'), astString, (err) => {
            if (err) {
                console.log(err);
            }
        });

        console.log(inspect(funcDeclareFile, { depth: 7 }));
    });
});