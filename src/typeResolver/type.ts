export class Type {
    private _type: string = '';

    constructor(type: string) {
        this._type = type;
    }

    get isEmpty(): boolean {
        return this._type !== '';
    }

    public toString(): string {
        return this._type;
    }
}

export class TypeComposite {
    private _types: Type[] = [];

    public push(type: Type | TypeComposite) {
        if (type instanceof Type) {
            this._types.push(type);
        } else {
            type.types.forEach((type) => {
                this._types.push(type);
            });
        }
    }

    get types(): Type[] {
        return this._types;
    }
}

export class KeywordType extends Type { }