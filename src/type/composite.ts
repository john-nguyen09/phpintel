import { TypeName } from "./name";
import { nonenumerable } from "../util/decorator";

export class TypeComposite {
    @nonenumerable
    protected existingTypes: { [key: string]: boolean } = {};
    protected _types: TypeName[] = [];

    push(type: TypeName | null) {
        if (type == null || type.name in this.existingTypes) {
            return;
        }

        this._types.push(type);
        this.existingTypes[type.name] = true;
    }

    clone(): TypeComposite {
        let result = new TypeComposite();

        for (let type of this.types) {
            result.push(type);
        }

        return result;
    }

    get types(): TypeName[] {
        return this._types;
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