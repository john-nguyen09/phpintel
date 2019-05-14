import { indexFiles, getCaseDir, dumpToDebug, dumpAstToDebug } from "../src/testHelper";
import * as path from "path";

describe('phpDoc', () => {
    it('snapshot of phpDocs', () => {
        let phpDocs = indexFiles([
            path.join(getCaseDir(), 'class_constants.php'),
            path.join(getCaseDir(), 'class_methods.php'),
            path.join(getCaseDir(), 'function_declare.php'),
            path.join(getCaseDir(), 'global_symbols.php'),
        ]);

        for (let phpDoc of phpDocs) {
            // dumpAstToDebug(path.basename(phpDoc.uri) + '.ast.json', phpDoc.getTree());
            expect(phpDoc.toObject()).toMatchSnapshot();
        }
    });
    
    it('snapshot of phpDocs 2', () => {
        let phpDocs = indexFiles([
            path.join(getCaseDir(), 'moodle_database.php'),
        ]);

        for (let phpDoc of phpDocs) {
            // dumpAstToDebug(path.basename(phpDoc.uri) + '.ast.json', phpDoc.getTree());
            console.log(phpDoc.methods.map((method) => {
                return method.types;
            }));
        }
    });
});