import { injectable } from "inversify";
import { DbStore, LevelDatasource, SubStore } from "../db";
import { GlobalVariable } from "../../symbol/variable/globalVariable";
import * as bytewise from "bytewise";
import { Variable } from "../../symbol/variable/variable";
import { Serializer, Deserializer } from "../serializer";
import { PhpDocument } from "../../symbol/phpDocument";
import { BelongsToDoc } from "./index/belongsToDoc";
import { NameIndex } from "./index/nameIndex";
import { inspect } from "util";

@injectable()
export class GlobalVariableTable {
    private db: DbStore;
    private nameIndex: DbStore;

    constructor(level: LevelDatasource) {
        this.db = new SubStore(level, {
            name: 'globalVariable',
            version: 1,
            keyEncoding: bytewise,
            valueEncoding: GlobalVariableCodec
        });
        this.nameIndex = new SubStore(level, {
            name: 'globalVariableNameIndex',
            version: 1,
        });
    }

    async put(phpDoc: PhpDocument, symbol: GlobalVariable) {
        if (!symbol.isDefinition) {
            return;
        }

        const promises: Promise<void[]>[] = [];

        for (const variable of symbol.variables) {
            const name = variable.name;

            promises.push(Promise.all([
                BelongsToDoc.put(this.db, phpDoc, name, variable),
                NameIndex.put(this.nameIndex, phpDoc, name),
            ]));
        }

        return Promise.all(promises);
    }

    async get(name: string): Promise<Variable[]> {
        const uris = await NameIndex.get(this.nameIndex, name);
        const variables: Promise<Variable>[] = [];

        for (const uri of uris) {
            variables.push(BelongsToDoc.get<Variable>(this.db, uri, name));
        }

        return Promise.all(variables);
    }

    async removeByDoc(uri: string) {
        const names = await BelongsToDoc.removeByDoc(this.db, uri);
        const promises: Promise<void>[] = [];

        for (const name of names) {
            promises.push(NameIndex.remove(this.nameIndex, uri, name));
        }

        return Promise.all(promises);
    }
}

const GlobalVariableCodec: Level.Encoding = {
    type: 'global_variable',
    encode: (symbol: Variable): string => {
        const serializer = new Serializer();

        serializer.setString(symbol.name);
        serializer.setTypeComposite(symbol.type);

        return serializer.getBuffer();
    },
    decode: (buffer: string): Variable => {
        const deserializer = new Deserializer(buffer);

        return new Variable(deserializer.readString(), deserializer.readTypeComposite());
    },
    buffer: false
};