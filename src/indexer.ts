import * as fs from "fs";
import * as path from "path";
import { promisify } from "util";
import { PhpFile } from "./phpFile";
import { TreeAnalyser } from "./treeAnalyser";

const readdirAsync = promisify(fs.readdir);
const statAsync = promisify(fs.stat);

export interface FileInfo {
    filePath: fs.PathLike;
    stats: fs.Stats;
}

export namespace FileInfo {
    export async function create(path: fs.PathLike): Promise<FileInfo> {
        return {
            filePath: path,
            stats: await statAsync(path)
        }
    }
}

export namespace Indexer {
    export async function indexWorkspace(directory: fs.PathLike): Promise<void> {
        let directories: fs.PathLike[] = [
            directory
        ];
        const promises: Promise<void>[] = [];

        while (directories.length > 0) {
            let dir = directories.shift();
            if (dir === undefined) {
                continue;
            }
            let files = await readdirAsync(dir);

            for (let file of files) {
                const fileInfo = await FileInfo.create(path.join(dir.toString(), file));

                if (fileInfo.stats.isDirectory()) {
                    directories.push(fileInfo.filePath);

                    continue;
                }

                if (file.endsWith('.php')) {
                    promises.push(PhpFile.create(fileInfo).then((phpFile) => {
                        TreeAnalyser.analyse(phpFile);
                    }));
                }
            }
        }

        await Promise.all(promises);
    }
}