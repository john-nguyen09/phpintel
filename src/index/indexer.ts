import * as fs from "fs";
import * as path from "path";
import { promisify } from "util";
import { pathToUri, uriToPath } from "../util/uri";
import { SymbolParser } from "../symbol/symbolParser";
import { PhpDocument } from "../symbol/phpDocument";
import { injectable } from "inversify";
import { Traverser } from "../traverser";
import { ClassTable } from "../storage/table/class";
import { ClassConstantTable } from "../storage/table/classConstant";
import { ConstantTable } from "../storage/table/constant";
import { FunctionTable } from "../storage/table/function";
import { MethodTable } from "../storage/table/method";
import { PropertyTable } from "../storage/table/property";
import { PhpDocumentTable } from "../storage/table/phpDoc";
import { GlobalVariableTable } from "../storage/table/globalVariable";

const readdirAsync = promisify(fs.readdir);
const readFileAsync = promisify(fs.readFile);
const statAsync = promisify(fs.stat);

export interface PhpFileInfo {
    filePath: string;
    fstats: fs.Stats;
}

export namespace PhpFileInfo {
    export async function createFileInfo(filePath: string): Promise<PhpFileInfo> {
        return {
            filePath: filePath,
            fstats: await statAsync(filePath)
        };
    }
}

@injectable()
export class Indexer {
    private openedUri: Set<string> = new Set<string>();

    constructor(
        private treeTraverser: Traverser,
        private phpDocTable: PhpDocumentTable,
        private classTable: ClassTable,
        private classConstantTable: ClassConstantTable,
        private constantTable: ConstantTable,
        private functionTable: FunctionTable,
        private methodTable: MethodTable,
        private propertyTable: PropertyTable,
        private globalVariableTable: GlobalVariableTable
    ) { }

    async getOrCreatePhpDoc(uri: string): Promise<PhpDocument> {
        let phpDoc = await this.phpDocTable.get(uri);

        if (phpDoc === null) {
            let fileContent = (await readFileAsync(uriToPath(uri))).toString('utf-8');
            phpDoc = new PhpDocument(uri, fileContent);
        }

        return phpDoc;
    }

    async syncFileSystem(fileInfo: PhpFileInfo): Promise<void> {
        let fileUri = pathToUri(fileInfo.filePath);
        const fileModifiedTime = Math.round(fileInfo.fstats.mtime.getTime() / 1000);

        let phpDoc = await this.getOrCreatePhpDoc(fileUri);
        if (phpDoc.modifiedTime !== fileModifiedTime) {
            phpDoc.modifiedTime = fileModifiedTime;
            await this.indexFile(phpDoc);
        }
    }

    async indexFile(phpDoc: PhpDocument): Promise<void> {
        let symbolParser = new SymbolParser(phpDoc);

        this.treeTraverser.traverse(phpDoc.getTree(), [symbolParser]);
        await this.indexPhpDocument(symbolParser.getPhpDoc());
    }

    async indexWorkspace(directory: string): Promise<void> {
        let directories: string[] = [
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
                const fileInfo = await PhpFileInfo.createFileInfo(path.join(dir, file));

                if (file.endsWith('.php')) {
                    promises.push(this.syncFileSystem(fileInfo));
                } else if (fileInfo.fstats.isDirectory()) {
                    directories.push(fileInfo.filePath);
                }
            }
        }

        await Promise.all(promises);
    }

    public async open(uri: string): Promise<void> {
        this.openedUri.add(uri);
        const phpDoc = await this.getOrCreatePhpDoc(uri);
        await this.indexFile(phpDoc);
    }

    public close(uri: string): void {
        this.openedUri.delete(uri);
    }

    private async removeSymbolsByDoc(uri: string) {
        return Promise.all([
            this.classTable.removeByDoc(uri),
            this.classConstantTable.removeByDoc(uri),
            this.constantTable.removeByDoc(uri),
            this.functionTable.removeByDoc(uri),
            this.methodTable.removeByDoc(uri),
            this.propertyTable.removeByDoc(uri),
            this.globalVariableTable.removeByDoc(uri),
        ]);
    }

    private async indexPhpDocument(doc: PhpDocument): Promise<void> {
        await this.removeSymbolsByDoc(doc.uri);

        const promises: Promise<void | void[] | void[][]>[] = [];

        for (const theClass of doc.classes) {
            promises.push(this.classTable.put(doc, theClass));
        }

        for (const classConstant of doc.classConstants) {
            promises.push(this.classConstantTable.put(doc, classConstant));
        }

        for (const constant of doc.constants) {
            promises.push(this.constantTable.put(doc, constant));
        }

        for (const func of doc.functions) {
            promises.push(this.functionTable.put(doc, func));
        }

        for (const method of doc.methods) {
            promises.push(this.methodTable.put(doc, method));
        }

        for (const property of doc.properties) {
            promises.push(this.propertyTable.put(doc, property));
        }

        for (const globalVariable of doc.globalVariables) {
            globalVariable.assignExtraTypeForVariables();
            promises.push(this.globalVariableTable.put(doc, globalVariable));
        }

        await Promise.all(promises);
        await this.phpDocTable.put(doc, this.openedUri.has(doc.uri));
    }
}