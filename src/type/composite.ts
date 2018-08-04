import { TypeName } from "./name";

export class TypeComposite {
    protected _types: TypeName[] = [];

    push(type: TypeName) {
        this._types.push(type);
    }

    get types(): TypeName[] {
        return this._types;
    }
}