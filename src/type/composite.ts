import { TypeName } from "./name";

export class TypeComposite {
    protected existingTypes: { [key: string]: boolean } = {};
    protected _types: TypeName[] = [];

    push(type: TypeName | TypeComposite | null) {
        if (type == null) {
            return;
        }

        const types: TypeName[] = [];

        if (type instanceof TypeName) {
            types.push(type);
        } else {
            types.push(...type.types.filter(type => !type.isEmpty()));
        }

        for (const type of types) {
            if (type.name in this.existingTypes) {
                return;
            }
    
            this._types.push(type);
            this.existingTypes[type.name] = true;
        }
    }

    clone(): TypeComposite {
        let result = new TypeComposite();

        for (let type of this.types) {
            result.push(type);
        }

        return result;
    }

    toString(): string {
        return this.types.map((type) => {
            return type.name;
        }).join('|');
    }

    get types(): TypeName[] {
        return this._types;
    }

    get isEmpty(): boolean {
        return this.types.length === 0;
    }
}

export namespace ResolveType {
    export function forType(types: TypeComposite | TypeName, callback: (type: TypeName) => void) {
        if (types instanceof TypeComposite) {
            for (const type of types.types) {
                callback(type);
            }
        } else {
            callback(types);
        }
    }
}