export class Type {
    private _type: string = '';

    constructor(type: string) {
        this._type = type;
    }

    public toString(): string {
        return this._type;
    }
}

export class TypeComposite {
    private _types: Type[] = [];

    public pushType(type: Type) {
        this._types.push(type);
    }
}

export class KeywordType extends Type { }