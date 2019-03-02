import { Indexer } from "./indexer";
import * as path from "path";
import { HrTime } from "./util/hrtime";

describe('Indexer', () => {
    it('should index the whole workspace', () => {
        const start = process.hrtime();
        Indexer
            .indexWorkspace('C:\\Users\\johnn\\Development\\moodle-lite')
            .then(() => {
                const elapsed = HrTime.elapsed(start);

                console.log(`Finished indexing in ${elapsed} ms`);
            });
    });
});