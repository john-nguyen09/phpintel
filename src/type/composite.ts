import { TypeName } from "./name";
import { Reference } from "../symbol/reference";
import { PhpDocument } from "../symbol/phpDocument";
import { RefResolver } from "../handler/refResolver";
import { isNamedSymbol } from "../symbol/symbol";
import { Method } from "../symbol/function/method";
import { Function } from "../symbol/function/function";

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

export class ExpressedType extends TypeComposite {
    private ref: Reference;

    public setReference(ref: Reference) {
        this.ref = ref;
    }

    public clone(): TypeComposite {
        const clone = new ExpressedType();
        clone.setReference(this.ref);

        for (const type of this.types) {
            clone.push(type);
        }

        return clone;
    }

    public async resolve(phpDoc: PhpDocument) {
        const symbols = await RefResolver.getSymbolsByReference(phpDoc, this.ref);

        for (const symbol of symbols) {
            if (
                symbol instanceof Method ||
                symbol instanceof Function
            ) {
                for (const type of symbol.types) {
                    this.push(type);
                }
            }
        }
    }
}