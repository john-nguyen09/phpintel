import * as fs from "fs";
import * as path from "path";
import { promisify } from "util";
import { pathToUri } from "../util/uri";
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
import { ReferenceTable } from "../storage/table/referenceTable";

const readdirAsync = promisify(fs.readdir);
const readFileAsync = promisify(fs.readFile);
const statAsync = promisify(fs.stat);

@injectable()
export class Indexer {
    constructor(
        private treeTraverser: Traverser,
        private phpDocTable: PhpDocumentTable,
        private classTable: ClassTable,
        private classConstantTable: ClassConstantTable,
        private constantTable: ConstantTable,
        private functionTable: FunctionTable,
        private methodTable: MethodTable,
        private propertyTable: PropertyTable,
        private referenceTable: ReferenceTable
    ) { }

    async indexFile(filePath: string, fstat?: fs.Stats): Promise<void> {
        if (typeof fstat === 'undefined') {
            fstat = await statAsync(filePath);
        }

        let fileUri = pathToUri(filePath);
        let lastIndexedPhpDoc = await this.phpDocTable.get(fileUri);
        let phpDoc: PhpDocument;
        let fileContent = (await readFileAsync(filePath)).toString('utf-8');

        if (lastIndexedPhpDoc !== null) {
            phpDoc = lastIndexedPhpDoc;
        } else {
            phpDoc = new PhpDocument(fileUri, fileContent);
        }

        const fileModifiedTime = Math.round(fstat.mtime.getTime() / 1000);

        if (fileModifiedTime !== phpDoc.modifiedTime) {
            phpDoc.text = fileContent;
            phpDoc.modifiedTime = fileModifiedTime;
            let symbolParser = new SymbolParser(phpDoc);

            this.treeTraverser.traverse(phpDoc.getTree(), [symbolParser]);
            await this.indexPhpDocument(symbolParser.getPhpDoc());
        }
    }

    async indexDir(directory: string): Promise<void> {
        let files = await readdirAsync(directory);

        for (let file of files) {
            let filePath = path.join(directory, file);
            let fstat = await statAsync(filePath);

            if (fstat.isDirectory()) {
                await this.indexDir(filePath);
            } else if (file.endsWith('.php')) {
                await this.indexFile(filePath);
            }
        }
    }

    private async removeSymbolsByDoc(uri: string) {
        return Promise.all([
            this.classTable.removeByDoc(uri),
            this.classConstantTable.removeByDoc(uri),
            this.constantTable.removeByDoc(uri),
            this.functionTable.removeByDoc(uri),
            this.methodTable.removeByDoc(uri),
            this.propertyTable.removeByDoc(uri),
            this.referenceTable.removeByDoc(uri),
            this.phpDocTable.remove(uri)
        ]);
    }

    private async indexPhpDocument(doc: PhpDocument): Promise<void> {
        await this.removeSymbolsByDoc(doc.uri);
        
        for (let theClass of doc.classes) {
            await this.classTable.put(doc, theClass);
        }

        for (let classConstant of doc.classConstants) {
            await this.classConstantTable.put(doc, classConstant);
        }

        for (let constant of doc.constants) {
            await this.constantTable.put(doc, constant);
        }

        for (let func of doc.functions) {
            await this.functionTable.put(doc, func);
        }

        for (let method of doc.methods) {
            await this.methodTable.put(doc, method);
        }

        for (let property of doc.properties) {
            await this.propertyTable.put(doc, property);
        }

        for (let reference of doc.references) {
            await this.referenceTable.put(reference);
        }
        
        await this.phpDocTable.put(doc);
    }
}