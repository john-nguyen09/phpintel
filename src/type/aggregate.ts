export class TypeAggregate {
    protected _types: string[] = [];

    push(type: string) {
        if (this._types.indexOf(type) == -1) {
            this._types.push(type);
        }
    }

    get types(): string[] {
        return this._types;
    }
}