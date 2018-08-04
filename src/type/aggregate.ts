import { Name } from "./name";

export class TypeAggregate {
    protected _types: Name[] = [];

    push(type: Name) {
        this._types.push(type);
    }

    get types(): Name[] {
        return this._types;
    }
}