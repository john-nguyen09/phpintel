import { injectable } from "inversify";
import { DbStore, LevelDatasource, SubStore } from "../db";
import { CompletionIndex, CompletionValue } from "./index/completionIndex";
import { Interface } from "../../symbol/interface/interface";
import { Serializer, Deserializer } from "../serializer";
import { TypeName } from "../../type/name";
import { PhpDocument } from "../../symbol/phpDocument";
import { BelongsToDoc } from "./index/belongsToDoc";
import { NameIndex } from "./index/nameIndex";
import { App } from "../../app";
import { Indexer } from "../../index/indexer";

@injectable()
export class InterfaceTable {
    private db: DbStore;
    private nameIndex: DbStore;
    private completionIndex: CompletionIndex;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'interface',
            version: 1,
            valueEncoding: InterfaceEncoding,
        });
        this.nameIndex = new SubStore(level, {
            name: 'interfaceNameIndex',
            version: 1
        });
        this.completionIndex = new CompletionIndex(level, 'interfaceCompletionIndex');
    }

    async put(phpDoc: PhpDocument, symbol: Interface) {
        const name = symbol.getName();

        return Promise.all([
            BelongsToDoc.put(this.db, phpDoc, name, symbol),
            NameIndex.put(this.nameIndex, phpDoc, name),
            this.completionIndex.put(phpDoc, name),
        ]);
    }

    async get(name: string): Promise<Interface[]> {
        const uris: string[] = await NameIndex.get(this.nameIndex, name);

        return BelongsToDoc.getMultiple(this.db, uris, name);
    }

    async search (keyword: string): Promise<CompletionValue[]> {
        return await this.completionIndex.search(keyword);
    }

    async getByDoc(phpDoc: PhpDocument): Promise<Interface[]> {
        const indexer: Indexer = App.get<Indexer>(Indexer);

        if (indexer.isOpen(phpDoc.uri)) {
            return phpDoc.interfaces;
        }

        return await BelongsToDoc.getByDoc<Interface>(this.db, phpDoc.uri);
    }

    async removeByDoc(uri: string): Promise<void> {
        const names = await BelongsToDoc.removeByDoc(this.db, uri);
        const promises: Promise<void>[] = [];

        for (const name of names) {
            promises.push(NameIndex.remove(this.nameIndex, uri, name));
            promises.push(this.completionIndex.del(uri, name));
        }

        await Promise.all(promises);
    }
}

const InterfaceEncoding: Level.Encoding = {
    type: 'interface',
    encode: (symbol: Interface): string => {
        const serializer = new Serializer();

        serializer.setTypeName(symbol.name);
        serializer.setString(symbol.description);
        serializer.setLocation(symbol.location);

        serializer.setInt32(symbol.parents.length);
        for (const parentName of symbol.parents) {
            serializer.setTypeName(parentName);
        }

        return serializer.getBuffer();
    },
    decode: (buffer: string): Interface => {
        const symbol = new Interface();
        const serializer = new Deserializer(buffer);

        symbol.name = serializer.readTypeName() || new TypeName('');
        symbol.description = serializer.readString();
        symbol.location = serializer.readLocation();
        
        const noParents = serializer.readInt32();
        for (let i = 0; i < noParents; i++) {
            const parentName = serializer.readTypeName();
            if (parentName === null) {
                continue;
            }

            symbol.parents.push(parentName);
        }

        return symbol;
    },
    buffer: false
};